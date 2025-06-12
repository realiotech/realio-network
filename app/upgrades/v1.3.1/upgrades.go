package v3

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.3.0
func CreateUpgradeHandler(
	_ *module.Manager,
	_ module.Configurator,
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

		// We have no version map changes so keep current vm
		return vm, nil
	}
}
