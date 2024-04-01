package keeper_test

import (
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestMsUnlockEndBlocker() {
	// val A
	// delegate to val A with X ario
	// undelegate from val A
	// val A got slash
	// nextblock
	// check A balance has X ario/ zero stake

	testCases := []struct {
		name        string
		lockAmount  math.Int
		slashFactor sdk.Dec
	}{
		{
			name:        "no slashing",
			lockAmount:  math.NewInt(3788),
			slashFactor: sdk.ZeroDec(),
		},
		{
			name:        "slash half of lock coin",
			lockAmount:  math.NewInt(123),
			slashFactor: sdk.MustNewDecFromStr("0.5"),
		},
		{
			name:        "slash all of lock coin",
			lockAmount:  math.NewInt(19090),
			slashFactor: sdk.ZeroDec(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			// height 1
			suite.SetupTest()

			vals := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
			val := vals[0]

			msDenom := suite.msKeeper.GetValidatorMultiStakingCoin(suite.ctx, val.GetOperator())

			msCoin := sdk.NewCoin(msDenom, tc.lockAmount)

			msStaker := suite.CreateAndFundAccount(sdk.NewCoins(msCoin))

			delegateMsg := &stakingtypes.MsgDelegate{
				DelegatorAddress: msStaker.String(),
				ValidatorAddress: val.OperatorAddress,
				Amount:           msCoin,
			}
			_, err := suite.msgServer.Delegate(suite.ctx, delegateMsg)
			suite.NoError(err)

			// height 2
			suite.NextBlock(time.Second)

			if !tc.slashFactor.IsZero() {
				val, found := suite.app.StakingKeeper.GetValidator(suite.ctx, val.GetOperator())
				require.True(suite.T(), found)

				slashedPow := suite.app.StakingKeeper.TokensToConsensusPower(suite.ctx, val.Tokens)

				valConsAddr, err := val.GetConsAddr()
				require.NoError(suite.T(), err)

				// height 3
				suite.NextBlock(time.Second)

				suite.app.SlashingKeeper.Slash(suite.ctx, valConsAddr, tc.slashFactor, slashedPow, 2)
			} else {
				// height 3
				suite.NextBlock(time.Second)
			}

			undelegateMsg := stakingtypes.MsgUndelegate{
				DelegatorAddress: msStaker.String(),
				ValidatorAddress: val.OperatorAddress,
				Amount:           msCoin,
			}

			_, err = suite.msgServer.Undelegate(suite.ctx, &undelegateMsg)
			suite.NoError(err)

			// pass unbonding period
			suite.NextBlock(time.Duration(1000000000000000000))
			suite.NextBlock(time.Duration(1))

			unlockAmount := suite.app.BankKeeper.GetBalance(suite.ctx, msStaker, msDenom).Amount

			expectedUnlockAmount := sdk.NewDecFromInt(tc.lockAmount).Mul(sdk.OneDec().Sub(tc.slashFactor)).TruncateInt()

			suite.True(SoftEqualInt(unlockAmount, expectedUnlockAmount) || DiffLTEThanOne(unlockAmount, expectedUnlockAmount))
		})
	}
}
