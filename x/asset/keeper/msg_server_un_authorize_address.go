package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (k msgServer) UnAuthorizeAddress(goCtx context.Context, msg *types.MsgUnAuthorizeAddress) (*types.MsgUnAuthorizeAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	token, isFound := k.GetToken(ctx, msg.Symbol)
	if !isFound {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "symbol %s does not exists", msg.Symbol)
	}

	// Checks if the token manager signed
	signers := msg.GetSigners()
	if len(signers) != 1 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "invalid signers")
	}

	// assert that the manager account is the only signer of the message
	if signers[0].String() != token.Manager {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "caller not authorized")
	}

	delete(token.Authorized, msg.Address)

	k.SetToken(ctx, token)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenUnAuthorized,
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
		),
	)

	return &types.MsgUnAuthorizeAddressResponse{}, nil
}
