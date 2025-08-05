package v4

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	precompileMultistaking "github.com/realiotech/realio-network/precompile/multistaking"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.3.0
func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	evmKeeper evmkeeper.Keeper,
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
		err := erc20Keeper.SetParams(sdkCtx, erc20types.DefaultParams())
		if err != nil {
			return nil, err
		}

		// Add multistaking and distributions precompiles to EVM active precompiles
		evmParams := evmKeeper.GetParams(sdkCtx)
		evmParams.ActiveStaticPrecompiles = append(evmParams.ActiveStaticPrecompiles, evmtypes.DistributionPrecompileAddress, precompileMultistaking.MultistakingPrecompileAddress)
		err = evmKeeper.SetParams(sdkCtx, evmParams)
		if err != nil {
			return nil, err
		}

		// We have no version map changes so keep current vm
		return mm.RunMigrations(ctx, cfg, vm)
	}
}
