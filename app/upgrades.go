package app

import (
	"fmt"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	multistaking "github.com/realiotech/realio-network/app/upgrades/multi-staking"
	v4 "github.com/realiotech/realio-network/app/upgrades/v4"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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
		v4.UpgradeName,
		v4.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.StakingKeeper,
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
	}

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
