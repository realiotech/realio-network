package app

import (
	"fmt"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/realiotech/realio-network/app/migrations"
	"github.com/realiotech/realio-network/app/upgrades/commission"

	v2 "github.com/realiotech/realio-network/app/upgrades/v1.2"
	v3 "github.com/realiotech/realio-network/app/upgrades/v1.3"
	v4 "github.com/realiotech/realio-network/app/upgrades/v1.4"
	v5 "github.com/realiotech/realio-network/app/upgrades/v1.5"
	v6 "github.com/realiotech/realio-network/app/upgrades/v1.6"
	v7 "github.com/realiotech/realio-network/app/upgrades/v1.7"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/cosmos/evm/x/vm/types"
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
		v2.UpgradeName,
		v2.CreateUpgradeHandler(
			app.mm,
			app.configurator,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v3.UpgradeName,
		v3.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.AccountKeeper,
			*app.EvmKeeper,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v4.UpgradeName,
		v4.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			*app.EvmKeeper,
			app.Erc20Keeper,
			app.AccountKeeper,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v5.UpgradeName,
		v5.CreateUpgradeHandler(
			app.keys[evmtypes.StoreKey],
			app.appCodec,
			app.mm,
			app.configurator,
			*app.EvmKeeper,
			app.Erc20Keeper,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v6.UpgradeName,
		v6.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			*app.EvmKeeper,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v7.UpgradeName,
		v7.CreateUpgradeHandler(
			app.mm,
			app.configurator,
		),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("failed to read upgrade info from disk: %w", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	// Every hardcoded StoreUpgrades candidate this binary might need to
	// apply on this restart, keyed by the height at which it must fire.
	// SetStoreLoader overwrites rather than stacks (baseapp/options.go:279),
	// so these must be combined into a single loader rather than each
	// calling SetStoreLoader independently — otherwise whichever call runs
	// last would silently discard the other.
	candidates := []heightStoreUpgrade{
		// x/blacklist (see app/forks.go): hardcoded, not routed through
		// upgrade-info.json, so it is unconditional here — the height check
		// happens inside newStoreLoader at load time instead.
		{height: migrations.BlacklistForkHeight, upgrades: migrations.BlacklistStoreUpgrades},
	}
	if upgradeInfo.Name == v6.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		candidates = append(candidates, heightStoreUpgrade{height: upgradeInfo.Height, upgrades: v6.V6StoreUpgrades})
	}

	app.SetStoreLoader(newStoreLoader(candidates))
}

// heightStoreUpgrade pairs a StoreUpgrades with the exact height at which
// the restarting binary must apply it.
type heightStoreUpgrade struct {
	height   int64
	upgrades storetypes.StoreUpgrades
}

// newStoreLoader mirrors upgradetypes.UpgradeStoreLoader, generalized to
// pick whichever (at most one) candidate's height matches this restart's
// committed version, instead of hardcoding a single height/StoreUpgrades
// pair. Falls back to baseapp.DefaultStoreLoader if none match — the normal
// case on every restart except the exact one where a given upgrade lands.
func newStoreLoader(candidates []heightStoreUpgrade) baseapp.StoreLoader {
	return func(ms storetypes.CommitMultiStore) error {
		version := ms.LastCommitID().Version
		for _, c := range candidates {
			if c.height != version+1 {
				continue
			}
			if len(c.upgrades.Added) == 0 && len(c.upgrades.Renamed) == 0 && len(c.upgrades.Deleted) == 0 {
				continue
			}
			return ms.LoadLatestVersionAndUpgrade(&c.upgrades)
		}
		return baseapp.DefaultStoreLoader(ms)
	}
}
