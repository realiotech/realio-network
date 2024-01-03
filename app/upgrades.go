package app

import (
	"fmt"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	multistaking "github.com/realiotech/realio-network/app/upgrades/multi-staking"

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
			app.MultiStakingKeeper,
			app.DistrKeeper,
			app.keys,
		),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("Failed to read upgrade info from disk: %w", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	var storeUpgrades *storetypes.StoreUpgrades

	switch upgradeInfo.Name {
	case multistaking.UpgradeName:
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{multistakingtypes.ModuleName},
		}
	}

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
