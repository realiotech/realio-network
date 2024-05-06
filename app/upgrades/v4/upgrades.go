package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	sk stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade for multi staking...")
		fixMinCommisionRate(ctx, sk)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func fixMinCommisionRate(ctx sdk.Context, staking stakingkeeper.Keeper) {
	// Upgrade every validators min-commission rate
	validators := staking.GetAllValidators(ctx)
	newComm := sdk.MustNewDecFromStr(NewMinCommisionRate)
	params := staking.GetParams(ctx)
	params.MinCommissionRate = newComm
	staking.SetParams(ctx, params)
	for _, v := range validators {
		// nolint
		if v.Commission.Rate.LT(newComm) {
			comm, err := staking.UpdateValidatorCommission(ctx, v, newComm)
			if err != nil {
				panic(err)
			}

			v.Commission = comm

			// call the before-modification hook since we're about to update the commission
			staking.BeforeValidatorModified(ctx, v.GetOperator())
			staking.SetValidator(ctx, v)
		}
	}
}
