package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/rststaking/types"
)

// SetRstStake set a specific rstStake in the store from its id
func (k Keeper) SetRstStake(ctx sdk.Context, rstStake types.RstStake) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RstStakeKeyPrefix))
	b := k.cdc.MustMarshal(&rstStake)
	store.Set(types.RstStakeKey(
		rstStake.Id,
	), b)
}

// GetRstStake returns a rstStake from its id
func (k Keeper) GetRstStake(
	ctx sdk.Context,
	id string,

) (val types.RstStake, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RstStakeKeyPrefix))

	b := store.Get(types.RstStakeKey(
		id,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveRstStake removes a rstStake from the store
func (k Keeper) RemoveRstStake(
	ctx sdk.Context,
	id string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RstStakeKeyPrefix))
	store.Delete(types.RstStakeKey(
		id,
	))
}

// GetAllRstStake returns all rstStake
func (k Keeper) GetAllRstStake(ctx sdk.Context) (list []types.RstStake) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RstStakeKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.RstStake
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
