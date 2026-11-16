package keeper

import (
	"context"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	corestore "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"

	"github.com/realiotech/realio-network/x/claim/types"
)

// Keeper links a leaked/compromised address to the new address its holder
// registered after re-doing KYC, and migrates that holder's active staking
// delegations over to the new address in the same step. See
// x/claim/README.md for the full design and the invariants
// MigrateDelegations (migrate.go) depends on.
type Keeper struct {
	Schema collections.Schema

	// Admin is the sole address authorized to submit ANY of the module's
	// messages. Stored as raw address bytes rather than a bech32
	// string so every comparison against an incoming message's admin field
	// is a byte comparison, not a string one -- immune to prefix/casing
	// quirks a string equality check would be sensitive to. Set at genesis
	// or by a chain-upgrade handler poking state directly (the same way
	// x/blacklist's leaked-address list and x/asset's manager rotations
	// are seeded on a live chain) — there is no message to change it.
	Admin collections.Item[sdk.AccAddress]

	// AddressLinks maps an old (leaked) address's raw bytes to the raw
	// bytes of the new address it was linked to (same byte-comparison
	// rationale as Admin). Many old addresses may point to the same new
	// address (consolidating several leaked wallets into one safe one); an
	// old address can only ever be linked once.
	AddressLinks collections.Map[sdk.AccAddress, sdk.AccAddress]

	// Attacker is the address MsgMigrateBankBalance pays hacked amounts
	// from -- the account the stolen funds were moved into. It lives in
	// state, not in the message, so no admin slip can point a payout at a
	// victim; like Admin, no message changes it. Raw bytes, same
	// byte-comparison rationale as above.
	Attacker collections.Item[sdk.AccAddress]

	// HackedAmounts records how much the attacker took from each victim:
	// (old address, denom) -> amount still owed. MsgMigrateBankBalance pays
	// it out of the attacker and deletes the entry, so a repeated message
	// pays nothing twice.
	//
	// One row per denom, so state needs no codec beyond the SDK's own: a
	// key pair and an Int.
	HackedAmounts collections.Map[collections.Pair[sdk.AccAddress, string], math.Int]

	stakingKeeper      *stakingkeeper.Keeper
	distrKeeper        distrkeeper.Keeper
	multiStakingKeeper multistakingkeeper.Keeper
	bankKeeper         types.BankKeeper

	// evmKeeper and erc20Keeper back MsgClaimTokens (keeper/evm.go).
	// Both are the concrete keepers rather than narrow interfaces because
	// CallEVM needs a *evmkeeper.Keeper to build a statedb.StateDB from --
	// the same reason x/erc20 holds it concretely.
	evmKeeper     *evmkeeper.Keeper
	erc20Keeper   erc20keeper.Keeper
	accountKeeper types.AccountKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	stakingKeeper *stakingkeeper.Keeper,
	distrKeeper distrkeeper.Keeper,
	multiStakingKeeper multistakingkeeper.Keeper,
	bankKeeper types.BankKeeper,
	evmKeeper *evmkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	accountKeeper types.AccountKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		Admin:        collections.NewItem(sb, types.AdminKey, "admin", collcodec.KeyToValueCodec(sdk.AccAddressKey)),
		AddressLinks: collections.NewMap(sb, types.AddressLinkKeyPrefix, "address_links", sdk.AccAddressKey, collcodec.KeyToValueCodec(sdk.AccAddressKey)),
		Attacker:     collections.NewItem(sb, types.AttackerKey, "attacker", collcodec.KeyToValueCodec(sdk.AccAddressKey)),
		HackedAmounts: collections.NewMap(sb, types.HackedAmountKeyPrefix, "hacked_amounts",
			collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), sdk.IntValue),
		stakingKeeper:      stakingKeeper,
		distrKeeper:        distrKeeper,
		multiStakingKeeper: multiStakingKeeper,
		bankKeeper:         bankKeeper,
		evmKeeper:          evmKeeper,
		erc20Keeper:        erc20Keeper,
		accountKeeper:      accountKeeper,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAdmin returns the address currently authorized to submit
// any of the module's messages, and whether one has been configured
// yet. An admin explicitly set to an empty address (as genesis does when
// no admin is configured) reports the same "not configured" result as one
// that was never set at all.
func (k Keeper) GetAdmin(ctx context.Context) (sdk.AccAddress, bool) {
	admin, err := k.Admin.Get(ctx)
	if err != nil || len(admin) == 0 {
		return nil, false
	}
	return admin, true
}

// SetAdmin sets the address authorized to submit the module's messages.
func (k Keeper) SetAdmin(ctx context.Context, admin sdk.AccAddress) error {
	return k.Admin.Set(ctx, admin)
}

// GetAttacker returns the address compensation is drawn from, and whether
// one has been configured. An attacker explicitly set to an empty address
// reports the same "not configured" result as one never set at all.
func (k Keeper) GetAttacker(ctx context.Context) (sdk.AccAddress, bool) {
	attacker, err := k.Attacker.Get(ctx)
	if err != nil || len(attacker) == 0 {
		return nil, false
	}
	return attacker, true
}

// SetAttacker sets the address compensation is drawn from.
func (k Keeper) SetAttacker(ctx context.Context, attacker sdk.AccAddress) error {
	return k.Attacker.Set(ctx, attacker)
}

// SetHackedAmount records amount as what the attacker took from addr and
// has not yet repaid, replacing anything recorded before. An empty amount
// clears the record. This is how the figure gets on chain -- genesis and
// chain-upgrade handlers call it; no message does, so the only way to
// change what a payout will be is to change state the way an upgrade does.
func (k Keeper) SetHackedAmount(ctx context.Context, addr sdk.AccAddress, amount sdk.Coins) error {
	if err := amount.Validate(); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "hacked amount for %s: %s", addr, err)
	}
	if err := k.ClearHackedAmount(ctx, addr); err != nil {
		return err
	}
	for _, coin := range amount {
		if err := k.HackedAmounts.Set(ctx, collections.Join(addr, coin.Denom), coin.Amount); err != nil {
			return err
		}
	}
	return nil
}

// GetHackedAmount returns what the attacker took from addr and has not yet
// repaid. Empty if nothing is recorded or it has already been paid.
func (k Keeper) GetHackedAmount(ctx context.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	iter, err := k.HackedAmounts.Iterate(ctx, collections.NewPrefixedPairRange[sdk.AccAddress, string](addr))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	kvs, err := iter.KeyValues()
	if err != nil {
		return nil, err
	}
	// Denoms iterate in ascending order, which is the order sdk.Coins
	// requires, so the result is already canonical.
	amount := make(sdk.Coins, 0, len(kvs))
	for _, kv := range kvs {
		amount = append(amount, sdk.Coin{Denom: kv.Key.K2(), Amount: kv.Value})
	}
	return amount, nil
}

// ClearHackedAmount removes whatever is recorded for addr. Called once the
// amount has been paid, so the same hacked amount is never paid twice.
func (k Keeper) ClearHackedAmount(ctx context.Context, addr sdk.AccAddress) error {
	return k.HackedAmounts.Clear(ctx, collections.NewPrefixedPairRange[sdk.AccAddress, string](addr))
}

// GetAllHackedAmounts returns every unpaid record, grouped by address, in
// ascending address order. Used by ExportGenesis.
func (k Keeper) GetAllHackedAmounts(ctx context.Context) ([]types.HackedAmount, error) {
	iter, err := k.HackedAmounts.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	kvs, err := iter.KeyValues()
	if err != nil {
		return nil, err
	}

	var out []types.HackedAmount
	for _, kv := range kvs {
		addr, denom := kv.Key.K1(), kv.Key.K2()
		coin := sdk.Coin{Denom: denom, Amount: kv.Value}
		if n := len(out); n > 0 && out[n-1].Address == addr.String() {
			out[n-1].Amount = append(out[n-1].Amount, coin)
			continue
		}
		out = append(out, types.HackedAmount{Address: addr.String(), Amount: sdk.Coins{coin}})
	}
	return out, nil
}

// GetLink returns the new address old was linked to, if any.
func (k Keeper) GetLink(ctx context.Context, old sdk.AccAddress) (sdk.AccAddress, bool) {
	newAddr, err := k.AddressLinks.Get(ctx, old)
	if err != nil {
		return nil, false
	}
	return newAddr, true
}

// IsLinked reports whether old has already been linked to a new address.
func (k Keeper) IsLinked(ctx context.Context, old sdk.AccAddress) bool {
	ok, err := k.AddressLinks.Has(ctx, old)
	return err == nil && ok
}

// SetLink records that old has been linked to new.
func (k Keeper) SetLink(ctx context.Context, old, new sdk.AccAddress) error {
	return k.AddressLinks.Set(ctx, old, new)
}

// GetAllLinks returns every recorded old->new address link, in ascending
// key (i.e. old address raw bytes) order. Used by ExportGenesis.
func (k Keeper) GetAllLinks(ctx context.Context) ([]types.AddressLink, error) {
	iter, err := k.AddressLinks.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	kvs, err := iter.KeyValues()
	if err != nil {
		return nil, err
	}

	links := make([]types.AddressLink, len(kvs))
	for i, kv := range kvs {
		links[i] = types.AddressLink{OldAddress: kv.Key.String(), NewAddress: kv.Value.String()}
	}
	return links, nil
}

// SpendableCoinsOf reports what addr can currently move. Used to report
// the attacker's remaining pool after a payout.
func (k Keeper) SpendableCoinsOf(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return k.bankKeeper.SpendableCoins(ctx, addr)
}
