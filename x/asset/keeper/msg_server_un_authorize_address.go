package keeper

import (
	"context"
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (k msgServer) UnAuthorizeAddress(goCtx context.Context, msg *types.MsgUnAuthorizeAddress) (*types.MsgUnAuthorizeAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	token, isFound := k.GetToken(ctx, msg.Symbol)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("index %v not set", msg.Symbol))
	}

	// Checks if the the msg sender is the same as the current owner
	if msg.Creator != token.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "caller not authorized")
	}

	delete(token.Authorized, msg.Address)

	k.SetToken(ctx, token)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenUnAuthorized,
			sdk.NewAttribute(types.AttributeKeySymbol, fmt.Sprint(token.Symbol)),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
		),
	)

	return &types.MsgUnAuthorizeAddressResponse{}, nil
}
