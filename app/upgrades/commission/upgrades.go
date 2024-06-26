package commission

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	sk *stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade for multi staking...")
		fixMinCommisionRate(ctx, sk)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func fixMinCommisionRate(ctx sdk.Context, staking *stakingkeeper.Keeper) {
	// Upgrade every validators min-commission rate
	validators := staking.GetAllValidators(ctx)
	minComm := sdk.MustNewDecFromStr(NewMinCommisionRate)

	for _, v := range validators {
		//nolint
		if v.Commission.Rate.LT(minComm) {
			comm, err := updateValidatorCommission(ctx, staking, v, minComm)
			if err != nil {
				panic(err)
			}

			// call the before-modification hook since we're about to update the commission
			staking.BeforeValidatorModified(ctx, v.GetOperator())
			v.Commission = comm
			staking.SetValidator(ctx, v)
		}
	}
}

func updateValidatorCommission(ctx sdk.Context, staking *stakingkeeper.Keeper,
	validator stakingtypes.Validator, newRate sdk.Dec,
) (stakingtypes.Commission, error) {
	commission := validator.Commission
	blockTime := ctx.BlockHeader().Time

	if newRate.LT(staking.MinCommissionRate(ctx)) {
		return commission, fmt.Errorf("cannot set validator commission to less than minimum rate of %s", staking.MinCommissionRate(ctx))
	}

	commission.Rate = newRate
	if commission.MaxRate.LT(newRate) {
		commission.MaxRate = newRate
	}

	commission.UpdateTime = blockTime

	return commission, nil
}
