package multistaking

import (
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
		appState, _, err := genutiltypes.GenesisStateFromGenFile(upgradeGenFile)
		if err != nil {
			panic("Unable to read genesis")
		}

		var bankGenesis = banktypes.GenesisState{}
		cdc.MustUnmarshalJSON(appState[banktypes.ModuleName], &bankGenesis)
		bk.InitGenesis(ctx, &bankGenesis)

		var distrGenesis = distrtypes.GenesisState{}
		cdc.MustUnmarshalJSON(appState[distrtypes.ModuleName], &distrGenesis)
		dk.InitGenesis(ctx, distrGenesis)

		var multistakingGenesis = multistakingtypes.GenesisState{}
		cdc.MustUnmarshalJSON(appState[multistakingtypes.ModuleName], &multistakingGenesis)
		msk.InitGenesis(ctx, multistakingGenesis)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
