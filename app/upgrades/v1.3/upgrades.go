package v3

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	vmtypes "github.com/cosmos/evm/x/vm/types"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.3.0
func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	accountKeeper authkeeper.AccountKeeper,
	evmKeeper evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.3.0...")

		// Add Burner permission for mint module
		err := addBurnerPermission(sdkCtx, accountKeeper, minttypes.ModuleName)
		if err != nil {
			return nil, fmt.Errorf("failed to add burner permission for mint module: %w", err)
		}

		// Migrate EVM params
		err = migrateEVMParams(sdkCtx, vm, evmKeeper)
		if err != nil {
			return nil, fmt.Errorf("failed to migrate EVM params: %w", err)
		}

		// We have no version map changes so keep current vm
		return mm.RunMigrations(ctx, cfg, vm)
	}
}

func addBurnerPermission(ctx sdk.Context, accountKeeper authkeeper.AccountKeeper, moduleName string) error {
	moduleAccount := accountKeeper.GetModuleAccount(ctx, moduleName)
	moduleAcc, ok := moduleAccount.(*authtypes.ModuleAccount)
	if !ok {
		return fmt.Errorf("not module account")
	}
	moduleAccount = authtypes.NewModuleAccount(moduleAcc.BaseAccount, moduleName, authtypes.Minter, authtypes.Burner)

	accountKeeper.SetModuleAccount(ctx, moduleAccount)
	return nil
}

func migrateEVMParams(sdkCtx sdk.Context, vm module.VersionMap, evmKeeper evmkeeper.Keeper) error {
	params := evmKeeper.GetParams(sdkCtx)
	params.ExtraEIPs = []int64{3855}
	err := evmKeeper.SetParams(sdkCtx, params)
	if err != nil {
		return err
	}

	vm[vmtypes.ModuleName] = 8
	vm[feemarkettypes.ModuleName] = 5

	return nil
}
