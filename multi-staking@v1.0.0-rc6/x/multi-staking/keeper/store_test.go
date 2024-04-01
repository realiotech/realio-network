package keeper_test

import (
	"github.com/realio-tech/multi-staking-module/test"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	gasDenom = "ario"
	govDenom = "arst"
)

func (suite *KeeperTestSuite) TestSetBondWeight() {
	suite.SetupTest()

	gasWeight := sdk.OneDec()
	govWeight := sdk.NewDecWithPrec(2, 4)

	suite.msKeeper.SetBondWeight(suite.ctx, gasDenom, gasWeight)
	suite.msKeeper.SetBondWeight(suite.ctx, govDenom, govWeight)

	expectedGasWeight, _ := suite.msKeeper.GetBondWeight(suite.ctx, gasDenom)
	expectedGovWeight, _ := suite.msKeeper.GetBondWeight(suite.ctx, govDenom)

	suite.Equal(gasWeight, expectedGasWeight)
	suite.Equal(govWeight, expectedGovWeight)
}

func (suite *KeeperTestSuite) TestSetValidatorMultiStakingCoin() {
	valA := test.GenValAddress()
	valB := test.GenValAddress()

	testCases := []struct {
		name     string
		malleate func(ctx sdk.Context, msKeeper *multistakingkeeper.Keeper) []string
		vals     []sdk.ValAddress
		expPanic bool
	}{
		{
			name: "1 val, 1 denom, success",
			malleate: func(ctx sdk.Context, msKeeper *multistakingkeeper.Keeper) []string {
				msKeeper.SetValidatorMultiStakingCoin(ctx, valA, gasDenom)
				return []string{gasDenom}
			},
			vals:     []sdk.ValAddress{valA},
			expPanic: false,
		},
		{
			name: "2 val, 2 denom, success",
			malleate: func(ctx sdk.Context, msKeeper *multistakingkeeper.Keeper) []string {
				msKeeper.SetValidatorMultiStakingCoin(ctx, valA, gasDenom)
				msKeeper.SetValidatorMultiStakingCoin(ctx, valB, govDenom)
				return []string{gasDenom, govDenom}
			},
			vals:     []sdk.ValAddress{valA, valB},
			expPanic: false,
		},
		{
			name: "1 val, 2 denom, failed",
			malleate: func(ctx sdk.Context, msKeeper *multistakingkeeper.Keeper) []string {
				msKeeper.SetValidatorMultiStakingCoin(ctx, valA, gasDenom)
				msKeeper.SetValidatorMultiStakingCoin(ctx, valA, govDenom)
				return []string{gasDenom, govDenom}
			},
			vals:     []sdk.ValAddress{valA, valB},
			expPanic: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			if tc.expPanic {
				suite.Require().PanicsWithValue("validator multi staking coin already set", func() {
					tc.malleate(suite.ctx, suite.msKeeper)
				})
			} else {
				inputs := tc.malleate(suite.ctx, suite.msKeeper)
				for idx, val := range tc.vals {
					actualDenom := suite.msKeeper.GetValidatorMultiStakingCoin(suite.ctx, val)
					suite.Require().Equal(inputs[idx], actualDenom)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSetMultiStakingLock() {
	suite.SetupTest()
	delAddr := test.GenAddress()
	valAddr := test.GenValAddress()

	lock := types.MultiStakingLock{
		LockID: types.LockID{
			MultiStakerAddr: delAddr.String(),
			ValAddr:         valAddr.String(),
		},
		LockedCoin: types.MultiStakingCoin{
			Denom:      gasDenom,
			Amount:     sdk.NewIntFromUint64(1000000),
			BondWeight: sdk.NewDec(1),
		},
	}

	testCases := []struct {
		name     string
		malleate func()
		expError bool
	}{
		{
			"Success",
			func() {
				suite.msKeeper.SetMultiStakingLock(suite.ctx, lock)
			},
			false,
		},
	}
	for _, tc := range testCases {
		if !tc.expError {
			tc.malleate()
			msLock, found := suite.msKeeper.GetMultiStakingLock(suite.ctx, lock.LockID)
			suite.Require().True(found)
			suite.Require().Equal(lock, msLock)
		}
	}
}

func (suite *KeeperTestSuite) TestMultiStakingLockIterator() {
	valA := test.GenValAddress()
	valB := test.GenValAddress()

	delA := test.GenAddress()
	delB := test.GenAddress()

	sampleLocks := []types.MultiStakingLock{
		types.NewMultiStakingLock(
			types.MultiStakingLockID(delA.String(), valA.String()),
			types.NewMultiStakingCoin(gasDenom, sdk.NewInt(1000), sdk.OneDec()),
		),
		types.NewMultiStakingLock(
			types.MultiStakingLockID(delA.String(), valB.String()),
			types.NewMultiStakingCoin(govDenom, sdk.NewInt(1234), sdk.MustNewDecFromStr("0.3")),
		),
		types.NewMultiStakingLock(
			types.MultiStakingLockID(delB.String(), valA.String()),
			types.NewMultiStakingCoin(gasDenom, sdk.NewInt(5678), sdk.OneDec()),
		),
		types.NewMultiStakingLock(
			types.MultiStakingLockID(delB.String(), valB.String()),
			types.NewMultiStakingCoin(govDenom, sdk.NewInt(3000), sdk.MustNewDecFromStr("0.3")),
		),
	}

	suite.SetupTest()
	expLocks := make(map[string]types.MultiStakingLock)
	suite.msKeeper.MultiStakingLockIterator(suite.ctx, func(multiStakingLock types.MultiStakingLock) (stop bool) {
		mapKey := multiStakingLock.LockID.MultiStakerAddr + multiStakingLock.LockID.ValAddr
		expLocks[mapKey] = multiStakingLock
		return false
	})

	for _, lock := range sampleLocks {
		suite.msKeeper.SetMultiStakingLock(suite.ctx, lock)
		mapKey := lock.LockID.MultiStakerAddr + lock.LockID.ValAddr
		expLocks[mapKey] = lock
	}

	suite.msKeeper.MultiStakingLockIterator(suite.ctx, func(multiStakingLock types.MultiStakingLock) (stop bool) {
		mapKey := multiStakingLock.LockID.MultiStakerAddr + multiStakingLock.LockID.ValAddr
		suite.Require().Equal(expLocks[mapKey], multiStakingLock)
		return false
	})
}

func (suite *KeeperTestSuite) TestMultiStakingUnlockIterator() {
	valA := test.GenValAddress()
	valB := test.GenValAddress()

	delA := test.GenAddress()
	delB := test.GenAddress()

	sampleUnlocks := []types.MultiStakingUnlock{
		types.NewMultiStakingUnlock(
			types.MultiStakingUnlockID(delA.String(), valA.String()),
			1,
			types.NewMultiStakingCoin(gasDenom, sdk.NewInt(1000), sdk.OneDec()),
		),
		types.NewMultiStakingUnlock(
			types.MultiStakingUnlockID(delA.String(), valB.String()),
			2,
			types.NewMultiStakingCoin(govDenom, sdk.NewInt(1234), sdk.MustNewDecFromStr("0.3")),
		),
		types.NewMultiStakingUnlock(
			types.MultiStakingUnlockID(delB.String(), valA.String()),
			3,
			types.NewMultiStakingCoin(gasDenom, sdk.NewInt(5678), sdk.OneDec()),
		),
		types.NewMultiStakingUnlock(
			types.MultiStakingUnlockID(delB.String(), valB.String()),
			4,
			types.NewMultiStakingCoin(govDenom, sdk.NewInt(3000), sdk.MustNewDecFromStr("0.3")),
		),
	}

	suite.SetupTest()
	expUnlocks := make(map[string]types.MultiStakingUnlock)
	suite.msKeeper.MultiStakingUnlockIterator(suite.ctx, func(multiStakingUnlock types.MultiStakingUnlock) (stop bool) {
		mapKey := multiStakingUnlock.UnlockID.MultiStakerAddr + multiStakingUnlock.UnlockID.ValAddr
		expUnlocks[mapKey] = multiStakingUnlock
		return false
	})

	for _, unlock := range sampleUnlocks {
		suite.msKeeper.SetMultiStakingUnlock(suite.ctx, unlock)
		mapKey := unlock.UnlockID.MultiStakerAddr + unlock.UnlockID.ValAddr
		expUnlocks[mapKey] = unlock
	}

	suite.msKeeper.MultiStakingUnlockIterator(suite.ctx, func(multiStakingUnlock types.MultiStakingUnlock) (stop bool) {
		mapKey := multiStakingUnlock.UnlockID.MultiStakerAddr + multiStakingUnlock.UnlockID.ValAddr
		suite.Require().Equal(expUnlocks[mapKey], multiStakingUnlock)
		return false
	})
}

func (suite *KeeperTestSuite) TestValidatorMultiStakingCoinIterator() {
	valA := test.GenValAddress()
	valB := test.GenValAddress()
	valC := test.GenValAddress()
	valD := test.GenValAddress()

	sampleRecords := []types.ValidatorMultiStakingCoin{
		{
			ValAddr:   valA.String(),
			CoinDenom: gasDenom,
		},
		{
			ValAddr:   valB.String(),
			CoinDenom: govDenom,
		},
		{
			ValAddr:   valC.String(),
			CoinDenom: govDenom,
		},
		{
			ValAddr:   valD.String(),
			CoinDenom: govDenom,
		},
	}

	suite.SetupTest()

	expRecords := make(map[string]types.ValidatorMultiStakingCoin)
	suite.msKeeper.ValidatorMultiStakingCoinIterator(suite.ctx, func(valAddr string, denom string) (stop bool) {
		expRecords[valAddr] = types.ValidatorMultiStakingCoin{
			ValAddr:   valAddr,
			CoinDenom: denom,
		}
		return false
	})

	for _, record := range sampleRecords {
		valAcc, _ := sdk.ValAddressFromBech32(record.ValAddr)
		suite.msKeeper.SetValidatorMultiStakingCoin(suite.ctx, valAcc, record.CoinDenom)
		expRecords[record.ValAddr] = types.ValidatorMultiStakingCoin{
			ValAddr:   record.ValAddr,
			CoinDenom: record.CoinDenom,
		}
	}

	suite.msKeeper.ValidatorMultiStakingCoinIterator(suite.ctx, func(valAddr string, denom string) (stop bool) {
		suite.Require().Equal(expRecords[valAddr].ValAddr, valAddr)
		suite.Require().Equal(expRecords[valAddr].CoinDenom, denom)
		return false
	})
}
