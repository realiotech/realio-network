package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/v2/x/asset/types"
)

func (k msgServer) AuthorizeAddress(goCtx context.Context, msg *types.MsgAuthorizeAddress) (*types.MsgAuthorizeAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	token, isFound := k.GetToken(ctx, msg.Symbol)
	if !isFound {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "symbol %s does not exists", msg.Symbol)
	}

	// Checks if the token manager signed
	signers := msg.GetSigners()
	if len(signers) != 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid signers")
	}

	// assert that the manager account is the only signer of the message
	if signers[0].String() != token.Manager {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "caller not authorized")
	}

	accAddress, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid address")
	}

	token.AuthorizeAddress(accAddress)
	k.SetToken(ctx, token)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenAuthorized,
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
		),
	)

	return &types.MsgAuthorizeAddressResponse{}, nil
}
