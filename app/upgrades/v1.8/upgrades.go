package v8

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.8.0. Its only
// non-standard step is RotateValidators (validator_rotation.go): every
// delegation on each validator being rotated out -- including its own
// operator self-bond -- gets redelegated to its replacement, through the
// real MsgBeginRedelegate path (not a direct re-key), so it goes through the
// exact same checks and safety nets an ordinary redelegation does. That
// function panics if ValidatorRotations still has an empty NewValidator
// entry, which halts this upgrade rather than silently completing it
// half-configured -- the real replacement validator addresses must be filled
// in there before this upgrade is proposed.
func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	stakingKeeper *stakingkeeper.Keeper,
	multiStakingKeeper multistakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.8.0...")

		RotateValidators(sdkCtx, stakingKeeper, multiStakingKeeper)

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
