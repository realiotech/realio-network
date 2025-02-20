package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/bridge/types"
)

func (ms msgServer) BridgeIn(goCtx context.Context, msg *types.MsgBridgeIn) (*types.MsgBridgeInResponse, error) {
	param, err := ms.Params.Get(goCtx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "failed to get bridge params")
	}

	coinRegistered, err := ms.Keeper.GetCoinsRegistered(goCtx, msg.Coin.Denom)
	if err != nil {
		return nil, err
	}
	// Check if token authority
	if msg.Authority != coinRegistered.Authority {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "invalid authority; expected %s, got %s", param.Authority, msg.Authority)
	}

	coin := msg.Coin
	if found, err := ms.RegisteredCoins.Has(goCtx, coin.Denom); err != nil || !found {
		return nil, errorsmod.Wrapf(types.ErrCoinNotRegister, "denom: %s", coin.Denom)
	}

	err = ms.bankKeeper.MintCoins(goCtx, types.ModuleName, sdk.Coins{coin})
	if err != nil {
		return nil, err
	}

	addrCodec := ms.authKeeper.AddressCodec()
	accAddr, err := addrCodec.StringToBytes(msg.Reciever)
	if err != nil {
		return nil, err
	}

	err = ms.bankKeeper.SendCoinsFromModuleToAccount(goCtx, types.ModuleName, accAddr, sdk.Coins{coin})
	if err != nil {
		return nil, err
	}

	err = ms.UpdateInflow(goCtx, coin)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBridgeIn,
			sdk.NewAttribute(types.AttributeKeyDenom, coin.Denom),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Authority),
		),
	)
	return &types.MsgBridgeInResponse{}, nil
}

func (ms msgServer) BridgeOut(goCtx context.Context, msg *types.MsgBridgeOut) (*types.MsgBridgeOutResponse, error) {
	addrCodec := ms.authKeeper.AddressCodec()
	accAddr, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	coin := msg.Coin
	if found, err := ms.RegisteredCoins.Has(goCtx, coin.Denom); err != nil || !found {
		return nil, errorsmod.Wrapf(types.ErrCoinNotRegister, "denom: %s", coin.Denom)
	}

	err = ms.bankKeeper.SendCoinsFromAccountToModule(goCtx, accAddr, types.ModuleName, sdk.Coins{coin})
	if err != nil {
		return nil, err
	}

	err = ms.bankKeeper.BurnCoins(goCtx, types.ModuleName, sdk.Coins{coin})
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBridgeOut,
			sdk.NewAttribute(types.AttributeKeyDenom, coin.Denom),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Signer),
		),
	)
	return &types.MsgBridgeOutResponse{}, nil
}
