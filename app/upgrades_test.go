package app

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/realiotech/realio-network/app/upgrades/commission"
	"github.com/stretchr/testify/require"
)

func TestCommissionUpgrade(t *testing.T) {
	app := Setup(false, nil, 4)
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{Height: app.LastBlockHeight() + 1})
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)

	comm0 := stakingtypes.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr("0.01"),
		MaxRate:       math.LegacyMustNewDecFromStr("0.01"),
		MaxChangeRate: math.LegacyMustNewDecFromStr("0.02"),
	}
	comm1 := stakingtypes.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr("0.02"),
		MaxRate:       math.LegacyMustNewDecFromStr("0.03"),
		MaxChangeRate: math.LegacyMustNewDecFromStr("0.02"),
	}
	comm2 := stakingtypes.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr("0.06"),
		MaxRate:       math.LegacyMustNewDecFromStr("0.07"),
		MaxChangeRate: math.LegacyMustNewDecFromStr("0.02"),
	}
	comm3 := stakingtypes.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr("0.1"),
		MaxRate:       math.LegacyMustNewDecFromStr("0.2"),
		MaxChangeRate: math.LegacyMustNewDecFromStr("0.1"),
	}

	validators[0].Commission.CommissionRates = comm0
	validators[1].Commission.CommissionRates = comm1
	validators[2].Commission.CommissionRates = comm2
	validators[3].Commission.CommissionRates = comm3

	err = app.StakingKeeper.SetValidator(ctx, validators[0])
	require.NoError(t, err)
	err = app.StakingKeeper.SetValidator(ctx, validators[1])
	require.NoError(t, err)
	err = app.StakingKeeper.SetValidator(ctx, validators[2])
	require.NoError(t, err)
	err = app.StakingKeeper.SetValidator(ctx, validators[3])
	require.NoError(t, err)

	upgradePlan := upgradetypes.Plan{
		Name:   commission.UpgradeName,
		Height: ctx.BlockHeight(),
	}
	err = app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
	err = app.UpgradeKeeper.ApplyUpgrade(ctx, upgradePlan)
	require.NoError(t, err)

	validatorsAfter, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)

	upgradeMinCommRate := math.LegacyMustNewDecFromStr(commission.NewMinCommisionRate)
	require.Equal(t, validatorsAfter[0].Commission.CommissionRates.Rate, upgradeMinCommRate)
	require.Equal(t, validatorsAfter[1].Commission.CommissionRates.Rate, upgradeMinCommRate)
	require.Equal(t, validatorsAfter[0].Commission.CommissionRates.MaxRate, upgradeMinCommRate)
	require.Equal(t, validatorsAfter[1].Commission.CommissionRates.MaxRate, upgradeMinCommRate)
	require.Equal(t, validatorsAfter[2].Commission.CommissionRates.Rate, validators[2].Commission.CommissionRates.Rate)
	require.Equal(t, validatorsAfter[3].Commission.CommissionRates.Rate, validators[3].Commission.CommissionRates.Rate)
}
