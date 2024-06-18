package app

import (
	"fmt"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	multistaking "github.com/realiotech/realio-network/v2/app/upgrades/multi-staking"
	v2 "github.com/realiotech/realio-network/v2/app/upgrades/v2"
)

func (app *RealioNetwork) setupUpgradeHandlers(appOpts servertypes.AppOptions) {
	app.UpgradeKeeper.SetUpgradeHandler(
		multistaking.UpgradeName,
		multistaking.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			appOpts,
			app.AppCodec(),
			app.BankKeeper,
			app.AccountKeeper,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v2.UpgradeName,
		v2.CreateUpgradeHandler(
			app.mm, app.configurator,
			app.ConsensusParamsKeeper,
			app.IBCKeeper.ClientKeeper,
			app.ParamsKeeper,
			app.StakingKeeper,
			app.MultiStakingKeeper,
			app.GetSubspace(stakingtypes.ModuleName),
			app.appCodec,
		),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("failed to read upgrade info from disk: %w", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	var storeUpgrades *storetypes.StoreUpgrades

	if upgradeInfo.Name == multistaking.UpgradeName {
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{multistakingtypes.ModuleName},
		}
	} else if upgradeInfo.Name == v2.UpgradeName {
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{crisistypes.ModuleName, consensusparamtypes.ModuleName},
		}
	}

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
