package transfer_auth

import (
	"context"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	// "github.com/cosmos/cosmos-sdk/store/types"
)

type updateFn = func(ctx sdk.Context, addr, tokenId string) error

func (mp TransferAuthPriviledge) UpdateAllowList(ctx sdk.Context, msg *MsgUpdateAllowList, tokenID string) error {

	var fn updateFn
	switch msg.ActionType {
	case ActionType_ACTION_TYPE_UNSPECIFIED, ActionType_ACTION_TYPE_ADD:
		fn = mp.AddAddr
	case ActionType_ACTION_TYPE_REMOVE:
		fn = mp.RemoveAddr
	default:
		return fmt.Errorf("invalid action type %s", msg.ActionType.String())
	}

	for _, addr := range msg.Addresses {
		err := fn(ctx, addr, msg.TokenId)
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
