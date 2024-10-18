package bridge

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/realiotech/realio-network/x/bridge/keeper"
	"github.com/realiotech/realio-network/x/bridge/types"
)

// BeginBlocker of epochs module.
func BeginBlocker(goCtx context.Context, k keeper.Keeper) error {
	start := telemetry.Now()
	defer telemetry.ModuleMeasureSince(types.ModuleName, start, telemetry.MetricKeyBeginBlocker)

	epochInfo, err := k.EpochInfo.Get(goCtx)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "failed to get bridge epoch info")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if ctx.BlockTime().Before(epochInfo.StartTime) {
		return nil
	}

	// if epoch counting hasn't started, signal we need to start.
	shouldInitialEpochStart := !epochInfo.EpochCountingStarted
	epochEndTime := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
	shouldEpochStart := (ctx.BlockTime().After(epochEndTime)) || shouldInitialEpochStart
	if !shouldEpochStart {
		return nil
	}

	epochInfo.CurrentEpochStartHeight = ctx.BlockHeight()
	if shouldInitialEpochStart {
		epochInfo.EpochCountingStarted = true
		epochInfo.CurrentEpochStartTime = epochInfo.StartTime
		k.Logger(goCtx).Debug(fmt.Sprintf("Starting new epoch at height %d", epochInfo.CurrentEpochStartHeight))
	} else {
		epochInfo.CurrentEpochStartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)

		err = k.RegisteredCoins.Walk(goCtx, nil, func(denom string, ratelimit types.RateLimit) (stop bool, err error) {
			ratelimit.ResetInflow()
			err = k.RegisteredCoins.Set(goCtx, denom, ratelimit)
			if err != nil {
				k.Logger(goCtx).Error(fmt.Sprintf("Error reset ratelimit with denom %s at height %d", denom, epochInfo.CurrentEpochStartHeight))
				return true, err
			}
			return false, nil
		})
		if err != nil {
			return errorsmod.Wrap(err, "failed to reset all rate limits")
		}
	}

	return k.EpochInfo.Set(ctx, epochInfo)
}
