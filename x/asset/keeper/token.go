package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/types"
)

// SetToken set a specific token in the store from its index
func (k Keeper) SetToken(ctx sdk.Context, token types.Token) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKeyPrefix))
	b := k.cdc.MustMarshal(&token)
	store.Set(types.TokenKey(
		token.Index,
	), b)
}

// GetToken returns a token from its index
func (k Keeper) GetToken(
	ctx sdk.Context,
	index string,

) (val types.Token, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKeyPrefix))

	b := store.Get(types.TokenKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveToken removes a token from the store
func (k Keeper) RemoveToken(
	ctx sdk.Context,
	index string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKeyPrefix))
	store.Delete(types.TokenKey(
		index,
	))
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

func (k Keeper) IsAddressAuthorizedToSend(ctx sdk.Context, index string, address sdk.AccAddress) (authorized bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKeyPrefix))

	b := store.Get(types.TokenKey(
		index,
	))
	if b == nil {
		return false
	}
	var t types.Token
	k.cdc.MustUnmarshal(b, &t)

	value, found := t.Authorized[address.String()]

	if found {
		return value.Authorized
	}

	return false
}
