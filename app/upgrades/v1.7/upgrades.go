package v7

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	wasmkeeper wasmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.7.0...")

		// Set CosmWasm params
		wasmParams := wasmtypes.DefaultParams()
		wasmParams.CodeUploadAccess = wasmtypes.AllowEverybody
		wasmParams.InstantiateDefaultPermission = wasmtypes.AccessTypeEverybody
		if err := wasmkeeper.SetParams(ctx, wasmParams); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to set CosmWasm params")
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
