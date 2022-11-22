package keeper

import (
	"context"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/asset/types"
)

func (k msgServer) CreateToken(goCtx context.Context, msg *types.MsgCreateToken) (*types.MsgCreateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value already exists
	_, isFound := k.GetToken(
		ctx,
		msg.Symbol,
	)
	if isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "symbol already set")
	}

	var creatorAccAddress, err = sdk.AccAddressFromBech32(msg.Creator)

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid creator address")
	}

	var token = types.Token{
		Creator:               msg.Creator,
		Name:                  msg.Name,
		Symbol:                msg.Symbol,
		Total:                 msg.Total,
		Decimals:              msg.Decimals,
		AuthorizationRequired: msg.AuthorizationRequired,
	}

	// mint coins for the current module
	amount, _ := sdk.NewIntFromString(msg.Total)
	var coin = sdk.Coins{{Denom: msg.Symbol, Amount: amount}}

	k.SetToken(
		ctx,
		token,
	)

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, coin)
	if err != nil {
		panic(err)
	}

	k.bankKeeper.SetDenomMetaData(ctx, bank.Metadata{Base: msg.Symbol, Display: msg.Symbol, DenomUnits: []*bank.DenomUnit{{Denom: msg.Symbol, Exponent: 0}}})

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAccAddress, coin)
	if err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenCreated,
			sdk.NewAttribute(sdk.AttributeKeyAmount, fmt.Sprint(msg.Total)),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
		),
	)

	return &types.MsgCreateTokenResponse{}, nil
}
