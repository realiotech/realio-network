package app

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/evm/x/erc20"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	"github.com/realiotech/realio-network/app/upgrades/commission"
	v9 "github.com/realiotech/realio-network/app/upgrades/v1.9"
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

func TestV19Upgrade(t *testing.T) {
	app := Setup(false, nil, 1)
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{Height: app.LastBlockHeight() + 1})

	// Mainnet before v1.9.0: erc20, wasm and 08-wasm are not in the version
	// map, and erc20 permissionless registration is off.
	versionMap := prefix.NewStore(ctx.KVStore(app.GetKey(upgradetypes.StoreKey)), []byte{upgradetypes.VersionMapByte})
	for _, name := range []string{erc20types.ModuleName, wasmtypes.ModuleName, ibcwasmtypes.ModuleName} {
		versionMap.Delete([]byte(name))
	}
	erc20Params := app.Erc20Keeper.GetParams(ctx)
	erc20Params.PermissionlessRegistration = false
	require.NoError(t, app.Erc20Keeper.SetParams(ctx, erc20Params))

	plan := upgradetypes.Plan{Name: v9.UpgradeName, Height: ctx.BlockHeight()}
	require.NoError(t, app.UpgradeKeeper.ScheduleUpgrade(ctx, plan))
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, plan))

	// erc20 joins the version map without InitGenesis resetting its params.
	require.Equal(t, erc20Params, app.Erc20Keeper.GetParams(ctx))
	vm, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	require.Equal(t, erc20.AppModuleBasic{}.ConsensusVersion(), vm[erc20types.ModuleName])

	// Only CodeUploader can upload, and code instantiates only for its creator.
	uploader := sdk.MustAccAddressFromBech32(v9.CodeUploader)
	wasmParams := app.WasmKeeper.GetParams(ctx)
	require.Equal(t, wasmtypes.AccessTypeAnyOfAddresses.With(uploader), wasmParams.CodeUploadAccess)
	require.Equal(t, wasmtypes.AccessTypeAnyOfAddresses, wasmParams.InstantiateDefaultPermission)
	require.True(t, wasmParams.CodeUploadAccess.Allowed(uploader))
	require.False(t, wasmParams.CodeUploadAccess.Allowed(sdk.AccAddress(make([]byte, 20))))
}
