package v2

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	evmkeeper "github.com/evmos/os/x/evm/keeper"
	evmtypes "github.com/evmos/os/x/evm/types"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
	bridgetypes "github.com/realiotech/realio-network/x/bridge/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	paramskeeper paramskeeper.Keeper,
	consensuskeeper consensusparamkeeper.Keeper,
	IBCKeeper ibckeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	evmKeeper *evmkeeper.Keeper,
	EvmStoreKey storetypes.StoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		for _, subspace := range paramskeeper.GetSubspaces() {
			subspace := subspace

			var keyTable paramstypes.KeyTable
			switch subspace.Name() {
			case evmtypes.ModuleName:
				keyTable = evmtypes.ParamKeyTable() //nolint:staticcheck
			case assettypes.ModuleName:
				keyTable = assettypes.ParamKeyTable() //nolint:staticcheck
			case bridgetypes.ModuleName:
				keyTable = bridgetypes.ParamKeyTable() //nolint:staticcheck
			case banktypes.ModuleName:
				keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
			case stakingtypes.ModuleName:
				keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
			case minttypes.ModuleName:
				keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
			case distrtypes.ModuleName:
				keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
			case slashingtypes.ModuleName:
				keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
			case govtypes.ModuleName:
				keyTable = govv1.ParamKeyTable() //nolint:staticcheck
			case crisistypes.ModuleName:
				keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
			case authtypes.ModuleName:
				keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		legacyBaseAppSubspace := paramskeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		baseapp.MigrateParams(sdkCtx, legacyBaseAppSubspace, consensuskeeper.ParamsStore)

		legacyClientSubspace, _ := paramskeeper.GetSubspace(exported.ModuleName)
		var params clienttypes.Params
		legacyClientSubspace.GetParamSet(sdkCtx, &params)

		params.AllowedClients = append(params.AllowedClients, exported.Localhost)
		IBCKeeper.ClientKeeper.SetParams(sdkCtx, params)

		err := deleteKVStore(sdkCtx.KVStore(EvmStoreKey))
		if err != nil {
			return nil, err
		}

		delete(vm, evmtypes.ModuleName)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func deleteKVStore(kv storetypes.KVStore) error {
	// Note that we cannot write while iterating, so load all keys here, delete below
	var keys [][]byte
	itr := kv.Iterator(nil, nil)
	for itr.Valid() {
		keys = append(keys, itr.Key())
		itr.Next()
	}
	_ = itr.Close()

	for _, k := range keys {
		kv.Delete(k)
	}
	return nil
}
