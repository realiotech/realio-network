package multistaking

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	"github.com/spf13/cast"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	appOpts servertypes.AppOptions,
	cdc codec.Codec,
	bk bankkeeper.Keeper,
	msk multistakingkeeper.Keeper,
	dk distrkeeper.Keeper,
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

		// burn old balances on bonded address
		var bankGenesis = banktypes.GenesisState{}
		err = cdc.UnmarshalJSON(appState[banktypes.ModuleName], &bankGenesis)
		if err != nil {
			fmt.Println("bankGenesis", err)
		}
		bk.InitGenesis(ctx, &bankGenesis)

		var distrGenesis = distrtypes.GenesisState{}
		err = cdc.UnmarshalJSON(appState[distrtypes.ModuleName], &distrGenesis)
		if err != nil {
			fmt.Println("distrGenesis", err)
		}
		dk.InitGenesis(ctx, distrGenesis)

		var multistakingGenesis = multistakingtypes.GenesisState{}

		fmt.Printf("%v\n", appState["multi-staking"])

		err = cdc.UnmarshalJSON(appState["multi-staking"], &multistakingGenesis)
		if err != nil {
			fmt.Println("multistakingGenesis", err)
		}
		msk.InitGenesis(ctx, multistakingGenesis)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
