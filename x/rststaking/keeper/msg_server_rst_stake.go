package keeper

import (
	"context"
	"fmt"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/realiotech/realio-network/x/rststaking/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	)

func (k msgServer) CreateRstStake(goCtx context.Context, msg *types.MsgCreateRstStake) (*types.MsgCreateRstStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value already exists
	_, isFound := k.GetRstStake(
		ctx,
		msg.Id,
	)
	if isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Id already set")
	}

	var creatorAccAddress, err = sdk.AccAddressFromBech32(msg.Creator)

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid creator address")
	}

	var holdingAccName = types.ModuleName + "/" + msg.Creator
	var holdingAcc = authtypes.NewEmptyModuleAccount(holdingAccName)
	k.accountKeeper.SetModuleAccount(ctx, holdingAcc)

	ctx.Logger().Info("im a log!!!")
	var rstCoin = sdk.Coins{{Denom: "rst", Amount: sdk.NewInt(msg.RstAmount)}}
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, rstCoin)
	if err != nil {
		ctx.Logger().Error(err.Error())
		panic(err)
	}

	var rioCoin = sdk.Coins{{Denom: "urio", Amount: sdk.NewInt(msg.RioAmount)}}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAccAddress, rioCoin)
	if err != nil {
		ctx.Logger().Error(err.Error())
		panic(err)
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, holdingAcc.GetAddress(), rstCoin)
	if err != nil {
		ctx.Logger().Error(err.Error())
		panic(err)
	}

	fmt.Println("4444444")

	var rstStake = types.RstStake{
		Creator:            msg.Creator,
		Id:                 msg.Id,
		Address:            msg.Address,
		RstAmount:          msg.RstAmount,
		RioAmount:          msg.RioAmount,
		IncomingRstTxnHash: msg.IncomingRstTxnHash,
		Created:            msg.Created,
		Updated:            msg.Created,
		Status:             msg.Status,
	}

	k.SetRstStake(
		ctx,
		rstStake,
	)

	var valAddress = sdk.ValAddress("kdsjfhasdkhfdsak")
	delegateMsg := stakingtypes.NewMsgDelegate(holdingAcc.GetAddress(), valAddress, rioCoin)

	encCfg := simapp.MakeTestEncodingConfig()

	// Create a new TxBuilder.
	txBuilder := encCfg.TxConfig.NewTxBuilder()

	err = txBuilder.SetMsgs(delegateMsg)
	if err != nil {
		ctx.Logger().Error(err.Error())
		panic(err)
	}

	return &types.MsgCreateRstStakeResponse{}, nil
}

func (k msgServer) UpdateRstStake(goCtx context.Context, msg *types.MsgUpdateRstStake) (*types.MsgUpdateRstStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	valFound, isFound := k.GetRstStake(
		ctx,
		msg.Id,
	)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "Id not set")
	}

	// Checks if the the msg creator is the same as the current owner
	if msg.Creator != valFound.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	var rstStake = types.RstStake{
		Creator:            valFound.Creator,
		Id:                 valFound.Id,
		Address:            valFound.Address,
		RstAmount:          valFound.RstAmount,
		RioAmount:          valFound.RioAmount,
		IncomingRstTxnHash: valFound.IncomingRstTxnHash,
		Created:            valFound.Created,
		Updated:            time.Now().Unix(),
		Status:             msg.Status,
	}

	k.SetRstStake(ctx, rstStake)

	return &types.MsgUpdateRstStakeResponse{}, nil
}