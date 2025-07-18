package v4

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	vmtypes "github.com/cosmos/evm/x/vm/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.3.0
func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	erc20Keeper erc20keeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.4.0...")

		// Set erc20 module account
		if acc := accountKeeper.GetModuleAccount(ctx, erc20types.ModuleName); acc == nil {
			return nil, fmt.Errorf("the erc20 module account has not been set")
		}

		// Set erc20 module params, mostly EnableErc20 = true to enable Erc20 registration
		erc20Keeper.SetParams(sdkCtx, erc20types.DefaultParams())

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
