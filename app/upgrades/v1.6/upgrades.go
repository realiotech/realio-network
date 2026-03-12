package v6

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	precompileFeeGrant "github.com/realiotech/realio-network/precompile/feegrant"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.6.0
func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	evmKeeper evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.6.0...")

		// Add feegrant precompile
		evmParams := evmKeeper.GetParams(sdkCtx)
		evmParams.ActiveStaticPrecompiles = append(evmParams.ActiveStaticPrecompiles, precompileFeeGrant.FeeGrantPrecompileAddress)
		err := evmKeeper.SetParams(sdkCtx, evmParams)
		if err != nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
