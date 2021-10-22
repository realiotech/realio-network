package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/network/x/asset/types"
)

func (k msgServer) SendToAlgorand(goCtx context.Context, msg *types.MsgSendToAlgorand) (*types.MsgSendToAlgorandResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// execute the transaction....
	err := ExecuteSendToAlgorand(goCtx, msg, k.Keeper)
	if err != nil {
		// In case message is error mint back the tokens
		sender, _ := sdk.AccAddressFromBech32(msg.Creator)
		err = k.Keeper.MintTokens(ctx, sender, sdk.NewCoin(msg.Denom, sdk.NewInt(msg.Amount)))
		if err != nil {
			return nil, err
		}
	}

	// TODO do we want to store an Object? Ie one of SendToAlgorand
	return &types.MsgSendToAlgorandResponse{}, nil
}
