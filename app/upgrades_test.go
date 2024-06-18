package app

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	v2 "github.com/realiotech/realio-network/v2/app/upgrades/v2"
	"github.com/stretchr/testify/require"
)

func TestV2Upgrade(t *testing.T) {
	app := Setup(false, nil, 4)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: app.LastBlockHeight() + 1})
	validators := app.StakingKeeper.GetAllValidators(ctx)

	comm0 := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.01"),
		MaxRate:       sdk.MustNewDecFromStr("0.01"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.02"),
	}
	comm1 := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.02"),
		MaxRate:       sdk.MustNewDecFromStr("0.03"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.02"),
	}
	comm2 := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.06"),
		MaxRate:       sdk.MustNewDecFromStr("0.07"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.02"),
	}
	comm3 := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.1"),
		MaxRate:       sdk.MustNewDecFromStr("0.2"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.1"),
	}

	validators[0].Commission.CommissionRates = comm0
	validators[1].Commission.CommissionRates = comm1
	validators[2].Commission.CommissionRates = comm2
	validators[3].Commission.CommissionRates = comm3

	app.StakingKeeper.SetValidator(ctx, validators[0])
	app.StakingKeeper.SetValidator(ctx, validators[1])
	app.StakingKeeper.SetValidator(ctx, validators[2])
	app.StakingKeeper.SetValidator(ctx, validators[3])

	upgradePlan := upgradetypes.Plan{
		Name:   v2.UpgradeName,
		Height: ctx.BlockHeight(),
	}
	err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)

	require.NoError(t, err)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgradePlan)

	validatorsAfter := app.StakingKeeper.GetAllValidators(ctx)

	upgradeMinCommRate := sdk.MustNewDecFromStr(v2.NewMinCommisionRate)
	require.Equal(t, validatorsAfter[0].Commission.CommissionRates.Rate, upgradeMinCommRate)
	require.Equal(t, validatorsAfter[1].Commission.CommissionRates.Rate, upgradeMinCommRate)
	require.Equal(t, validatorsAfter[0].Commission.CommissionRates.MaxRate, upgradeMinCommRate)
	require.Equal(t, validatorsAfter[1].Commission.CommissionRates.MaxRate, upgradeMinCommRate)
	require.Equal(t, validatorsAfter[2].Commission.CommissionRates.Rate, validators[2].Commission.CommissionRates.Rate)
	require.Equal(t, validatorsAfter[3].Commission.CommissionRates.Rate, validators[3].Commission.CommissionRates.Rate)
}
