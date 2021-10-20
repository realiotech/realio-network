package keeper

import (
	"context"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/network/x/asset/types"
)

func (k msgServer) TransferToken(goCtx context.Context, msg *types.MsgTransferToken) (*types.MsgTransferTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var err error
	var fromAddress, toAddress sdk.AccAddress
	fromAddress, err = sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}

	toAddress, err = sdk.AccAddressFromBech32(msg.To)
	if err != nil {
		panic(err)
	}

	isAuthorizedFrom := k.IsAddressAuthorizedToSend(ctx, msg.Symbol, fromAddress)
	isAuthorizedTo := k.IsAddressAuthorizedToSend(ctx, msg.Symbol, toAddress)

	if isAuthorizedFrom && isAuthorizedTo {
		var coin = sdk.Coins{{Denom: msg.Symbol, Amount: sdk.NewInt(msg.Amount)}}
		err := k.bankKeeper.SendCoins(ctx, fromAddress, toAddress, coin)
		if err != nil {
			panic(err)
		}
	} else {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s transfer not authorized", msg.Symbol)
	}

	return &types.MsgTransferTokenResponse{}, nil
}
