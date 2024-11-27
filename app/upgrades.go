package app

import (
	"fmt"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/realiotech/realio-network/app/upgrades/commission"
	v1 "github.com/realiotech/realio-network/app/upgrades/v1"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/evmos/os/x/evm/types"
)

// BaseAppParamManager defines an interrace that BaseApp is expected to fullfil
// that allows upgrade handlers to modify BaseApp parameters.
type BaseAppParamManager interface {
	GetConsensusParams(ctx sdk.Context) *tmproto.ConsensusParams
	StoreConsensusParams(ctx sdk.Context, cp *tmproto.ConsensusParams)
}

// Upgrade defines a struct containing necessary fields that a SoftwareUpgradeProposal
// must have written, in order for the state migration to go smoothly.
// An upgrade must implement this struct, and then set it in the app.go.
// The app.go will then define the handler.
type Upgrade struct {
	// Upgrade version name, for the upgrade handler, e.g. `v3`
	UpgradeName string

	// CreateUpgradeHandler defines the function that creates an upgrade handler
	CreateUpgradeHandler func(*module.Manager, module.Configurator, BaseAppParamManager) upgradetypes.UpgradeHandler

	// Store upgrades, should be used for any new modules introduced, new modules deleted, or store names renamed.
	StoreUpgrades storetypes.StoreUpgrades
}

func (app *RealioNetwork) setupUpgradeHandlers() {
	// commission
	app.UpgradeKeeper.SetUpgradeHandler(
		commission.UpgradeName,
		commission.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.StakingKeeper,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v1.UpgradeName,
		v1.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.ParamsKeeper,
			app.ConsensusParamsKeeper,
			*app.IBCKeeper,
			app.BridgeKeeper,
			app.AccountKeeper,
			app.MintKeeper,
			app.EvmKeeper,
			app.keys[evmtypes.StoreKey],
		),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("failed to read upgrade info from disk: %w", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	if upgradeInfo.Name == v1.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &v1.V1StoreUpgrades))
	}
}
