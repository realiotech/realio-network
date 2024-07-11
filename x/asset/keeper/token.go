package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/types"
)

// SetToken store the token with specific token id
func (k Keeper) SetToken(ctx sdk.Context, tokenId string, token types.Token) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenKey)
	key := []byte(tokenId)
	bz := k.cdc.MustMarshal(&token)

	store.Set(key, bz)
}

// GetToken get the token with the specific token id
func (k Keeper) GetToken(ctx sdk.Context, tokenId string) (types.Token, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenKey)
	key := []byte(tokenId)

	bz := store.Get(key)
	if bz == nil {
		return types.Token{}, false
	}

	var token types.Token
	k.cdc.MustUnmarshal(bz, &token)

	return token, true
}

// SetTokenManagement set the token management with the specific token id
func (k Keeper) SetTokenManagement(ctx sdk.Context, tokenId string, tm types.TokenManagement) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenManagementKey)
	key := []byte(tokenId)
	bz := k.cdc.MustMarshal(&tm)

	store.Set(key, bz)
}

// GetTokenManagement get the token management with the specific token id
func (k Keeper) GetTokenManagement(ctx sdk.Context, tokenId string) (types.TokenManagement, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenManagementKey)
	key := []byte(tokenId)

	bz := store.Get(key)
	if bz == nil {
		return types.TokenManagement{}, false
	}

	var token types.TokenManagement
	k.cdc.MustUnmarshal(bz, &token)

	return token, true
}

func (k Keeper) SetTokenPrivilegedAccount(
	ctx sdk.Context,
	tokenId string,
	privilege string,
	address sdk.AccAddress,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PrivilegedAccountsKey)
	key := append([]byte(tokenId), []byte(privilege)...)
	store.Set(key, address)
}

func (k Keeper) GetTokenPrivilegedAccount(
	ctx sdk.Context,
	tokenId string,
	privilege string,
) (sdk.AccAddress, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PrivilegedAccountsKey)
	key := append([]byte(tokenId), []byte(privilege)...)

	bz := store.Get(key)
	if bz == nil {
		return sdk.AccAddress{}, false
	}

	return sdk.AccAddress(bz), true
}
