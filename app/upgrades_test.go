package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	v4 "github.com/realiotech/realio-network/app/upgrades/v4"
)

func TestV4Upgrade(t *testing.T) {
	app := Setup(false, nil, 4)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: app.LastBlockHeight() + 1})

	upgradePlan := upgradetypes.Plan{
		Name:   v4.UpgradeName,
		Height: ctx.BlockHeight(),
	}

	err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
	require.NoError(t, err)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgradePlan)

	validators := app.StakingKeeper.GetAllValidators(ctx)

	upgradeMinCommRate := sdk.MustNewDecFromStr(v4.NewMinCommisionRate)
	for _, val := range validators {
		require.Equal(t, val.Commission.CommissionRates.Rate, upgradeMinCommRate)
	}
}
