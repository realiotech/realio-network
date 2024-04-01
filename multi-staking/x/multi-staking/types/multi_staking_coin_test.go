package types_test

import (
	"testing"

	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBondValue(t *testing.T) {
	testCases := []struct {
		name         string
		msCoin       types.MultiStakingCoin
		expBondValue math.Int
	}{
		{
			name:         "3001 x 0.3 = 900",
			msCoin:       types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(3001), sdk.MustNewDecFromStr("0.3")),
			expBondValue: sdk.NewInt(900),
		},
		{
			name:         "604 x 0.2 = 120",
			msCoin:       types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(604), sdk.MustNewDecFromStr("0.2")),
			expBondValue: sdk.NewInt(120),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.msCoin.BondValue(), tc.expBondValue)
		})
	}
}

func TestSafeAdd(t *testing.T) {
	testCases := []struct {
		name         string
		originMSCoin types.MultiStakingCoin
		addingMSCoin types.MultiStakingCoin
		expMSCoin    types.MultiStakingCoin
		expErr       bool
	}{
		{
			name:         "success",
			originMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(100000), sdk.MustNewDecFromStr("0.3")),
			addingMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(23456), sdk.MustNewDecFromStr("0.3")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(123456), sdk.MustNewDecFromStr("0.3")),
			expErr:       false,
		},
		{
			name:         "success and change bond weight",
			originMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(100000), sdk.OneDec()),
			addingMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(200000), sdk.MustNewDecFromStr("0.25")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(300000), sdk.MustNewDecFromStr("0.5")),
			expErr:       false,
		},
		{
			name:         "success from zero coin",
			originMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.ZeroInt(), sdk.MustNewDecFromStr("0.3")),
			addingMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(23456), sdk.MustNewDecFromStr("0.3")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(23456), sdk.MustNewDecFromStr("0.3")),
			expErr:       false,
		},
		{
			name:         "denom mismatch",
			originMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(100000), sdk.MustNewDecFromStr("0.3")),
			addingMSCoin: types.NewMultiStakingCoin(MultiStakingDenomB, sdk.NewInt(23456), sdk.MustNewDecFromStr("0.3")),
			expErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualMSCoin, err := tc.originMSCoin.SafeAdd(tc.addingMSCoin)

			if tc.expErr {
				require.Error(t, err, tc.name)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expMSCoin.Amount, actualMSCoin.Amount)
				require.Equal(t, tc.expMSCoin.Denom, actualMSCoin.Denom)
				require.Equal(t, tc.expMSCoin.BondWeight, actualMSCoin.BondWeight)
			}
		})
	}
}

func TestSafeSub(t *testing.T) {
	testCases := []struct {
		name      string
		msCoinA   types.MultiStakingCoin
		msCoinB   types.MultiStakingCoin
		expMSCoin types.MultiStakingCoin
		expErr    bool
	}{
		{
			name:      "success",
			msCoinA:   types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(123456), sdk.MustNewDecFromStr("0.3")),
			msCoinB:   types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(23456), sdk.MustNewDecFromStr("0.3")),
			expMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(100000), sdk.MustNewDecFromStr("0.3")),
			expErr:    false,
		},
		{
			name:    "denom mismatch",
			msCoinA: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(100000), sdk.MustNewDecFromStr("0.3")),
			msCoinB: types.NewMultiStakingCoin(MultiStakingDenomB, sdk.NewInt(23456), sdk.MustNewDecFromStr("0.3")),
			expErr:  true,
		},
		{
			name:    "insufficient amount",
			msCoinA: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(100000), sdk.MustNewDecFromStr("0.3")),
			msCoinB: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(234567), sdk.MustNewDecFromStr("0.3")),
			expErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualMSCoin, err := tc.msCoinA.SafeSub(tc.msCoinB)

			if tc.expErr {
				require.Error(t, err, tc.name)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expMSCoin.Amount, actualMSCoin.Amount)
				require.Equal(t, tc.expMSCoin.Denom, actualMSCoin.Denom)
				require.Equal(t, tc.expMSCoin.BondWeight, actualMSCoin.BondWeight)
			}
		})
	}
}
