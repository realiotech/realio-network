package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/rststaking/types"
)

func (k msgServer) CreateRstStake(goCtx context.Context, msg *types.MsgCreateRstStake) (*types.MsgCreateRstStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value already exists
	_, isFound := k.GetRstStake(
		ctx,
		msg.Index,
	)
	if isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "index already set")
	}

	var rstStake = types.RstStake{
		Creator:            msg.Creator,
		Index:              msg.Index,
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
		msg.Index,
	)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	// Checks if the the msg creator is the same as the current owner
	if msg.Creator != valFound.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	var rstStake = types.RstStake{
		Creator:            msg.Creator,
		Index:              msg.Index,
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
		msg.Index,
	)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	// Checks if the the msg creator is the same as the current owner
	if msg.Creator != valFound.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.RemoveRstStake(
		ctx,
		msg.Index,
	)

	return &types.MsgDeleteRstStakeResponse{}, nil
}
