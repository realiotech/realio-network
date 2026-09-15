package keeper

import (
	"context"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

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

	// Admin is the sole address authorized to submit MsgLinkAddress. Set at
	// genesis or by a chain-upgrade handler poking state directly (the same
	// way x/blacklist's leaked-address list and x/asset's manager
	// rotations are seeded on a live chain) — there is no message to
	// change it.
	Admin collections.Item[string]

	// AddressLinks maps an old (leaked) address's raw bytes to the bech32
	// string of the new address it was linked to. Many old addresses may
	// point to the same new address (consolidating several leaked wallets
	// into one safe one); an old address can only ever be linked once.
	AddressLinks collections.Map[sdk.AccAddress, string]

	stakingKeeper      *stakingkeeper.Keeper
	distrKeeper        distrkeeper.Keeper
	multiStakingKeeper multistakingkeeper.Keeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	stakingKeeper *stakingkeeper.Keeper,
	distrKeeper distrkeeper.Keeper,
	multiStakingKeeper multistakingkeeper.Keeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		Admin:              collections.NewItem(sb, types.AdminKey, "admin", collections.StringValue),
		AddressLinks:       collections.NewMap(sb, types.AddressLinkKeyPrefix, "address_links", sdk.AccAddressKey, collections.StringValue),
		stakingKeeper:      stakingKeeper,
		distrKeeper:        distrKeeper,
		multiStakingKeeper: multiStakingKeeper,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAdmin returns the bech32 address currently authorized to submit
// MsgLinkAddress, or "" if none has been configured yet.
func (k Keeper) GetAdmin(ctx context.Context) string {
	admin, err := k.Admin.Get(ctx)
	if err != nil {
		return ""
	}
	return admin
}

// SetAdmin sets the address authorized to submit MsgLinkAddress.
func (k Keeper) SetAdmin(ctx context.Context, admin string) error {
	return k.Admin.Set(ctx, admin)
}

// GetLink returns the new address old was linked to, if any.
func (k Keeper) GetLink(ctx context.Context, old sdk.AccAddress) (string, bool) {
	newAddr, err := k.AddressLinks.Get(ctx, old)
	if err != nil {
		return "", false
	}
	return newAddr, true
}

// IsLinked reports whether old has already been linked to a new address.
func (k Keeper) IsLinked(ctx context.Context, old sdk.AccAddress) bool {
	ok, err := k.AddressLinks.Has(ctx, old)
	return err == nil && ok
}

// SetLink records that old has been linked to new.
func (k Keeper) SetLink(ctx context.Context, old sdk.AccAddress, new string) error {
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
		links[i] = types.AddressLink{OldAddress: kv.Key.String(), NewAddress: kv.Value}
	}
	return links, nil
}
