package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ScheduleForkUpgrade executes any necessary fork logic for based upon the current
// block height and chain ID (mainnet or testnet). It sets an upgrade plan once
// the chain reaches the pre-defined upgrade height.
//
// CONTRACT: for this logic to work properly it is required to:
//
//  1. Release a non-breaking patch version so that the chain can set the scheduled upgrade plan at upgrade-height.
//  2. Release the software defined in the upgrade-info
func (app *RealioNetwork) ScheduleForkUpgrade(_ sdk.Context) {
	// NOTE: there are no testnet forks for the existing versions
	// if !types.IsMainnet(ctx.ChainID()) {
	//	return
	//}
	//
	// upgradePlan := upgradetypes.Plan{
	//	Height: ctx.BlockHeight(),
	//}
	//
	//// handle mainnet forks with their corresponding upgrade name and info
	// switch ctx.BlockHeight() {
	// case v2.MainnetUpgradeHeight:
	//	upgradePlan.Name = v2.UpgradeName
	//	upgradePlan.Info = v2.UpgradeInfo
	//default:
	//	// No-op
	//	return
	//}
	//
	//// schedule the upgrade plan to the current block height, effectively performing
	//// a hard fork that uses the upgrade handler to manage the migration.
	// if err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan); err != nil {
	//	panic(
	//		fmt.Errorf(
	//			"failed to schedule upgrade %s during BeginBlock at height %d: %w",
	//			upgradePlan.Name, ctx.BlockHeight(), err,
	//		),
	//	)
	//}
}
