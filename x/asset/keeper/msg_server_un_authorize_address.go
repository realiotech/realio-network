package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (ms msgServer) UnAuthorizeAddress(goCtx context.Context, msg *types.MsgUnAuthorizeAddress) (*types.MsgUnAuthorizeAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	token, err := ms.Token.Get(ctx, types.TokenKey(msg.Symbol))
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "symbol %s does not exists: %s", msg.Symbol, err.Error())
	}

	// Checks if the token manager signed
	signers, _, err := ms.cdc.GetMsgV1Signers(msg)
	if err != nil {
		return nil, err
	}

	if len(signers) != 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid signers")
	}

	// assert that the manager account is the only signer of the message
	if msg.Manager != token.Manager {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "caller not authorized")
	}

	accAddress, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid address")
	}

	token.UnAuthorizeAddress(accAddress)
	err = ms.Token.Set(goCtx, types.TokenKey(msg.Symbol), token)
	if err != nil {
		return nil, types.ErrSetTokenUnable
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenUnAuthorized,
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
		),
	)

	return &types.MsgUnAuthorizeAddressResponse{}, nil
}
