package keeper

import (
	"context"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (k msgServer) AuthorizeAddress(goCtx context.Context, msg *types.MsgAuthorizeAddress) (*types.MsgAuthorizeAddressResponse, error) {
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

	if token.Authorized == nil {
		// initialize map on first write
		m := make(map[string]*types.TokenAuthorization)
		token.Authorized = m
	}
	var newAuthorization = types.TokenAuthorization{Address: msg.Address, TokenSymbol: strings.ToLower(msg.Symbol), Authorized: true}

	token.Authorized[msg.Address] = &newAuthorization

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
