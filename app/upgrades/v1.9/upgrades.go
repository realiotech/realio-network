package v9

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/evm/x/erc20"
	erc20types "github.com/cosmos/evm/x/erc20/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	wasmkeeper wasmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.9.0...")

		// x/erc20 state exists since v1.4 but the module was never in the
		// module manager. Without a version, RunMigrations runs its InitGenesis
		// with default params, which turns permissionless registration back on.
		if _, ok := vm[erc20types.ModuleName]; !ok {
			vm[erc20types.ModuleName] = erc20.AppModuleBasic{}.ConsensusVersion()
		}

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Set CosmWasm params after RunMigrations, whose x/wasm InitGenesis
		// writes the default params.
		uploader, err := sdk.AccAddressFromBech32(CodeUploader)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "invalid CosmWasm code uploader")
		}
		wasmParams := wasmtypes.Params{
			CodeUploadAccess:             wasmtypes.AccessTypeAnyOfAddresses.With(uploader),
			InstantiateDefaultPermission: wasmtypes.AccessTypeAnyOfAddresses,
		}
		if err := wasmkeeper.SetParams(ctx, wasmParams); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to set CosmWasm params")
		}

		return vm, nil
	}
}
