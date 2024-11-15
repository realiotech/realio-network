package commission

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	sk *stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for multi staking...")
		fixMinCommisionRate(sdkCtx, sk)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func fixMinCommisionRate(ctx sdk.Context, staking *stakingkeeper.Keeper) {
	// Upgrade every validators min-commission rate
	validators, err := staking.GetAllValidators(ctx)
	if err != nil {
		panic(err)
	}
	minComm := math.LegacyMustNewDecFromStr(NewMinCommisionRate)

	for _, v := range validators {
		//nolint
		if v.Commission.Rate.LT(minComm) {
			comm, err := updateValidatorCommission(ctx, staking, v, minComm)
			if err != nil {
				panic(err)
			}
			valAddr, err := staking.ValidatorAddressCodec().StringToBytes(v.GetOperator())
			if err != nil {
				panic(err)
			}

			// call the before-modification hook since we're about to update the commission
			staking.Hooks().BeforeValidatorModified(ctx, valAddr)
			v.Commission = comm
			staking.SetValidator(ctx, v)
		}
	}
}

func updateValidatorCommission(ctx sdk.Context, staking *stakingkeeper.Keeper,
	validator stakingtypes.Validator, newRate math.LegacyDec,
) (stakingtypes.Commission, error) {
	commission := validator.Commission
	blockTime := ctx.BlockHeader().Time

	minRate, err := staking.MinCommissionRate(ctx)
	if err != nil {
		return commission, err
	}
	if newRate.LT(minRate) {
		return commission, fmt.Errorf("cannot set validator commission to less than minimum rate of %s", minRate)
	}

	commission.Rate = newRate
	if commission.MaxRate.LT(newRate) {
		commission.MaxRate = newRate
	}

	commission.UpdateTime = blockTime

	return commission, nil
}
