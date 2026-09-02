package keeper

import (
	"context"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/blacklist/types"
)

// Keeper stores the set of blacklisted account addresses, keyed by their raw
// bytes (sdk.AccAddress). Entries are added by genesis or by the gov-gated
// MsgUpdateBlacklist — never by an ordinary user-submitted transaction.
type Keeper struct {
	Schema collections.Schema
	// Blacklisted is a set, not a map to any richer value: presence of a
	// key is all that matters. Keyed by the raw AccAddress bytes (not its
	// bech32 string) to keep keys fixed-length and canonical — a
	// common.Address.Bytes() for an EVM sender is the same 20 raw bytes on
	// this chain.
	Blacklisted collections.KeySet[sdk.AccAddress]

	// authority is the address permitted to call MsgUpdateBlacklist
	// (the x/gov module account, unless overridden).
	authority string
}

func NewKeeper(storeService corestore.KVStoreService, authority string) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		Blacklisted: collections.NewKeySet(sb, types.AddressKeyPrefix, "blacklisted", sdk.AccAddressKey),
		authority:   authority,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAuthority returns the x/blacklist module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetBlacklisted adds an address to the blacklist.
func (k Keeper) SetBlacklisted(ctx context.Context, addr sdk.AccAddress) error {
	return k.Blacklisted.Set(ctx, addr)
}

// RemoveBlacklisted removes an address from the blacklist.
func (k Keeper) RemoveBlacklisted(ctx context.Context, addr sdk.AccAddress) error {
	return k.Blacklisted.Remove(ctx, addr)
}

// IsBlacklisted reports whether an address is currently blacklisted. It
// never returns an error: any store-access failure is treated as "not
// blacklisted" rather than blocking transaction processing on a storage
// hiccup — the blacklist is a hardening measure, not the primary auth path.
func (k Keeper) IsBlacklisted(ctx context.Context, addr sdk.AccAddress) bool {
	ok, err := k.Blacklisted.Has(ctx, addr)
	return err == nil && ok
}

// GetAllBlacklisted returns the bech32 string of every currently blacklisted
// address, in ascending key (i.e. raw address bytes) order. Used by
// ExportGenesis and for inspection/debugging.
func (k Keeper) GetAllBlacklisted(ctx context.Context) ([]string, error) {
	iter, err := k.Blacklisted.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	keys, err := iter.Keys()
	if err != nil {
		return nil, err
	}

	addrs := make([]string, len(keys))
	for i, addr := range keys {
		addrs[i] = addr.String()
	}
	return addrs, nil
}
