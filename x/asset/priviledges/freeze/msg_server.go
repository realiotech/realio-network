package freeze

import (
	"context"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	// "github.com/cosmos/cosmos-sdk/store/types"
)

func (mp FreezePriviledge) FreezeToken(ctx sdk.Context, msg *MsgFreezeToken, tokenID string) error {
	for _, addr := range msg.Accounts {
		mp.SetFreezeAddress(ctx, tokenID, addr)
	}
	return nil
}

func (mp FreezePriviledge) UnfreezeToken(ctx sdk.Context, msg *MsgUnfreezeToken, tokenID string) error {
	for _, addr := range msg.Accounts {
		mp.RemoveFreezeAddress(ctx, tokenID, addr)
	}
	return nil
}

func (mp FreezePriviledge) MsgHandler(context context.Context, msg proto.Message, tokenID string, privAcc sdk.AccAddress) (proto.Message, error) {
	ctx := sdk.UnwrapSDKContext(context)

	switch msg := msg.(type) {
	case *MsgFreezeToken:
		return nil, mp.FreezeToken(ctx, msg, tokenID)
	case *MsgUnfreezeToken:
		return nil, mp.UnfreezeToken(ctx, msg, tokenID)
	default:
		errMsg := fmt.Sprintf("unrecognized message type: %T for Transfer auth priviledge", msg)
		return nil, errors.Errorf(errMsg)
	}
}
