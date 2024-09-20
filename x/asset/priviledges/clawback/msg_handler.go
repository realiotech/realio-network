package clawback

import (
	"context"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
	// "github.com/cosmos/cosmos-sdk/store/types"
)

func (cp ClawbackPriviledge) clawbackToken(ctx sdk.Context, msg *MsgClawbackToken, tokenID string, privAcc sdk.AccAddress) error {

	clawbackCoin := sdk.NewCoin(tokenID, sdk.NewIntFromUint64(msg.Amount))

	senderAddr, err := sdk.AccAddressFromBech32(msg.Account)
	if err != nil {
		return fmt.Errorf("invalid bech 32 address: %v", err)
	}

	spendable := cp.bk.SpendableCoins(ctx, senderAddr)

	if spendable.IsAllLT(sdk.NewCoins(sdk.NewCoin(tokenID, sdk.NewIntFromUint64(msg.Amount)))) {
		return fmt.Errorf("insufficient funds want %s have %s", clawbackCoin.String(), spendable.String())
	}

	err = cp.bk.SendCoins(ctx, senderAddr, privAcc, sdk.NewCoins(clawbackCoin))
	if err != nil {
		return err
	}

	return err
}

func (cp ClawbackPriviledge) MsgHandler() assettypes.MsgHandler {
	return func(context context.Context, msg proto.Message, tokenID string, privAcc sdk.AccAddress) (proto.Message, error) {
		ctx := sdk.UnwrapSDKContext(context)

		switch msg := msg.(type) {
		case *MsgClawbackToken:
			return nil, cp.clawbackToken(ctx, msg, tokenID, privAcc)
		default:
			errMsg := fmt.Sprintf("unrecognized message type: %T for Clawback priviledge", msg)
			return nil, errors.Errorf(errMsg)
		}
	}
}
