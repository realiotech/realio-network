package v2

import (
	"context"
	"time"

	"cosmossdk.io/math"
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
	"github.com/ethereum/go-ethereum/common"
	evmkeeper "github.com/evmos/os/x/evm/keeper"
	evmtypes "github.com/evmos/os/x/evm/types"
	evmaccount "github.com/realiotech/realio-network/crypto/account"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
	bridgekeeper "github.com/realiotech/realio-network/x/bridge/keeper"
	bridgetypes "github.com/realiotech/realio-network/x/bridge/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	paramsKeeper paramskeeper.Keeper,
	consensusKeeper consensusparamkeeper.Keeper,
	ibcKeeper ibckeeper.Keeper,
	bridgeKeeper bridgekeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	evmKeeper *evmkeeper.Keeper,
	evmStoreKey storetypes.StoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		for _, subspace := range paramsKeeper.GetSubspaces() {
			subspace := subspace

			var keyTable paramstypes.KeyTable
			switch subspace.Name() {
			case evmtypes.ModuleName:
				keyTable = evmtypes.ParamKeyTable() //nolint: staticcheck // SA1019
			case assettypes.ModuleName:
				keyTable = assettypes.ParamKeyTable()
			case bridgetypes.ModuleName:
				keyTable = bridgetypes.ParamKeyTable()
			case banktypes.ModuleName:
				keyTable = banktypes.ParamKeyTable() //nolint: staticcheck // SA1019
			case stakingtypes.ModuleName:
				keyTable = stakingtypes.ParamKeyTable() //nolint: staticcheck // SA1019
			case minttypes.ModuleName:
				keyTable = minttypes.ParamKeyTable() //nolint: staticcheck // SA1019
			case distrtypes.ModuleName:
				keyTable = distrtypes.ParamKeyTable() //nolint: staticcheck // SA1019
			case slashingtypes.ModuleName:
				keyTable = slashingtypes.ParamKeyTable() //nolint: staticcheck // SA1019
			case govtypes.ModuleName:
				keyTable = govv1.ParamKeyTable() //nolint: staticcheck // SA1019
			case crisistypes.ModuleName:
				keyTable = crisistypes.ParamKeyTable() //nolint: staticcheck // SA1019
			case authtypes.ModuleName:
				keyTable = authtypes.ParamKeyTable() //nolint: staticcheck // SA1019
			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		legacyBaseAppSubspace := paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		err := baseapp.MigrateParams(sdkCtx, legacyBaseAppSubspace, consensusKeeper.ParamsStore)
		if err != nil {
			return nil, err
		}

		legacyClientSubspace, _ := paramsKeeper.GetSubspace(exported.ModuleName)
		var params clienttypes.Params
		legacyClientSubspace.GetParamSet(sdkCtx, &params)

		params.AllowedClients = append(params.AllowedClients, exported.Localhost)
		ibcKeeper.ClientKeeper.SetParams(sdkCtx, params)

		deleteKVStore(sdkCtx.KVStore(evmStoreKey))
		delete(vm, evmtypes.ModuleName)
		MigrateEthAccountsToBaseAccounts(sdkCtx, accountKeeper, evmKeeper)

		// Run migrations and init genesis for bridge module
		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		// Update bridge genesis state
		err = bridgeKeeper.Params.Set(ctx, bridgetypes.NewParams("realio15md2mg7w62xf53gdnv7m06lpumunuhqrm5fuxl"))
		if err != nil {
			return nil, err
		}
		err = bridgeKeeper.RegisteredCoins.Set(ctx, "ario", bridgetypes.RateLimit{
			Ratelimit:     math.Int(math.NewUintFromString("1000000000000000000000000")),
			CurrentInflow: math.ZeroInt(),
		})
		if err != nil {
			return nil, err
		}
		err = bridgeKeeper.RegisteredCoins.Set(ctx, "arst", bridgetypes.RateLimit{
			Ratelimit:     math.Int(math.NewUintFromString("1000000000000000000000000")),
			CurrentInflow: math.ZeroInt(),
		})
		if err != nil {
			return nil, err
		}
		err = bridgeKeeper.EpochInfo.Set(ctx, bridgetypes.EpochInfo{
			StartTime:            time.Unix(int64(1729763876), 0),
			Duration:             time.Minute,
			EpochCountingStarted: false,
		})
		if err != nil {
			return nil, err
		}

		return newVM, nil
	}
}

// MigrateEthAccountsToBaseAccounts is used to store the code hash of the associated
// smart contracts in the dedicated store in the EVM module and convert the former
// EthAccounts to standard Cosmos SDK accounts.
func MigrateEthAccountsToBaseAccounts(ctx sdk.Context, ak authkeeper.AccountKeeper, ek *evmkeeper.Keeper) {
	ak.IterateAccounts(ctx, func(account sdk.AccountI) (stop bool) {
		ethAcc, ok := account.(*evmaccount.EthAccount)
		if !ok {
			return false
		}

		// NOTE: we only need to add store entries for smart contracts
		codeHashBytes := common.HexToHash(ethAcc.CodeHash).Bytes()
		if !evmtypes.IsEmptyCodeHash(codeHashBytes) {
			ek.SetCodeHash(ctx, ethAcc.EthAddress().Bytes(), codeHashBytes)
		}

		// Set the base account in the account keeper instead of the EthAccount
		ak.SetAccount(ctx, ethAcc.BaseAccount)

		return false
	})
}

func deleteKVStore(kv storetypes.KVStore) {
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
}
