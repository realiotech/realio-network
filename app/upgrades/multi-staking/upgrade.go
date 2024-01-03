package multistaking

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"

	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	"github.com/spf13/cast"
)

var (
	bondedPoolAddress   = authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
	unbondedPoolAddress = authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName)
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	appOpts servertypes.AppOptions,
	cdc codec.Codec,
	bk bankkeeper.Keeper,
	msk multistakingkeeper.Keeper,
	dk distrkeeper.Keeper,
	keys map[string]*storetypes.KVStoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade for multi staking...")

		nodeHome := cast.ToString(appOpts.Get(flags.FlagHome))
		upgradeGenFile := nodeHome + "/config/state.json"
		fmt.Println(upgradeGenFile)
		appState, _, err := genutiltypes.GenesisStateFromGenFile(upgradeGenFile)
		if err != nil {
			fmt.Println(err)
			panic("Unable to read genesis")
		}

		// migrate bank
		// Send coins from bonded and unbonded pool to multistaking account

		// Mint stake to replace multi-staking coins

		// migrate distribute
		// delaccount to interdiate account
		//

		// migrate multistaking
		var multistakingGenesis = multistakingtypes.GenesisState{}
		err = cdc.UnmarshalJSON(appState["multi-staking"], &multistakingGenesis)
		if err != nil {
			fmt.Println("multistakingGenesis", err)
		}
		msk.InitGenesis(ctx, multistakingGenesis)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
