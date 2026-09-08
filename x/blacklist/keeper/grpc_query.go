package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/realiotech/realio-network/x/blacklist/types"
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

// IsBlacklisted implements the Query/IsBlacklisted gRPC method: it lets
// anyone check whether a given address is currently blacklisted.
func (q queryServer) IsBlacklisted(ctx context.Context, req *types.QueryIsBlacklistedRequest) (*types.QueryIsBlacklistedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	accAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, sdkerrors.ErrInvalidAddress.Wrapf("invalid address %q: %s", req.Address, err).Error())
	}

	return &types.QueryIsBlacklistedResponse{Blacklisted: q.k.IsBlacklisted(ctx, accAddr)}, nil
}
