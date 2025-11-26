package v4

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	realiotypes "github.com/realiotech/realio-network/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.3.0
func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	evmKeeper evmkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.5.0...")

		// Update EVM params
		evmParams := evmKeeper.GetParams(sdkCtx)
		evmParams.EvmDenom = realiotypes.AttoRio
		evmParams.HistoryServeWindow = evmtypes.DefaultHistoryServeWindow

		err := evmKeeper.SetParams(sdkCtx, evmParams)
		if err != nil {
			return nil, err
		}

		err = evmKeeper.InitEvmCoinInfo(sdkCtx)
		if err != nil {
			return nil, err
		}

		// Update erc20 params
		erc20Params := erc20Keeper.GetParams(sdkCtx)
		// Disable permissionless registration,
		// only register new erc20 through gov
		erc20Params.PermissionlessRegistration = false
		err = erc20Keeper.SetParams(sdkCtx, erc20Params)
		if err != nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
