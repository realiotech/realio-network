package keeper_test

import (
	"time"

	"github.com/realio-tech/multi-staking-module/test"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	MultiStakingDenomA = "ario"
	MultiStakingDenomB = "arst"
)

func (suite *KeeperTestSuite) TestCreateValidator() {
	delAddr := test.GenAddress()
	valPubKey := test.GenPubKey()
	valAddr := sdk.ValAddress(valPubKey.Address())

	testCases := []struct {
		name     string
		malleate func(ctx sdk.Context, msKeeper *multistakingkeeper.Keeper, msgServer stakingtypes.MsgServer) (sdk.Coin, error)
		expOut   sdk.Coin
		expErr   bool
	}{
		{
			name: "3001 token, weight 0.3, expect 900",
			malleate: func(ctx sdk.Context, msKeeper *multistakingkeeper.Keeper, msgServer stakingtypes.MsgServer) (sdk.Coin, error) {
				msKeeper.SetBondWeight(ctx, MultiStakingDenomA, sdk.MustNewDecFromStr("0.3"))
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

				_, err := msgServer.CreateValidator(ctx, &msg)
				return bondAmount, err
			},
			expOut: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(900)),
			expErr: false,
		},
		{
			name: "25 token, weight 0.5, expect 12",
			malleate: func(ctx sdk.Context, msKeeper *multistakingkeeper.Keeper, msgServer stakingtypes.MsgServer) (sdk.Coin, error) {
				msKeeper.SetBondWeight(ctx, MultiStakingDenomB, sdk.MustNewDecFromStr("0.5"))
				bondAmount := sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(25))

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

				_, err := msgServer.CreateValidator(ctx, &msg)
				return bondAmount, err
			},
			expOut: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(12)),
			expErr: false,
		},
		{
			name: "invalid bond token",
			malleate: func(ctx sdk.Context, msKeeper *multistakingkeeper.Keeper, msgServer stakingtypes.MsgServer) (sdk.Coin, error) {
				msKeeper.RemoveBondWeight(ctx, MultiStakingDenomA)
				bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(25))

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
				_, err := msgServer.CreateValidator(ctx, &msg)
				return bondAmount, err
			},
			expOut: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(12)),
			expErr: true,
		},
		{
			name: "invalid validator address",
			malleate: func(ctx sdk.Context, msKeeper *multistakingkeeper.Keeper, msgServer stakingtypes.MsgServer) (sdk.Coin, error) {
				msKeeper.SetBondWeight(ctx, MultiStakingDenomA, sdk.MustNewDecFromStr("0.3"))
				bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(3001))

				msg := stakingtypes.MsgCreateValidator{
					Description: stakingtypes.Description{
						Moniker: "NewValidator",
					},
					Commission: stakingtypes.CommissionRates{
						Rate:          sdk.MustNewDecFromStr("0.05"),
						MaxRate:       sdk.MustNewDecFromStr("0.1"),
						MaxChangeRate: sdk.MustNewDecFromStr("0.05"),
					},
					MinSelfDelegation: sdk.NewInt(1),
					DelegatorAddress:  delAddr.String(),
					ValidatorAddress:  sdk.AccAddress([]byte("invalid")).String(),
					Pubkey:            codectypes.UnsafePackAny(valPubKey),
					Value:             bondAmount,
				}

				_, err := msgServer.CreateValidator(ctx, &msg)
				return bondAmount, err
			},
			expOut: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(12)),
			expErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			valCoins := sdk.NewCoins(sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(10000)), sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(10000)))
			suite.FundAccount(delAddr, valCoins)

			bondAmount, err := tc.malleate(suite.ctx, suite.msKeeper, suite.msgServer)
			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				lockId := multistakingtypes.MultiStakingLockID(delAddr.String(), valAddr.String())
				lockRecord, found := suite.msKeeper.GetMultiStakingLock(suite.ctx, lockId)
				suite.Require().True(found)
				actualBond, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().True(found)
				suite.Require().Equal(bondAmount.Amount, lockRecord.LockedCoin.Amount)
				suite.Require().Equal(tc.expOut.Amount, actualBond.Shares.TruncateInt())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEditValidator() {
	delAddr := test.GenAddress()
	valPubKey := test.GenPubKey()
	valAddr := sdk.ValAddress(valPubKey.Address())

	testCases := []struct {
		name     string
		malleate func(ctx sdk.Context, msgServer stakingtypes.MsgServer) (stakingtypes.MsgEditValidator, error)
		expErr   bool
	}{
		{
			name: "success",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer) (stakingtypes.MsgEditValidator, error) {
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
				_, err := msgServer.EditValidator(ctx, editMsg)
				return *editMsg, err
			},
			expErr: false,
		},
		{
			name: "not found validator",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer) (stakingtypes.MsgEditValidator, error) {
				newRate := sdk.MustNewDecFromStr("0.03")
				newMinSelfDelegation := sdk.NewInt(300)
				editMsg := stakingtypes.NewMsgEditValidator(test.GenValAddress(), stakingtypes.Description{
					Moniker:         "test",
					Identity:        "test",
					Website:         "test",
					SecurityContact: "test",
					Details:         "test",
				},
					&newRate,
					&newMinSelfDelegation,
				)
				_, err := msgServer.EditValidator(ctx, editMsg)
				return *editMsg, err
			},
			expErr: true,
		},
		{
			name: "negative rate",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer) (stakingtypes.MsgEditValidator, error) {
				newRate := sdk.MustNewDecFromStr("-0.01")
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
				_, err := msgServer.EditValidator(ctx, editMsg)
				return *editMsg, err
			},
			expErr: true,
		},
		{
			name: "less than minimum rate",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer) (stakingtypes.MsgEditValidator, error) {
				newRate := sdk.MustNewDecFromStr("0.01")
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
				_, err := msgServer.EditValidator(ctx, editMsg)
				return *editMsg, err
			},
			expErr: true,
		},
		{
			name: "more than max rate",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer) (stakingtypes.MsgEditValidator, error) {
				newRate := sdk.MustNewDecFromStr("0.11")
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
				_, err := msgServer.EditValidator(ctx, editMsg)
				return *editMsg, err
			},
			expErr: true,
		},
		{
			name: "min self delegation more than validator tokens",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer) (stakingtypes.MsgEditValidator, error) {
				newRate := sdk.MustNewDecFromStr("0.03")
				newMinSelfDelegation := sdk.NewInt(10000)
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
				_, err := msgServer.EditValidator(ctx, editMsg)
				return *editMsg, err
			},
			expErr: true,
		},
		{
			name: "min self delegation more than old min delegation value",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer) (stakingtypes.MsgEditValidator, error) {
				newRate := sdk.MustNewDecFromStr("0.03")
				newMinSelfDelegation := sdk.NewInt(100)
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
				_, err := msgServer.EditValidator(ctx, editMsg)
				return *editMsg, err
			},
			expErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			newParam := stakingtypes.DefaultParams()
			newParam.MinCommissionRate = sdk.MustNewDecFromStr("0.02")
			suite.app.StakingKeeper.SetParams(suite.ctx, newParam)
			suite.msKeeper.SetBondWeight(suite.ctx, MultiStakingDenomA, sdk.OneDec())
			bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
			suite.FundAccount(delAddr, sdk.NewCoins(bondAmount))

			createMsg := stakingtypes.MsgCreateValidator{
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
				MinSelfDelegation: sdk.NewInt(200),
				DelegatorAddress:  delAddr.String(),
				ValidatorAddress:  valAddr.String(),
				Pubkey:            codectypes.UnsafePackAny(valPubKey),
				Value:             bondAmount,
			}
			_, err := suite.msgServer.CreateValidator(suite.ctx, &createMsg)
			suite.Require().NoError(err)

			suite.ctx = suite.ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
			originMsg, err := tc.malleate(suite.ctx, suite.msgServer)

			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				validatorInfo, found := suite.app.StakingKeeper.GetValidator(suite.ctx, sdk.ValAddress(originMsg.ValidatorAddress))
				if found {
					suite.Require().Equal(validatorInfo.Description, originMsg.Description)
					suite.Require().Equal(validatorInfo.MinSelfDelegation, &originMsg.MinSelfDelegation)
					suite.Require().Equal(validatorInfo.Commission.CommissionRates.Rate, &originMsg.CommissionRate)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDelegate() {
	delAddr := test.GenAddress()
	valDelAddr := test.GenAddress()
	valPubKey := test.GenPubKey()
	valAddr := sdk.ValAddress(valPubKey.Address())

	testCases := []struct {
		name     string
		malleate func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) (sdk.Coin, error)
		expRate  sdk.Dec
		expErr   bool
	}{
		{
			name: "success and not change rate",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) (sdk.Coin, error) {
				multiStakingAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
				delMsg := stakingtypes.NewMsgDelegate(delAddr, valAddr, multiStakingAmount)
				_, err := msgServer.Delegate(ctx, delMsg)
				return multiStakingAmount, err
			},
			expRate: sdk.OneDec(),
			expErr:  false,
		},
		{
			name: "rate change from 1 to 0.75 (1000 * 1 + 3000 * 0.5 = 4000 * 0.625)",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) (sdk.Coin, error) {
				multiStakingAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
				delMsg := stakingtypes.NewMsgDelegate(delAddr, valAddr, multiStakingAmount)
				_, err := msgServer.Delegate(ctx, delMsg)
				if err != nil {
					return multiStakingAmount, err
				}
				msKeeper.SetBondWeight(ctx, MultiStakingDenomA, sdk.MustNewDecFromStr("0.5"))
				multiStakingAmount1 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(3000))
				delMsg1 := stakingtypes.NewMsgDelegate(delAddr, valAddr, multiStakingAmount1)
				_, err = msgServer.Delegate(ctx, delMsg1)
				return multiStakingAmount.Add(multiStakingAmount1), err
			},
			expRate: sdk.MustNewDecFromStr("0.625"),
			expErr:  false,
		},
		{
			name: "not found validator",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) (sdk.Coin, error) {
				multiStakingAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))

				delMsg := stakingtypes.NewMsgDelegate(delAddr, test.GenValAddress(), multiStakingAmount)
				_, err := msgServer.Delegate(ctx, delMsg)
				return multiStakingAmount, err
			},
			expErr: true,
		},
		{
			name: "not allow token",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) (sdk.Coin, error) {
				multiStakingAmount := sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(1000))

				delMsg := stakingtypes.NewMsgDelegate(delAddr, valAddr, multiStakingAmount)
				_, err := msgServer.Delegate(ctx, delMsg)
				return multiStakingAmount, err
			},
			expErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			newParam := stakingtypes.DefaultParams()
			newParam.MinCommissionRate = sdk.MustNewDecFromStr("0.02")
			suite.app.StakingKeeper.SetParams(suite.ctx, newParam)
			suite.msKeeper.SetBondWeight(suite.ctx, MultiStakingDenomA, sdk.OneDec())
			bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
			userBalance := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(10000))
			suite.FundAccount(delAddr, sdk.NewCoins(userBalance))
			suite.FundAccount(valDelAddr, sdk.NewCoins(userBalance))

			createMsg := stakingtypes.MsgCreateValidator{
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
				MinSelfDelegation: sdk.NewInt(200),
				DelegatorAddress:  valDelAddr.String(),
				ValidatorAddress:  valAddr.String(),
				Pubkey:            codectypes.UnsafePackAny(valPubKey),
				Value:             bondAmount,
			}
			_, err := suite.msgServer.CreateValidator(suite.ctx, &createMsg)
			suite.Require().NoError(err)

			multiStakingAmount, err := tc.malleate(suite.ctx, suite.msgServer, *suite.msKeeper)

			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				lockId := multistakingtypes.MultiStakingLockID(delAddr.String(), valAddr.String())
				lockRecord, found := suite.msKeeper.GetMultiStakingLock(suite.ctx, lockId)
				suite.Require().True(found)
				suite.Require().Equal(tc.expRate, lockRecord.GetBondWeight())

				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().True(found)
				validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
				suite.Require().True(found)

				multiStakingCoin := multistakingtypes.NewMultiStakingCoin(multiStakingAmount.Denom, multiStakingAmount.Amount, tc.expRate)
				expShares, err := validator.SharesFromTokens(multiStakingCoin.BondValue())
				suite.Require().NoError(err)
				suite.Require().Equal(expShares, delegation.GetShares())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBeginRedelegate() {
	delAddr := test.GenAddress()
	valDelAddr := test.GenAddress()
	valPubKey1 := test.GenPubKey()
	valPubKey2 := test.GenPubKey()

	valAddr1 := sdk.ValAddress(valPubKey1.Address())
	valAddr2 := sdk.ValAddress(valPubKey2.Address())

	testCases := []struct {
		name     string
		malleate func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) ([]sdk.Coin, error)
		expRate  []sdk.Dec
		expLock  []math.Int
		expErr   bool
	}{
		{
			name: "redelegate from val1 to val2",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) ([]sdk.Coin, error) {
				multiStakingAmount1 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
				delMsg := stakingtypes.NewMsgDelegate(delAddr, valAddr1, multiStakingAmount1)
				_, err := msgServer.Delegate(ctx, delMsg)
				suite.Require().NoError(err)

				multiStakingAmount2 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				redelegateMsg := stakingtypes.NewMsgBeginRedelegate(delAddr, valAddr1, valAddr2, multiStakingAmount2)
				_, err = msgServer.BeginRedelegate(ctx, redelegateMsg)
				return []sdk.Coin{multiStakingAmount1.Sub(multiStakingAmount2), multiStakingAmount2}, err
			},
			expRate: []sdk.Dec{sdk.OneDec(), sdk.OneDec()},
			expLock: []math.Int{sdk.NewInt(500), sdk.NewInt(500)},
			expErr:  false,
		},
		{
			name: "delegate 2000 more to val1 then change rate and redelegate 600 to val2",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) ([]sdk.Coin, error) {
				multiStakingAmount1 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
				delMsg1 := stakingtypes.NewMsgDelegate(delAddr, valAddr1, multiStakingAmount1)
				_, err := msgServer.Delegate(ctx, delMsg1)
				suite.Require().NoError(err)

				multiStakingAmount2 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
				delMsg3 := stakingtypes.NewMsgDelegate(delAddr, valAddr2, multiStakingAmount2)
				_, err = msgServer.Delegate(ctx, delMsg3)
				suite.Require().NoError(err)

				msKeeper.SetBondWeight(ctx, MultiStakingDenomA, sdk.MustNewDecFromStr("0.25"))
				multiStakingAmount3 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(2000))
				delMsg2 := stakingtypes.NewMsgDelegate(delAddr, valAddr1, multiStakingAmount3)
				_, err = msgServer.Delegate(ctx, delMsg2)
				suite.Require().NoError(err)

				multiStakingAmount4 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(600))
				redelMsg := stakingtypes.NewMsgBeginRedelegate(delAddr, valAddr1, valAddr2, multiStakingAmount4)
				if err != nil {
					return []sdk.Coin{}, err
				}

				_, err = msgServer.BeginRedelegate(ctx, redelMsg)
				return []sdk.Coin{multiStakingAmount1.Add(multiStakingAmount3).Sub(multiStakingAmount4), multiStakingAmount2.Add(multiStakingAmount4)}, err
			},
			expRate: []sdk.Dec{sdk.MustNewDecFromStr("0.5"), sdk.MustNewDecFromStr("0.8125")},
			expLock: []math.Int{sdk.NewInt(2400), sdk.NewInt(1600)},
			expErr:  false,
		},
		{
			name: "not found validator",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) ([]sdk.Coin, error) {
				bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				multiStakingMsg := stakingtypes.NewMsgBeginRedelegate(delAddr, valAddr1, test.GenValAddress(), bondAmount)
				_, err := msgServer.BeginRedelegate(ctx, multiStakingMsg)
				return []sdk.Coin{}, err
			},
			expErr: true,
		},
		{
			name: "not allow token",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) ([]sdk.Coin, error) {
				bondAmount := sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(1000))

				multiStakingMsg := stakingtypes.NewMsgBeginRedelegate(delAddr, valAddr1, valAddr2, bondAmount)
				_, err := msgServer.BeginRedelegate(ctx, multiStakingMsg)
				return []sdk.Coin{}, err
			},
			expErr: true,
		},
		{
			name: "setup val3 with bond denom is arst then redelgate from val1 to val3",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) ([]sdk.Coin, error) {
				valPubKey3 := test.GenPubKey()
				bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				valAddr3 := sdk.ValAddress(valPubKey3.Address())
				createMsg := stakingtypes.MsgCreateValidator{Description: stakingtypes.Description{Moniker: "test", Identity: "test", Website: "test", SecurityContact: "test", Details: "test"}, Commission: stakingtypes.CommissionRates{Rate: sdk.MustNewDecFromStr("0.05"), MaxRate: sdk.MustNewDecFromStr("0.1"), MaxChangeRate: sdk.MustNewDecFromStr("0.1")}, MinSelfDelegation: sdk.NewInt(200), DelegatorAddress: delAddr.String(), ValidatorAddress: valAddr3.String(), Pubkey: codectypes.UnsafePackAny(valPubKey3), Value: sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(1000))}
				_, err := msgServer.CreateValidator(suite.ctx, &createMsg)
				suite.Require().NoError(err)

				multiStakingMsg := stakingtypes.NewMsgBeginRedelegate(delAddr, valAddr1, valAddr3, bondAmount)
				_, err = msgServer.BeginRedelegate(ctx, multiStakingMsg)
				return []sdk.Coin{}, err
			},
			expRate: []sdk.Dec{},
			expLock: []math.Int{},
			expErr:  true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			newParam := stakingtypes.DefaultParams()
			newParam.MinCommissionRate = sdk.MustNewDecFromStr("0.02")
			suite.app.StakingKeeper.SetParams(suite.ctx, newParam)
			suite.msKeeper.SetBondWeight(suite.ctx, MultiStakingDenomA, sdk.OneDec())
			suite.msKeeper.SetBondWeight(suite.ctx, MultiStakingDenomB, sdk.OneDec())

			bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
			userBalance := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(10000))
			suite.FundAccount(delAddr, sdk.NewCoins(userBalance, sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(10000))))
			suite.FundAccount(valDelAddr, sdk.NewCoins(userBalance, sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(10000))))

			createMsg := stakingtypes.MsgCreateValidator{
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
				MinSelfDelegation: sdk.NewInt(200),
				DelegatorAddress:  valDelAddr.String(),
				ValidatorAddress:  valAddr1.String(),
				Pubkey:            codectypes.UnsafePackAny(valPubKey1),
				Value:             bondAmount,
			}
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
				DelegatorAddress:  valDelAddr.String(),
				ValidatorAddress:  valAddr2.String(),
				Pubkey:            codectypes.UnsafePackAny(valPubKey2),
				Value:             bondAmount,
			}
			_, err := suite.msgServer.CreateValidator(suite.ctx, &createMsg)
			suite.Require().NoError(err)

			_, err = suite.msgServer.CreateValidator(suite.ctx, &createMsg2)
			suite.Require().NoError(err)

			suite.ctx = suite.ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})

			multiStakingAmounts, err := tc.malleate(suite.ctx, suite.msgServer, *suite.msKeeper)

			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				lockId1 := multistakingtypes.MultiStakingLockID(delAddr.String(), valAddr1.String())
				lockRecord1, found := suite.msKeeper.GetMultiStakingLock(suite.ctx, lockId1)
				suite.Require().True(found)
				suite.Require().Equal(tc.expRate[0], lockRecord1.GetBondWeight())
				suite.Require().Equal(tc.expLock[0], lockRecord1.LockedCoin.Amount)

				delegation1, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, delAddr, valAddr1)
				suite.Require().True(found)
				validator1, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr1)
				suite.Require().True(found)

				multiStakingCoin1 := multistakingtypes.NewMultiStakingCoin(multiStakingAmounts[0].Denom, multiStakingAmounts[0].Amount, tc.expRate[0])
				expShares1, err := validator1.SharesFromTokens(multiStakingCoin1.BondValue())
				suite.Require().NoError(err)
				suite.Require().Equal(expShares1, delegation1.GetShares())

				lockId2 := multistakingtypes.MultiStakingLockID(delAddr.String(), valAddr2.String())
				lockRecord2, found := suite.msKeeper.GetMultiStakingLock(suite.ctx, lockId2)
				suite.Require().True(found)
				suite.Require().Equal(tc.expRate[1], lockRecord2.GetBondWeight())
				suite.Require().Equal(tc.expLock[1], lockRecord2.LockedCoin.Amount)

				delegation2, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, delAddr, valAddr2)
				suite.Require().True(found)
				validator2, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr2)
				suite.Require().True(found)

				multiStakingCoin2 := multistakingtypes.NewMultiStakingCoin(multiStakingAmounts[1].Denom, multiStakingAmounts[1].Amount, tc.expRate[1])
				expShares2, err := validator2.SharesFromTokens(multiStakingCoin2.BondValue())
				suite.Require().NoError(err)
				suite.Require().Equal(expShares2, delegation2.GetShares())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUndelegate() {
	delAddr := test.GenAddress()
	valPubKey := test.GenPubKey()
	valAddr := sdk.ValAddress(valPubKey.Address())

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error
		expUnlock math.Int
		expLock   math.Int
		expErr    bool
	}{
		{
			name: "undelegate success",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error {
				undelegateAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				multiStakingMsg := stakingtypes.NewMsgUndelegate(delAddr, valAddr, undelegateAmount)
				_, err := msgServer.Undelegate(ctx, multiStakingMsg)
				return err
			},
			expUnlock: sdk.NewInt(500),
			expLock:   sdk.NewInt(500),
			expErr:    false,
		},
		{
			name: "undelegate 250 then undelegate 500",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error {
				undelegateAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(250))
				multiStakingMsg := stakingtypes.NewMsgUndelegate(delAddr, valAddr, undelegateAmount)
				_, err := msgServer.Undelegate(ctx, multiStakingMsg)
				if err != nil {
					return err
				}
				undelegateAmount1 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				multiStakingMsg1 := stakingtypes.NewMsgUndelegate(delAddr, valAddr, undelegateAmount1)
				_, err = msgServer.Undelegate(ctx, multiStakingMsg1)
				return err
			},
			expUnlock: sdk.NewInt(750),
			expLock:   sdk.NewInt(250),
			expErr:    false,
		},
		{
			name: "not found validator",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error {
				undelegateAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				multiStakingMsg := stakingtypes.NewMsgUndelegate(delAddr, test.GenValAddress(), undelegateAmount)
				_, err := msgServer.Undelegate(ctx, multiStakingMsg)
				return err
			},
			expErr: true,
		},
		{
			name: "not allow token",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error {
				undelegateAmount := sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(1000))

				multiStakingMsg := stakingtypes.NewMsgUndelegate(delAddr, test.GenValAddress(), undelegateAmount)
				_, err := msgServer.Undelegate(ctx, multiStakingMsg)
				return err
			},
			expErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			newParam := stakingtypes.DefaultParams()
			newParam.MinCommissionRate = sdk.MustNewDecFromStr("0.02")
			suite.app.StakingKeeper.SetParams(suite.ctx, newParam)

			initialWeight := sdk.MustNewDecFromStr("0.5")
			suite.msKeeper.SetBondWeight(suite.ctx, MultiStakingDenomA, initialWeight)
			bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000))
			userBalance := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(10000))
			suite.FundAccount(delAddr, sdk.NewCoins(userBalance, sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(10000))))

			createMsg := stakingtypes.MsgCreateValidator{
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
				MinSelfDelegation: sdk.NewInt(200),
				DelegatorAddress:  delAddr.String(),
				ValidatorAddress:  valAddr.String(),
				Pubkey:            codectypes.UnsafePackAny(valPubKey),
				Value:             bondAmount,
			}

			_, err := suite.msgServer.CreateValidator(suite.ctx, &createMsg)
			suite.Require().NoError(err)

			suite.ctx = suite.ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
			curHeight := suite.ctx.BlockHeight()
			err = tc.malleate(suite.ctx, suite.msgServer, *suite.msKeeper)

			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				lockId := multistakingtypes.MultiStakingLockID(delAddr.String(), valAddr.String())
				lockRecord, found := suite.msKeeper.GetMultiStakingLock(suite.ctx, lockId)
				suite.Require().True(found)
				suite.Require().Equal(tc.expLock, lockRecord.LockedCoin.Amount)

				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().True(found)
				validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
				suite.Require().True(found)

				multiStakingCoin := multistakingtypes.NewMultiStakingCoin(MultiStakingDenomA, tc.expLock, initialWeight)
				expShares, err := validator.SharesFromTokens(multiStakingCoin.BondValue())
				suite.Require().NoError(err)
				suite.Require().Equal(expShares, delegation.GetShares())

				unlockID := multistakingtypes.MultiStakingUnlockID(delAddr.String(), valAddr.String())
				unbondRecord, found := suite.msKeeper.GetMultiStakingUnlock(suite.ctx, unlockID)
				suite.Require().True(found)
				suite.Require().Equal(tc.expUnlock, unbondRecord.Entries[0].UnlockingCoin.Amount)

				ubd, found := suite.app.StakingKeeper.GetUnbondingDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().True(found)
				unlockStakingCoin := multistakingtypes.NewMultiStakingCoin(MultiStakingDenomA, tc.expUnlock, initialWeight)
				totalUBDAmount := math.ZeroInt()

				for _, ubdEntry := range ubd.Entries {
					if ubdEntry.CreationHeight == curHeight {
						totalUBDAmount = totalUBDAmount.Add(ubdEntry.Balance)
					}
				}
				suite.Require().Equal(unlockStakingCoin.BondValue(), totalUBDAmount)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCancelUnbondingDelegation() {
	delAddr := test.GenAddress()
	valPubKey := test.GenPubKey()
	valAddr := sdk.ValAddress(valPubKey.Address())

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error
		expUnlock math.Int
		expLock   math.Int
		expErr    bool
	}{
		{
			name: "cancel unbonding success",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error {
				cancelAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				multiStakingMsg := stakingtypes.NewMsgCancelUnbondingDelegation(delAddr, valAddr, ctx.BlockHeight(), cancelAmount)
				_, err := msgServer.CancelUnbondingDelegation(ctx, multiStakingMsg)
				return err
			},
			expUnlock: sdk.NewInt(500),
			expLock:   sdk.NewInt(1500),
			expErr:    false,
		},
		{
			name: "cancel unbonding 250 then cancel unbonding 500",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error {
				cancelAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(250))
				multiStakingMsg := stakingtypes.NewMsgCancelUnbondingDelegation(delAddr, valAddr, ctx.BlockHeight(), cancelAmount)
				_, err := msgServer.CancelUnbondingDelegation(ctx, multiStakingMsg)
				if err != nil {
					return err
				}
				cancelAmount1 := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				multiStakingMsg1 := stakingtypes.NewMsgCancelUnbondingDelegation(delAddr, valAddr, ctx.BlockHeight(), cancelAmount1)
				_, err = msgServer.CancelUnbondingDelegation(ctx, multiStakingMsg1)
				return err
			},
			expUnlock: sdk.NewInt(250),
			expLock:   sdk.NewInt(1750),
			expErr:    false,
		},
		{
			name: "not found validator",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error {
				cancelAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				multiStakingMsg := stakingtypes.NewMsgCancelUnbondingDelegation(delAddr, test.GenValAddress(), ctx.BlockHeight(), cancelAmount)
				_, err := msgServer.CancelUnbondingDelegation(ctx, multiStakingMsg)
				return err
			},
			expErr: true,
		},
		{
			name: "not allow token",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error {
				cancelAmount := sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(1000))

				multiStakingMsg := stakingtypes.NewMsgCancelUnbondingDelegation(delAddr, valAddr, ctx.BlockHeight(), cancelAmount)
				_, err := msgServer.CancelUnbondingDelegation(ctx, multiStakingMsg)
				return err
			},
			expErr: true,
		},
		{
			name: "not found entry at height 20",
			malleate: func(ctx sdk.Context, msgServer stakingtypes.MsgServer, msKeeper multistakingkeeper.Keeper) error {
				cancelAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(500))
				multiStakingMsg := stakingtypes.NewMsgCancelUnbondingDelegation(delAddr, valAddr, 20, cancelAmount)
				_, err := msgServer.CancelUnbondingDelegation(ctx, multiStakingMsg)
				return err
			},
			expErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			newParam := stakingtypes.DefaultParams()
			newParam.MinCommissionRate = sdk.MustNewDecFromStr("0.02")
			suite.app.StakingKeeper.SetParams(suite.ctx, newParam)

			initialWeight := sdk.MustNewDecFromStr("0.5")
			suite.msKeeper.SetBondWeight(suite.ctx, MultiStakingDenomA, initialWeight)
			bondAmount := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(2000))
			userBalance := sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(10000))
			suite.FundAccount(delAddr, sdk.NewCoins(userBalance, sdk.NewCoin(MultiStakingDenomB, sdk.NewInt(10000))))

			createMsg := stakingtypes.MsgCreateValidator{
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
				MinSelfDelegation: sdk.NewInt(200),
				DelegatorAddress:  delAddr.String(),
				ValidatorAddress:  valAddr.String(),
				Pubkey:            codectypes.UnsafePackAny(valPubKey),
				Value:             bondAmount,
			}

			_, err := suite.msgServer.CreateValidator(suite.ctx, &createMsg)
			suite.Require().NoError(err)

			suite.ctx = suite.ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
			curHeight := suite.ctx.BlockHeight()

			unbondMsg := stakingtypes.NewMsgUndelegate(delAddr, valAddr, sdk.NewCoin(MultiStakingDenomA, sdk.NewInt(1000)))
			_, err = suite.msgServer.Undelegate(suite.ctx, unbondMsg)
			suite.Require().NoError(err)

			err = tc.malleate(suite.ctx, suite.msgServer, *suite.msKeeper)

			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				lockId := multistakingtypes.MultiStakingLockID(delAddr.String(), valAddr.String())
				lockRecord, found := suite.msKeeper.GetMultiStakingLock(suite.ctx, lockId)
				suite.Require().True(found)
				suite.Require().Equal(tc.expLock, lockRecord.LockedCoin.Amount)

				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().True(found)
				validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
				suite.Require().True(found)

				multiStakingCoin := multistakingtypes.NewMultiStakingCoin(MultiStakingDenomA, tc.expLock, initialWeight)
				expShares, err := validator.SharesFromTokens(multiStakingCoin.BondValue())
				suite.Require().NoError(err)
				suite.Require().Equal(expShares, delegation.GetShares())

				unlockID := multistakingtypes.MultiStakingUnlockID(delAddr.String(), valAddr.String())
				unbondRecord, found := suite.msKeeper.GetMultiStakingUnlock(suite.ctx, unlockID)
				suite.Require().True(found)
				suite.Require().Equal(tc.expUnlock, unbondRecord.Entries[0].UnlockingCoin.Amount)

				ubd, found := suite.app.StakingKeeper.GetUnbondingDelegation(suite.ctx, delAddr, valAddr)
				suite.Require().True(found)
				unlockStakingCoin := multistakingtypes.NewMultiStakingCoin(MultiStakingDenomA, tc.expUnlock, initialWeight)
				totalUBDAmount := math.ZeroInt()

				for _, ubdEntry := range ubd.Entries {
					if ubdEntry.CreationHeight == curHeight {
						totalUBDAmount = totalUBDAmount.Add(ubdEntry.Balance)
					}
				}
				suite.Require().Equal(unlockStakingCoin.BondValue(), totalUBDAmount)
			}
		})
	}
}
