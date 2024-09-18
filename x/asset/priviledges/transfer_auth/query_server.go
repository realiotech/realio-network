package transfer_auth

import (
	"context"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

func (tp TransferAuthPriviledge) QueryWhitelistedAddresses(ctx sdk.Context, req *QueryWhitelistedAddressesRequest, tokenID string) (*QueryWhitelistedAddressesResponse, error) {
	return &QueryWhitelistedAddressesResponse{
		WhitelistedAddrs: tp.GetWhitelistedAddrs(ctx, req.TokenId),
	}, nil
}

func (mp TransferAuthPriviledge) QueryIsAddressWhitelisted(ctx sdk.Context, req *QueryIsAddressWhitelistedRequest, tokenID string) (*QueryIsAddressWhitelistedRespones, error) {
	isWhitelisted := mp.CheckAddressIsWhitelisted(ctx, tokenID, req.Address)

	return &QueryIsAddressWhitelistedRespones{
		IsWhitelisted: isWhitelisted,
	}, nil
}

func (mp TransferAuthPriviledge) QueryHandler(context context.Context, req proto.Message, tokenID string) (proto.Message, error) {
	ctx := sdk.UnwrapSDKContext(context)

	switch req := req.(type) {
	case *QueryWhitelistedAddressesRequest:
		return mp.QueryWhitelistedAddresses(ctx, req, tokenID)
	case *QueryIsAddressWhitelistedRequest:
		return mp.QueryIsAddressWhitelisted(ctx, req, tokenID)
	default:
		errMsg := fmt.Sprintf("unrecognized query request type: %T for Transfer auth priviledge", req)
		return nil, errors.Errorf(errMsg)
	}
}
