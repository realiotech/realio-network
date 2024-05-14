package v4

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func fixMinCommisionRate(ctx sdk.Context, staking *stakingkeeper.Keeper) {
	// Upgrade every validators min-commission rate
	validators := staking.GetAllValidators(ctx)
	minComm := sdk.MustNewDecFromStr(NewMinCommisionRate)
	params := staking.GetParams(ctx)
	params.MinCommissionRate = minComm

	err := staking.SetParams(ctx, params)
	if err != nil {
		panic(err)
	}

	for _, v := range validators {
		//nolint
		if v.Commission.Rate.LT(minComm) {
			comm, err := updateValidatorCommission(ctx, staking, v, minComm)
			if err != nil {
				panic(err)
			}

			v.Commission = comm

			// call the before-modification hook since we're about to update the commission
			staking.Hooks().BeforeValidatorModified(ctx, v.GetOperator())
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
