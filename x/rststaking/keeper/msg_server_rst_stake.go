package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/realiotech/realio-network/x/rststaking/types"
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

	//var creatorAccAddress, err = sdk.AccAddressFromBech32(msg.Creator)
	//
	//if err != nil {
	//	return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid creator address")
	//}

	var holdingAccName = types.ModuleName + "/" + msg.Creator
	var holdingAcc = authtypes.NewEmptyModuleAccount(holdingAccName)
	k.accountKeeper.SetModuleAccount(ctx, holdingAcc)

	var rstStake = types.RstStake{
		Creator:            msg.Creator,
		Id:                 msg.Id,
		Address:            msg.Address,
		RstAmount:          msg.RstAmount,
		RioAmount:          msg.RioAmount,
		IncomingRstTxnHash: msg.IncomingRstTxnHash,
		FundedRioTxnHash:   msg.FundedRioTxnHash,
		RstOriginChain:     msg.RstOriginChain,
		RstOriginAddress:   msg.RstOriginAddress,
		Created:            msg.Created,
		Status:             msg.Status,
	}

	k.SetRstStake(
		ctx,
		rstStake,
	)
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
		Creator:            msg.Creator,
		Id:                 msg.Id,
		Address:            msg.Address,
		RstAmount:          msg.RstAmount,
		RioAmount:          msg.RioAmount,
		IncomingRstTxnHash: msg.IncomingRstTxnHash,
		FundedRioTxnHash:   msg.FundedRioTxnHash,
		RstOriginChain:     msg.RstOriginChain,
		RstOriginAddress:   msg.RstOriginAddress,
		Created:            msg.Created,
		Status:             msg.Status,
	}

	k.SetRstStake(ctx, rstStake)

	return &types.MsgUpdateRstStakeResponse{}, nil
}

func (k msgServer) DeleteRstStake(goCtx context.Context, msg *types.MsgDeleteRstStake) (*types.MsgDeleteRstStakeResponse, error) {
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

	k.RemoveRstStake(
		ctx,
		msg.Id,
	)

	return &types.MsgDeleteRstStakeResponse{}, nil
}
