package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/realiotech/realio-network/x/rststaking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) RstStakeAll(c context.Context, req *types.QueryAllRstStakeRequest) (*types.QueryAllRstStakeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var rstStakes []types.RstStake
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	rstStakeStore := prefix.NewStore(store, types.KeyPrefix(types.RstStakeKeyPrefix))

	pageRes, err := query.Paginate(rstStakeStore, req.Pagination, func(key []byte, value []byte) error {
		var rstStake types.RstStake
		if err := k.cdc.Unmarshal(value, &rstStake); err != nil {
			return err
		}

		rstStakes = append(rstStakes, rstStake)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllRstStakeResponse{RstStake: rstStakes, Pagination: pageRes}, nil
}

func (k Keeper) RstStake(c context.Context, req *types.QueryGetRstStakeRequest) (*types.QueryGetRstStakeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetRstStake(
		ctx,
		req.Id,
	)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "not found")
	}

	return &types.QueryGetRstStakeResponse{RstStake: val}, nil
}
