package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func fixMinCommisionRate(ctx sdk.Context, staking *stakingkeeper.Keeper, stakingLegacySubspace paramstypes.Subspace) {
	// Upgrade every validators min-commission rate
	validators := staking.GetAllValidators(ctx)
	minComm := sdk.MustNewDecFromStr(NewMinCommisionRate)
	if stakingLegacySubspace.HasKeyTable() {
		stakingLegacySubspace.Set(ctx, stakingtypes.KeyMinCommissionRate, minComm)
	} else {
		stakingLegacySubspace.WithKeyTable(stakingtypes.ParamKeyTable())
		stakingLegacySubspace.Set(ctx, stakingtypes.KeyMinCommissionRate, minComm)
	}

	for _, v := range validators {
		//nolint
		if v.Commission.Rate.LT(minComm) {
			comm, err := updateValidatorCommission(ctx, v, minComm)
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

func updateValidatorCommission(ctx sdk.Context,
	validator stakingtypes.Validator, newRate sdk.Dec,
) (stakingtypes.Commission, error) {
	commission := validator.Commission
	blockTime := ctx.BlockHeader().Time

	commission.Rate = newRate
	if commission.MaxRate.LT(newRate) {
		commission.MaxRate = newRate
	}

	commission.UpdateTime = blockTime

	return commission, nil
}
