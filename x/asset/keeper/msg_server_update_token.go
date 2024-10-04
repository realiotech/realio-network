package keeper

import (
	"context"
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (k msgServer) UpdateToken(goCtx context.Context, msg *types.MsgUpdateToken) (*types.MsgUpdateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lowerCaseSymbol := strings.ToLower(msg.Symbol)
	existing, err := k.Token.Get(ctx, lowerCaseSymbol)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "symbol %s does not exists: %s", lowerCaseSymbol, err.Error())
	}

	// Checks if the token manager signed
	signers := msg.GetSigners()
	if len(signers) != 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid signers")
	}

	// assert that the manager account is the only signer of the message
	if signers[0].String() != existing.Manager {
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

	err = k.Token.Set(goCtx, lowerCaseSymbol, token)
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
