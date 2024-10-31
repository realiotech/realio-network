package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (ms msgServer) UpdateToken(goCtx context.Context, msg *types.MsgUpdateToken) (*types.MsgUpdateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, err := ms.Token.Get(ctx, types.TokenKey(msg.Symbol))
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "symbol %s does not exists: %s", msg.Symbol, err.Error())
	}

	// assert that the manager account is the only signer of the message
	if msg.Manager != existing.Manager {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "caller not authorized")
	}

	// only Authorization Flag is updatable at this time
	token := types.Token{
		Name:                  existing.Name,
		Symbol:                existing.Symbol,
		Total:                 existing.Total,
		Manager:               existing.Manager,
		AuthorizationRequired: msg.AuthorizationRequired,
	}

	err = ms.Token.Set(goCtx, types.TokenKey(msg.Symbol), token)
	if err != nil {
		return nil, types.ErrSetTokenUnable
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenUpdated,
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
		),
	)

	return &types.MsgUpdateTokenResponse{}, nil
}
