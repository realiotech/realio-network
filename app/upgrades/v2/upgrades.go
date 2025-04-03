package v2

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmosvm "github.com/evmos/os/x/evm/core/vm"
	evmkeeper "github.com/evmos/os/x/evm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	evmKeeper *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v2...")
		evmKeeper.WithStaticPrecompiles(evmosvm.PrecompiledContractsBerlin)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
