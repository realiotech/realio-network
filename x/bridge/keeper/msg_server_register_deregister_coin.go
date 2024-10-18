package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/bridge/types"
)

func (ms msgServer) RegisterNewCoins(goCtx context.Context, msg *types.MsgRegisterNewCoins) (*types.MsgRegisterNewCoinsResponse, error) {
	param, err := ms.Params.Get(goCtx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "failed to get bridge params")
	}

	if msg.Authority != param.Authority {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "invalid authority; expected %s, got %s", param.Authority, msg.Authority)
	}

	for _, coin := range msg.Coins {
		if found, err := ms.RegisteredCoins.Has(goCtx, coin.Denom); err != nil || found {
			return nil, errorsmod.Wrapf(types.ErrCoinAlreadyRegister, "denom: %s", coin.Denom)
		}

		err = ms.RegisteredCoins.Set(goCtx, coin.Denom, types.RateLimit{
			Ratelimit:     coin.Amount,
			CurrentInflow: math.ZeroInt(),
		})
		if err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRegisterNewCoin,
			sdk.NewAttribute(types.AttributeKeyCoins, msg.Coins.String()),
		),
	)
	return &types.MsgRegisterNewCoinsResponse{}, nil
}

func (ms msgServer) DeregisterCoins(goCtx context.Context, msg *types.MsgDeregisterCoins) (*types.MsgDeregisterCoinsResponse, error) {
	param, err := ms.Params.Get(goCtx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "failed to get bridge params")
	}

	if msg.Authority != param.Authority {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "invalid authority; expected %s, got %s", param.Authority, msg.Authority)
	}

	for _, denom := range msg.Denoms {
		if found, err := ms.RegisteredCoins.Has(goCtx, denom); err != nil || !found {
			return nil, errorsmod.Wrapf(types.ErrCoinNotRegister, "denom: %s", denom)
		}

		err = ms.RegisteredCoins.Remove(goCtx, denom)
		if err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeregisterCoin,
		),
	)
	return &types.MsgDeregisterCoinsResponse{}, nil
}
