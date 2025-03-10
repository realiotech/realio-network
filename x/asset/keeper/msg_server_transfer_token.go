package keeper

import (
	"context"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/asset/types"
)

func (ms msgServer) TransferToken(goCtx context.Context, msg *types.MsgTransferToken) (*types.MsgTransferTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var fromAddress, toAddress sdk.AccAddress
	isAuthorizedFrom, isAuthorizedTo := true, true

	lowerCaseSymbol := strings.ToLower(msg.Symbol)

	fromAddress, _ = sdk.AccAddressFromBech32(msg.From)
	toAddress, _ = sdk.AccAddressFromBech32(msg.To)
	// Check if the value already exists
	token, err := ms.Token.Get(
		ctx,
		types.TokenKey(msg.Symbol),
	)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "token %s not found: %s", lowerCaseSymbol, err.Error())
	}

	if ms.bankKeeper.BlockedAddr(toAddress) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.To)
	}

	if token.AuthorizationRequired {
		isAuthorizedFrom = ms.IsAddressAuthorizedToSend(ctx, lowerCaseSymbol, fromAddress)
		isAuthorizedTo = ms.IsAddressAuthorizedToSend(ctx, lowerCaseSymbol, toAddress)
	}

	if isAuthorizedFrom && isAuthorizedTo {
		totalInt, totalIsValid := math.NewIntFromString(msg.Amount)
		if !totalIsValid {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "invalid coin amount %s", msg.Amount)
		}

		baseDenom := fmt.Sprintf("a%s", lowerCaseSymbol)
		coin := sdk.Coins{{Denom: baseDenom, Amount: totalInt}}
		err := ms.bankKeeper.SendCoins(ctx, fromAddress, toAddress, coin)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s transfer not authorized", lowerCaseSymbol)
	}

	return &types.MsgTransferTokenResponse{}, nil
}
