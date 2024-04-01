package keeper_test

import (
	"time"

	"github.com/realio-tech/multi-staking-module/test"
	"github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestModuleAccountInvariants() {
	delAddr := test.GenAddress()
	priv, valAddr := test.GenValAddressWithPrivKey()
	valPubKey := priv.PubKey()

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name:     "Success",
			malleate: func() {},
			expPass:  true,
		},
		{
			name: "Success Edit Validator",
			malleate: func() {
				suite.ctx = suite.ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
				newRate := sdk.MustNewDecFromStr("0.03")
				newMinSelfDelegation := sdk.NewInt(300)
				editMsg := stakingtypes.NewMsgEditValidator(valAddr, stakingtypes.Description{
					Moniker:         "test 1",
					Identity:        "test 1",
					Website:         "test 1",
					SecurityContact: "test 1",
					Details:         "test 1",
				},
					&newRate,
					&newMinSelfDelegation,
				)
				_, err := suite.msgServer.EditValidator(suite.ctx, editMsg)
				suite.Require().NoError(err)
			},
			expPass: true,
		},
		{
			name: "Success Delegate",
			malleate: func() {
				bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
				delMsg := stakingtypes.NewMsgDelegate(delAddr, valAddr, bondAmount)
				_, err := suite.msgServer.Delegate(suite.ctx, delMsg)
				suite.Require().NoError(err)
			},
			expPass: true,
		},
		{
			name: "Success Delegate",
			malleate: func() {
				bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
				delMsg := stakingtypes.NewMsgDelegate(delAddr, valAddr, bondAmount)
				_, err := suite.msgServer.Delegate(suite.ctx, delMsg)
				suite.Require().NoError(err)
			},
			expPass: true,
		},
		{
			name: "Success BeginRedelegate",
			malleate: func() {
				priv, valAddr2 := test.GenValAddressWithPrivKey()
				valPubKey2 := priv.PubKey()
				bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				createMsg2 := stakingtypes.MsgCreateValidator{
					Description: stakingtypes.Description{
						Moniker:         "test",
						Identity:        "test",
						Website:         "test",
						SecurityContact: "test",
						Details:         "test",
					},
					Commission: stakingtypes.CommissionRates{
						Rate:          sdk.MustNewDecFromStr("0.05"),
						MaxRate:       sdk.MustNewDecFromStr("0.1"),
						MaxChangeRate: sdk.MustNewDecFromStr("0.1"),
					},
					MinSelfDelegation: sdk.NewInt(200),
					DelegatorAddress:  delAddr.String(),
					ValidatorAddress:  valAddr2.String(),
					Pubkey:            codectypes.UnsafePackAny(valPubKey2),
					Value:             bondAmount,
				}

				_, err := suite.msgServer.CreateValidator(suite.ctx, &createMsg2)
				suite.Require().NoError(err)

				multiStakingMsg := stakingtypes.NewMsgBeginRedelegate(delAddr, valAddr, valAddr2, bondAmount)
				_, err = suite.msgServer.BeginRedelegate(suite.ctx, multiStakingMsg)
				suite.Require().NoError(err)
			},
			expPass: true,
		},
		{
			name: "Success Undelegate",
			malleate: func() {
				bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(250))
				multiStakingMsg := stakingtypes.NewMsgUndelegate(delAddr, valAddr, bondAmount)
				_, err := suite.msgServer.Undelegate(suite.ctx, multiStakingMsg)
				suite.Require().NoError(err)

				bondAmount1 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				multiStakingMsg1 := stakingtypes.NewMsgUndelegate(delAddr, valAddr, bondAmount1)
				_, err = suite.msgServer.Undelegate(suite.ctx, multiStakingMsg1)
				suite.Require().NoError(err)
			},
			expPass: true,
		},
		{
			name: "Fail invariant",
			malleate: func() {
				multiStakingLock := types.NewMultiStakingLock(types.MultiStakingLockID(delAddr.String(), valAddr.String()), types.NewMultiStakingCoin(MultiStakingDenomA, sdk.NewInt(200), sdk.OneDec()))
				suite.app.MultiStakingKeeper.SetMultiStakingLock(suite.ctx, multiStakingLock)
			},
			expPass: false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset

		valCoins := sdk.NewCoins(sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(10000)), sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(10000)))
		suite.FundAccount(delAddr, valCoins)

		suite.msKeeper.SetBondWeight(suite.ctx, MultiStakingDenomA, sdk.MustNewDecFromStr("0.3"))
		bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(3001))
		msg := stakingtypes.MsgCreateValidator{
			Description: stakingtypes.Description{
				Moniker:         "test",
				Identity:        "test",
				Website:         "test",
				SecurityContact: "test",
				Details:         "test",
			},
			Commission: stakingtypes.CommissionRates{
				Rate:          sdk.MustNewDecFromStr("0.05"),
				MaxRate:       sdk.MustNewDecFromStr("0.1"),
				MaxChangeRate: sdk.MustNewDecFromStr("0.05"),
			},
			MinSelfDelegation: sdk.NewInt(1),
			DelegatorAddress:  delAddr.String(),
			ValidatorAddress:  valAddr.String(),
			Pubkey:            codectypes.UnsafePackAny(valPubKey),
			Value:             bondAmount,
		}

		_, err := suite.msgServer.CreateValidator(suite.ctx, &msg)
		suite.Require().NoError(err)

		tc.malleate()
		_, broken := keeper.ModuleAccountInvariants(*suite.msKeeper)(suite.ctx)

		if tc.expPass {
			suite.Require().False(broken)
		}
	}
}
