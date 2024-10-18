package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/realiotech/realio-network/x/bridge/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams updates the params.
func (ms msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := ms.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (ms msgServer) UpdateEpochDuration(ctx context.Context, msg *types.MsgUpdateEpochDuration) (*types.MsgUpdateEpochDurationResponse, error) {
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	if msg.Duration == 0 {
		return nil, types.ErrEpochDurationZero
	}

	epochInfo, err := ms.EpochInfo.Get(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "failed to get bridge epoch info")
	}
	epochInfo.Duration = msg.Duration
	err = ms.EpochInfo.Set(ctx, epochInfo)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateEpochDurationResponse{}, nil
}
