package transfer_auth

import (
	"context"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

func (mp TransferAuthPriviledge) QueryAllowAddresses(ctx sdk.Context, req *QueryAllowAddressRequest, tokenID string) (*QueryAllowAddressRespones, error) {
	allowAddr, err := mp.GetAddrList(ctx, req.TokenId)
	if err != nil {
		return nil, err
	}

	return &QueryAllowAddressRespones{
		AllowAddrs: &allowAddr,
	}, nil
}

func (mp TransferAuthPriviledge) QueryIsAllow(ctx sdk.Context, req *QueryIsAllowedRequest, tokenID string) (*QueryIsAllowedRespones, error) {
	allowAddr, err := mp.GetAddrList(ctx, req.TokenId)
	if err != nil {
		return nil, err
	}

	var isAllow bool
	isAllow, has := allowAddr.Addrs[req.Address]
	if !has {
		isAllow = false
	}

	return &QueryIsAllowedRespones{
		IsAllow: isAllow,
	}, nil
}

func (mp TransferAuthPriviledge) QueryHandler(context context.Context, req proto.Message, tokenID string) (proto.Message, error) {
	ctx := sdk.UnwrapSDKContext(context)

	switch req := req.(type) {
	case *QueryAllowAddressRequest:
		return mp.QueryAllowAddresses(ctx, req, tokenID)
	case *QueryIsAllowedRequest:
		return mp.QueryIsAllow(ctx, req, tokenID)
	default:
		errMsg := fmt.Sprintf("unrecognized query request type: %T for Transfer auth priviledge", req)
		return nil, errors.Errorf(errMsg)
	}
}
