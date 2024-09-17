package transfer_auth

import (
	"context"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	// "github.com/cosmos/cosmos-sdk/store/types"
)

func (mp TransferAuthPriviledge) UpdateAllowList(ctx sdk.Context, msg *MsgUpdateAllowList, tokenID string) error {

	for _, addr := range msg.AllowedAddresses {
		err := mp.AddAddr(ctx, addr, tokenID)
		if err != nil {
			return err
		}
	}
	for _, addr := range msg.DisallowedAddresses {
		err := mp.RemoveAddr(ctx, addr, tokenID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mp TransferAuthPriviledge) MsgHandler(context context.Context, msg proto.Message, tokenID string, privAcc sdk.AccAddress) (proto.Message, error) {
	ctx := sdk.UnwrapSDKContext(context)

	switch msg := msg.(type) {
	case *MsgUpdateAllowList:
		return nil, mp.UpdateAllowList(ctx, msg, tokenID)
	default:
		errMsg := fmt.Sprintf("unrecognized message type: %T for Transfer auth priviledge", msg)
		return nil, errors.Errorf(errMsg)
	}
}
