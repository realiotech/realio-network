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
	multistaking "github.com/realiotech/realio-network/app/upgrades/multi-staking"
	v3 "github.com/realiotech/realio-network/app/upgrades/v3"
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
		v3.UpgradeName,
		v3.CreateUpgradeHandler(
			app.mm, app.configurator,
			app.ConsensusParamsKeeper,
			app.IBCKeeper.ClientKeeper,
			app.ParamsKeeper,
			app.StakingKeeper,
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
	} else if upgradeInfo.Name == v3.UpgradeName {
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{crisistypes.ModuleName, consensusparamtypes.ModuleName},
		}
	}

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
