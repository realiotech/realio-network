package keeper

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/types"
)

// SetToken set a specific token in the store from its symbol
func (k Keeper) SetToken(ctx sdk.Context, token types.Token) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKeyPrefix))
	lowerCased := strings.ToLower(token.Symbol)
	b := k.cdc.MustMarshal(&token)
	store.Set(types.TokenKey(
		lowerCased,
	), b)
}

// GetToken returns a token from its symbol
func (k Keeper) GetToken(
	ctx sdk.Context,
	symbol string,
) (val types.Token, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKeyPrefix))
	lowerCased := strings.ToLower(symbol)
	b := store.Get(types.TokenKey(
		lowerCased,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetAllToken returns all token
func (k Keeper) GetAllToken(ctx sdk.Context) (list []types.Token) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Token
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) IsAddressAuthorizedToSend(ctx sdk.Context, symbol string, address sdk.AccAddress) (authorized bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKeyPrefix))
	lowerCased := strings.ToLower(symbol)
	b := store.Get(types.TokenKey(
		lowerCased,
	))
	if b == nil {
		return false
	}
	var t types.Token
	k.cdc.MustUnmarshal(b, &t)

	return t.AddressIsAuthorized(address)
}
