package transfer_auth

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	WhitelistPrefix = []byte{0x2}
	TokenPrefix     = []byte{0x1}
)

func tokenPrefix(tokenID string) []byte {
	return append(TokenPrefix, tokenID...)
}

func (tp TransferAuthPriviledge) WhitelistStore(ctx sdk.Context, tokenID string) storetypes.KVStore {
	tokenStore := prefix.NewStore(ctx.KVStore(tp.storeKey), tokenPrefix(tokenID))
	return prefix.NewStore(tokenStore, WhitelistPrefix)
}

func (tp TransferAuthPriviledge) AddAddressToWhiteList(ctx sdk.Context, tokenID string, address string) {
	store := tp.WhitelistStore(ctx, tokenID)
	store.Set([]byte(address), []byte{0x01})
}

func (tp TransferAuthPriviledge) RemoveAddressFromWhiteList(ctx sdk.Context, tokenID string, address string) {
	store := tp.WhitelistStore(ctx, tokenID)
	store.Delete([]byte(address))
}

func (tp TransferAuthPriviledge) CheckAddressIsWhitelisted(ctx sdk.Context, tokenID string, address string) bool {
	store := tp.WhitelistStore(ctx, tokenID)
	return store.Has([]byte(address))
}

func (tp TransferAuthPriviledge) GetWhitelistedAddrs(ctx sdk.Context, tokenID string) (whitelistedAddrs []string) {
	store := tp.WhitelistStore(ctx, tokenID)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		whitelistedAddrs = append(whitelistedAddrs, string(iterator.Value()))
	}
	return
}
