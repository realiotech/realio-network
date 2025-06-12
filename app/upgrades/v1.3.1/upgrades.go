package v3

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	vmtypes "github.com/cosmos/evm/x/vm/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.3.0
func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	evmKeeper evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.3.1...")

		params := evmKeeper.GetParams(sdkCtx)
		params.ExtraEIPs = []int64{3855}
		err := evmKeeper.SetParams(sdkCtx, params)
		if err != nil {
			return nil, err
		}

		vm[vmtypes.ModuleName] = 8
		vm[feemarkettypes.ModuleName] = 5

		// We have no version map changes so keep current vm
		return mm.RunMigrations(ctx, cfg, vm)
	}
}
