package keeper

import (
	"context"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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

func (k Keeper) QueryPrivilege(c context.Context, req *types.QueryPrivilegeRequest) (*types.QueryPrivilegeResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if strings.Trim(req.PrivilegeName, " ") == "" {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "privilege name must not be empty")
	}

	priv, ok := k.PrivilegeManager[req.PrivilegeName]
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, fmt.Sprintf("privilege with name %s not registered yet", req.PrivilegeName))
	}

	protoMsg, err := UnpackAnyMsg(req.Request)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid privilege message")
	}

	queryHandler := priv.QueryHandler()
	res, err := queryHandler(ctx, protoMsg)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	resAny, err := codectypes.NewAnyWithValue(res)
	if err != nil {
		return nil, err
	}

	return &types.QueryPrivilegeResponse{
		Response: resAny,
	}, nil
}
