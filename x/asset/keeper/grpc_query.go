package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/realiotech/realio-network/x/asset/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) Tokens(c context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryTokensResponse{Tokens: k.GetAllToken(ctx)}, nil
}

func (k Keeper) Token(c context.Context, req *types.QueryTokenRequest) (*types.QueryTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if t, found := k.GetToken(ctx, req.Symbol); !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "not found")
	} else { //nolint:revive // fixing this causes t to be inaccessible, so let's leave all as is.
		return &types.QueryTokenResponse{Token: t}, nil
	}
}

func (k Keeper) IsAuthorized(c context.Context, req *types.QueryIsAuthorizedRequest) (*types.QueryIsAuthorizedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if t, found := k.GetToken(ctx, req.Symbol); !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "not found")
	} else { //nolint:revive // fixing this causes t to be inaccessible, so let's leave all as is.
		accAddress, _ := sdk.AccAddressFromBech32(req.Address)
		isAuthorized := t.AddressIsAuthorized(accAddress)
		return &types.QueryIsAuthorizedResponse{IsAuthorized: isAuthorized}, nil
	}
}
