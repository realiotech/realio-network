package keeper

import (
	"context"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/v1/x/asset/types"
)

func (k msgServer) TransferToken(goCtx context.Context, msg *types.MsgTransferToken) (*types.MsgTransferTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var fromAddress, toAddress sdk.AccAddress
	var isAuthorizedFrom, isAuthorizedTo = true, true

	fromAddress, _ = sdk.AccAddressFromBech32(msg.From)
	toAddress, _ = sdk.AccAddressFromBech32(msg.To)
	// Check if the value already exists
	token, isFound := k.GetToken(
		ctx,
		msg.Symbol,
	)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "token not found")
	}

	if token.AuthorizationRequired == true {
		isAuthorizedFrom = k.IsAddressAuthorizedToSend(ctx, msg.Symbol, fromAddress)
		isAuthorizedTo = k.IsAddressAuthorizedToSend(ctx, msg.Symbol, toAddress)
	}

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
