package app

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/realiotech/realio-network/app/upgrades/commission"

	upgradetypes "cosmossdk.io/x/upgrade/types"
)

func (app *RealioNetwork) setupUpgradeHandlers(appOpts servertypes.AppOptions) {
	app.UpgradeKeeper.SetUpgradeHandler(
		commission.UpgradeName,
		commission.CreateUpgradeHandler(
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

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
