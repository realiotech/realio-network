package v5

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/realiotech/realio-network/app/upgrades/v1.5/legacy"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.3.0
func CreateUpgradeHandler(
	storeKey *storetypes.KVStoreKey,
	codec codec.BinaryCodec,
	mm *module.Manager,
	cfg module.Configurator,
	evmKeeper evmkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.5.0...")

		// Migrate EVM params
		err := migrateEVMParams(sdkCtx, storeKey, codec, evmKeeper)
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

func migrateEVMParams(sdkCtx sdk.Context, storeKey *storetypes.KVStoreKey, codec codec.BinaryCodec, evmKeeper evmkeeper.Keeper) error {
	store := sdkCtx.KVStore(storeKey)
	bz := store.Get(evmtypes.KeyPrefixParams)

	var legacyParams legacy.Params
	codec.MustUnmarshal(bz, &legacyParams)

	// Update EVM params
	var evmParams evmtypes.Params
	evmParams.EvmDenom = legacyParams.EvmDenom
	evmParams.ExtraEIPs = legacyParams.ExtraEIPs
	evmParams.AccessControl = evmtypes.AccessControl{
		Create: evmtypes.AccessControlType{
			AccessType:        evmtypes.AccessType(legacyParams.AccessControl.Create.AccessType),
			AccessControlList: legacyParams.AccessControl.Create.AccessControlList,
		},
		Call: evmtypes.AccessControlType{
			AccessType:        evmtypes.AccessType(legacyParams.AccessControl.Call.AccessType),
			AccessControlList: legacyParams.AccessControl.Call.AccessControlList,
		},
	}
	evmParams.EVMChannels = legacyParams.EVMChannels
	evmParams.ActiveStaticPrecompiles = legacyParams.ActiveStaticPrecompiles
	evmParams.ExtendedDenomOptions = nil
	evmParams.HistoryServeWindow = evmtypes.DefaultHistoryServeWindow

	return evmKeeper.SetParams(sdkCtx, evmParams)
}
