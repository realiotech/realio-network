package keeper

import (
	"context"

	corestore "cosmossdk.io/core/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/blacklist/types"
)

// blacklistedFlag is the stored value for a blacklisted address. Its content
// doesn't matter (the module is a set, not a map to any richer value); a
// single non-empty byte keeps the store entry cheap and its presence
// unambiguous.
var blacklistedFlag = []byte{1}

// Keeper stores the set of blacklisted account addresses, keyed by their raw
// bytes (sdk.AccAddress). It intentionally has no Msg service: entries are
// only ever added by genesis or by a chain upgrade handler calling
// SetBlacklisted directly, never by a user-submitted transaction.
type Keeper struct {
	storeService corestore.KVStoreService
}

func NewKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{storeService: storeService}
}

// SetBlacklisted adds an address to the blacklist.
func (k Keeper) SetBlacklisted(ctx context.Context, addr sdk.AccAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.AddressKey(addr), blacklistedFlag)
}

// RemoveBlacklisted removes an address from the blacklist.
func (k Keeper) RemoveBlacklisted(ctx context.Context, addr sdk.AccAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.AddressKey(addr))
}

// IsBlacklisted reports whether an address is currently blacklisted. It
// never returns an error: any store-access failure is treated as "not
// blacklisted" rather than blocking transaction processing on a storage
// hiccup — the blacklist is a hardening measure, not the primary auth path.
func (k Keeper) IsBlacklisted(ctx context.Context, addr sdk.AccAddress) bool {
	store := k.storeService.OpenKVStore(ctx)
	ok, err := store.Has(types.AddressKey(addr))
	return err == nil && ok
}

// GetAllBlacklisted returns the bech32 string of every currently blacklisted
// address, in ascending key (i.e. raw address bytes) order. Used by
// ExportGenesis and for inspection/debugging.
func (k Keeper) GetAllBlacklisted(ctx context.Context) ([]string, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.AddressKeyPrefix, prefixEnd(types.AddressKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var addrs []string
	for ; iterator.Valid(); iterator.Next() {
		raw := iterator.Key()[len(types.AddressKeyPrefix):]
		addrs = append(addrs, sdk.AccAddress(raw).String())
	}
	return addrs, nil
}

// prefixEnd returns the smallest key greater than every key starting with
// prefix, i.e. the exclusive upper bound to pass as an iterator's end.
func prefixEnd(prefix []byte) []byte {
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	return nil // prefix was all 0xff bytes; unbounded end
}
