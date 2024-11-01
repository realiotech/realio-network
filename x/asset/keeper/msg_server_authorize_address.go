package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (ms msgServer) AuthorizeAddress(goCtx context.Context, msg *types.MsgAuthorizeAddress) (*types.MsgAuthorizeAddressResponse, error) {
	// Check if the value exists
	token, err := ms.Token.Get(goCtx, types.TokenKey(msg.Symbol))
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "symbol %s does not exists : %s", msg.Symbol, err.Error())
	}

	// assert that the manager account is the only signer of the message
	if msg.Manager != token.Manager {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "caller not authorized")
	}

	accAddress, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid address")
	}

	token.AuthorizeAddress(accAddress)
	err = ms.Token.Set(goCtx, types.TokenKey(msg.Symbol), token)
	if err != nil {
		return nil, types.ErrSetTokenUnable
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenAuthorized,
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
		),
	)

	return &types.MsgAuthorizeAddressResponse{}, nil
}
