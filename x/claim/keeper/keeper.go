package keeper

import (
	"context"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	corestore "cosmossdk.io/core/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// Attacker is the address MsgClaimBalances draws compensation from --
	// the account the stolen funds were moved into. Kept in state rather
	// than taken from the message so that no admin slip can point a
	// compensation leg at a victim; like Admin, there is no message to
	// change it. Raw bytes, same byte-comparison rationale as above.
	Attacker collections.Item[sdk.AccAddress]

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
		Admin:              collections.NewItem(sb, types.AdminKey, "admin", collcodec.KeyToValueCodec(sdk.AccAddressKey)),
		AddressLinks:       collections.NewMap(sb, types.AddressLinkKeyPrefix, "address_links", sdk.AccAddressKey, collcodec.KeyToValueCodec(sdk.AccAddressKey)),
		Attacker:           collections.NewItem(sb, types.AttackerKey, "attacker", collcodec.KeyToValueCodec(sdk.AccAddressKey)),
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
// the attacker's remaining pool after a restitution payment.
func (k Keeper) SpendableCoinsOf(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return k.bankKeeper.SpendableCoins(ctx, addr)
}
