package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/realiotech/realio-network/x/claim/types"
)

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// LinkedAddress implements the Query/LinkedAddress gRPC method: anyone can
// look up the new address an old address was linked to.
func (q queryServer) LinkedAddress(ctx context.Context, req *types.QueryLinkedAddressRequest) (*types.QueryLinkedAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	oldAddr, err := sdk.AccAddressFromBech32(req.OldAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, sdkerrors.ErrInvalidAddress.Wrapf("invalid old_address %q: %s", req.OldAddress, err).Error())
	}

	newAddr, found := q.k.GetLink(ctx, oldAddr)
	if !found {
		return &types.QueryLinkedAddressResponse{Found: false}, nil
	}
	return &types.QueryLinkedAddressResponse{NewAddress: newAddr.String(), Found: true}, nil
}

// Admin implements the Query/Admin gRPC method: anyone can check which
// address is currently authorized to submit MsgLinkAddress.
func (q queryServer) Admin(ctx context.Context, req *types.QueryAdminRequest) (*types.QueryAdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	admin, ok := q.k.GetAdmin(ctx)
	if !ok {
		return &types.QueryAdminResponse{Admin: ""}, nil
	}
	return &types.QueryAdminResponse{Admin: admin.String()}, nil
}
