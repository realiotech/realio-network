package freeze

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	FreezePrefix = []byte{0x2}
	TokenPrefix  = []byte{0x1}
)

func tokenPrefix(tokenID string) []byte {
	return append(TokenPrefix, tokenID...)
}

func (tp FreezePriviledge) FreezeStore(ctx sdk.Context, tokenID string) storetypes.KVStore {
	tokenStore := prefix.NewStore(ctx.KVStore(tp.storeKey), tokenPrefix(tokenID))
	return prefix.NewStore(tokenStore, FreezePrefix)
}

func (tp FreezePriviledge) SetFreezeAddress(ctx sdk.Context, tokenID string, address string) {
	store := tp.FreezeStore(ctx, tokenID)
	store.Set([]byte(address), []byte{0x01})
}

func (tp FreezePriviledge) RemoveFreezeAddress(ctx sdk.Context, tokenID string, address string) {
	store := tp.FreezeStore(ctx, tokenID)
	store.Delete([]byte(address))
}

func (tp FreezePriviledge) CheckAddressIsFreezed(ctx sdk.Context, tokenID string, address string) bool {
	store := tp.FreezeStore(ctx, tokenID)
	return store.Has([]byte(address))
}

func (tp FreezePriviledge) GetFreezeedAddrs(ctx sdk.Context, tokenID string) (whitelistedAddrs []string) {
	store := tp.FreezeStore(ctx, tokenID)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		whitelistedAddrs = append(whitelistedAddrs, string(iterator.Value()))
	}
	return
}
