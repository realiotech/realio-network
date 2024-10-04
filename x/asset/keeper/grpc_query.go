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

func (q queryServer) Tokens(c context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	tokens := []types.Token{}
	err := q.k.Token.Walk(c, nil, func(symbol string, token types.Token) (stop bool, err error) {
		tokens = append(tokens, token)
		return false, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTokensResponse{Tokens: tokens}, nil
}

func (q queryServer) Token(c context.Context, req *types.QueryTokenRequest) (*types.QueryTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if t, err := q.k.Token.Get(ctx, req.Symbol); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "not found")
	} else { //nolint:revive // fixing this causes t to be inaccessible, so let's leave all as is.
		return &types.QueryTokenResponse{Token: t}, nil
	}
}

func (q queryServer) IsAuthorized(c context.Context, req *types.QueryIsAuthorizedRequest) (*types.QueryIsAuthorizedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if t, err := q.k.Token.Get(ctx, req.Symbol); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "not found")
	} else { //nolint:revive // fixing this causes t to be inaccessible, so let's leave all as is.
		accAddress, _ := sdk.AccAddressFromBech32(req.Address)
		isAuthorized := t.AddressIsAuthorized(accAddress)
		return &types.QueryIsAuthorizedResponse{IsAuthorized: isAuthorized}, nil
	}
}
