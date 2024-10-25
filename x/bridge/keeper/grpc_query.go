package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/realiotech/realio-network/x/bridge/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

func (q queryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	params, err := q.k.Params.Get(c)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

func (q queryServer) RateLimits(c context.Context, req *types.QueryRateLimitsRequest) (*types.QueryRateLimitsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ratelimits := []types.DenomAndRateLimit{}
	err := q.k.RegisteredCoins.Walk(c, nil, func(denom string, ratelimit types.RateLimit) (stop bool, err error) {
		ratelimits = append(ratelimits, types.DenomAndRateLimit{
			Denom:     denom,
			RateLimit: ratelimit,
		})
		return false, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryRateLimitsResponse{Ratelimits: ratelimits}, nil
}

func (q queryServer) RateLimit(c context.Context, req *types.QueryRateLimitRequest) (*types.QueryRateLimitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if r, err := q.k.RegisteredCoins.Get(ctx, req.Denom); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "not found")
	} else { //nolint:revive // fixing this causes t to be inaccessible, so let's leave all as is.
		return &types.QueryRateLimitResponse{Ratelimit: r}, nil
	}
}

func (q queryServer) EpochInfo(c context.Context, req *types.QueryEpochInfoRequest) (*types.QueryEpochInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if epochInfo, err := q.k.EpochInfo.Get(ctx); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "not found")
	} else { //nolint:revive // fixing this causes t to be inaccessible, so let's leave all as is.
		return &types.QueryEpochInfoResponse{EpochInfo: epochInfo}, nil
	}
}
