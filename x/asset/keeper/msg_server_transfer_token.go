package keeper

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/asset/types"
)

func (k msgServer) TransferToken(goCtx context.Context, msg *types.MsgTransferToken) (*types.MsgTransferTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var fromAddress, toAddress sdk.AccAddress
	isAuthorizedFrom, isAuthorizedTo := true, true

	fromAddress, _ = sdk.AccAddressFromBech32(msg.From)
	toAddress, _ = sdk.AccAddressFromBech32(msg.To)
	// Check if the value already exists
	token, isFound := k.GetToken(
		ctx,
		msg.Symbol,
	)
	if !isFound {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "token %s not found", msg.Symbol)
	}

	if token.AuthorizationRequired {
		isAuthorizedFrom = k.IsAddressAuthorizedToSend(ctx, msg.Symbol, fromAddress)
		isAuthorizedTo = k.IsAddressAuthorizedToSend(ctx, msg.Symbol, toAddress)
	}

	if isAuthorizedFrom && isAuthorizedTo {
		totalInt, totalIsValid := math.NewIntFromString(msg.Amount)
		if !totalIsValid {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid coin amount %s", msg.Amount)
		}

		baseDenom := fmt.Sprintf("a%s", strings.ToLower(msg.Symbol))
		var coin = sdk.Coins{{Denom: baseDenom, Amount: totalInt}}
		err := k.bankKeeper.SendCoins(ctx, fromAddress, toAddress, coin)
		if err != nil {
			panic(err)
		}
	} else {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s transfer not authorized", msg.Symbol)
	}

	return &types.MsgTransferTokenResponse{}, nil
}
