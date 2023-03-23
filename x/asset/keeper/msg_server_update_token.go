package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (k msgServer) UpdateToken(goCtx context.Context, msg *types.MsgUpdateToken) (*types.MsgUpdateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, isFound := k.GetToken(ctx, msg.Symbol)
	if !isFound {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "symbol %s does not exists", msg.Symbol)
	}

	// Checks if the token manager signed
	signers := msg.GetSigners()
	if len(signers) != 1 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "invalid signers")
	}

	// assert that the manager account is the only signer of the message
	if signers[0].String() != existing.Manager {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "caller not authorized")
	}

	// only Authorization Flag is updatable at this time
	token := types.Token{
		Name:                  existing.Name,
		Symbol:                existing.Symbol,
		Total:                 existing.Total,
		Manager:               existing.Manager,
		AuthorizationRequired: msg.AuthorizationRequired,
	}

	k.SetToken(ctx, token)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenUpdated,
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
		),
	)

	return &types.MsgUpdateTokenResponse{}, nil
}
