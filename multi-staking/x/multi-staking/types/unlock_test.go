package types_test

import (
	"testing"

	"github.com/realio-tech/multi-staking-module/test"
	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestFindEntryIndexByHeight(t *testing.T) {
	valAddr := test.GenValAddress()
	delAddr := test.GenAddress()
	unlockID := types.MultiStakingUnlockID(delAddr.String(), valAddr.String())
	initalEntries := []types.UnlockEntry{
		types.NewUnlockEntry(1, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(123456), sdk.MustNewDecFromStr("0.3"))),
		types.NewUnlockEntry(2, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(100000), sdk.OneDec())),
		types.NewUnlockEntry(3, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(200000), sdk.MustNewDecFromStr("0.5"))),
		types.NewUnlockEntry(4, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(300000), sdk.MustNewDecFromStr("0.5"))),
		types.NewUnlockEntry(5, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(500000), sdk.MustNewDecFromStr("0.2"))),
	}

	testCases := []struct {
		name     string
		height   int64
		expFound bool
	}{
		{
			name:     "success",
			height:   1,
			expFound: true,
		},
		{
			name:     "success and found",
			height:   4,
			expFound: true,
		},
		{
			name:     "not found",
			height:   1243,
			expFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unlockRecord := types.MultiStakingUnlock{
				UnlockID: unlockID,
				Entries:  initalEntries,
			}
			index, found := unlockRecord.FindEntryIndexByHeight(tc.height)

			if !tc.expFound {
				require.False(t, found)
				require.Equal(t, index, -1)
			} else {
				require.True(t, found)
			}
		})
	}
}

func TestAddEntry(t *testing.T) {
	valAddr := test.GenValAddress()
	delAddr := test.GenAddress()
	unlockID := types.MultiStakingUnlockID(delAddr.String(), valAddr.String())
	initalEntries := []types.UnlockEntry{
		types.NewUnlockEntry(1, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(123456), sdk.MustNewDecFromStr("0.3"))),
		types.NewUnlockEntry(2, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(100000), sdk.OneDec())),
		types.NewUnlockEntry(3, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(200000), sdk.MustNewDecFromStr("0.5"))),
		types.NewUnlockEntry(4, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(300000), sdk.MustNewDecFromStr("0.5"))),
		types.NewUnlockEntry(5, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(500000), sdk.MustNewDecFromStr("0.2"))),
	}

	testCases := []struct {
		name         string
		height       int64
		addingMSCoin types.MultiStakingCoin
		expMSCoin    types.MultiStakingCoin
		expPanic     bool
	}{
		{
			name:         "success",
			height:       1,
			addingMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(123000), sdk.MustNewDecFromStr("0.3")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(246456), sdk.MustNewDecFromStr("0.3")),
			expPanic:     false,
		},
		{
			name:         "success and change rate",
			height:       2,
			addingMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(400000), sdk.MustNewDecFromStr("0.5")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(500000), sdk.MustNewDecFromStr("0.6")),
			expPanic:     false,
		},
		{
			name:         "success add new entry",
			height:       12,
			addingMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(123456), sdk.MustNewDecFromStr("0.3")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(123456), sdk.MustNewDecFromStr("0.3")),
			expPanic:     false,
		},
		{
			name:         "denom mismatch",
			height:       3,
			addingMSCoin: types.NewMultiStakingCoin(MultiStakingDenomB, sdk.NewInt(123000), sdk.MustNewDecFromStr("0.3")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(246456), sdk.MustNewDecFromStr("0.3")),
			expPanic:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unlockRecord := types.MultiStakingUnlock{
				UnlockID: unlockID,
				Entries:  initalEntries,
			}
			if tc.expPanic {
				require.Panics(t, func() {
					unlockRecord.AddEntry(tc.height, tc.addingMSCoin)
				})
			} else {
				unlockRecord.AddEntry(tc.height, tc.addingMSCoin)

				entryIndex, found := unlockRecord.FindEntryIndexByHeight(tc.height)
				require.True(t, found)
				unlockEntry := unlockRecord.Entries[entryIndex]

				require.Equal(t, unlockEntry.UnlockingCoin.Amount, tc.expMSCoin.Amount)
				require.Equal(t, unlockEntry.UnlockingCoin.Denom, tc.expMSCoin.Denom)
				require.Equal(t, unlockEntry.UnlockingCoin.BondWeight, tc.expMSCoin.BondWeight)
			}
		})
	}
}

func TestRemoveCoinFromEntry(t *testing.T) {
	valAddr := test.GenValAddress()
	delAddr := test.GenAddress()
	unlockID := types.MultiStakingUnlockID(delAddr.String(), valAddr.String())
	initalEntries := []types.UnlockEntry{
		types.NewUnlockEntry(1, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(123456), sdk.MustNewDecFromStr("0.3"))),
		types.NewUnlockEntry(2, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(100000), sdk.OneDec())),
		types.NewUnlockEntry(3, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(200000), sdk.MustNewDecFromStr("0.5"))),
		types.NewUnlockEntry(4, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(300000), sdk.MustNewDecFromStr("0.5"))),
		types.NewUnlockEntry(5, types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(500000), sdk.MustNewDecFromStr("0.2"))),
	}

	testCases := []struct {
		name         string
		index        int
		removeMSCoin types.MultiStakingCoin
		expMSCoin    types.MultiStakingCoin
		expErr       bool
	}{
		{
			name:         "success",
			index:        0,
			removeMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(123000), sdk.MustNewDecFromStr("0.3")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(456), sdk.MustNewDecFromStr("0.3")),
			expErr:       false,
		},
		{
			name:         "success and remove all",
			index:        4,
			removeMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(500000), sdk.MustNewDecFromStr("0.3")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.ZeroInt(), sdk.MustNewDecFromStr("0.3")),
			expErr:       false,
		},
		{
			name:         "entry index is out of bound",
			index:        10,
			removeMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(400000), sdk.MustNewDecFromStr("0.3")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(246456), sdk.MustNewDecFromStr("0.3")),
			expErr:       true,
		},
		{
			name:         "remove too much",
			index:        5,
			removeMSCoin: types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(1000000), sdk.MustNewDecFromStr("0.3")),
			expMSCoin:    types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(246456), sdk.MustNewDecFromStr("0.3")),
			expErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unlockRecord := types.MultiStakingUnlock{
				UnlockID: unlockID,
				Entries:  initalEntries,
			}
			beforeLen := len(unlockRecord.Entries)
			err := unlockRecord.RemoveCoinFromEntry(tc.index, tc.removeMSCoin.Amount)

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				afterLen := len(unlockRecord.Entries)
				if beforeLen > afterLen {
					require.Equal(t, beforeLen, afterLen+1)
				} else {
					unlockEntry := unlockRecord.Entries[tc.index]

					require.Equal(t, unlockEntry.UnlockingCoin.Amount, tc.expMSCoin.Amount)
					require.Equal(t, unlockEntry.UnlockingCoin.Denom, tc.expMSCoin.Denom)
					require.Equal(t, unlockEntry.UnlockingCoin.BondWeight, tc.expMSCoin.BondWeight)
				}
			}
		})
	}
}
