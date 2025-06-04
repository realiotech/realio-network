package v3

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.3.0
func CreateUpgradeHandler(
	_ *module.Manager,
	_ module.Configurator,
	accountKeeper authkeeper.AccountKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.3.0...")

		// Add Burner permission for mint module
		mintModule := accountKeeper.GetModuleAccount(ctx, minttypes.ModuleName)
		mintAcc, ok := mintModule.(*authtypes.ModuleAccount)
		if !ok {
			return nil, fmt.Errorf("not module account")
		}
		mintModule = authtypes.NewModuleAccount(mintAcc.BaseAccount, minttypes.ModuleName, authtypes.Minter, authtypes.Burner)

		// Overwrite
		accountKeeper.SetModuleAccount(ctx, mintModule)

		// We have no version map changes so keep current vm
		return vm, nil
	}
}
