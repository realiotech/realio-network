package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/testutil"
	"github.com/realiotech/realio-network/x/bridge/keeper"
	"github.com/realiotech/realio-network/x/bridge/types"
)

func (suite *KeeperTestSuite) TestRegisterNewCoins() {
	testAccAddress := testutil.GenAddress().String()
	testCases := []struct {
		name         string
		msg          types.MsgRegisterNewCoins
		setAuthority bool
		expectErr    bool
		errString    string
	}{
		{
			name: "valid MsgRegisterNewCoins",
			msg: types.MsgRegisterNewCoins{
				Authority: "",
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin("eth", 1000000),
				),
			},
			setAuthority: true,
			expectErr:    false,
		},
		{
			name: "invalid MsgRegisterNewCoins; duplicated denom ario",
			msg: types.MsgRegisterNewCoins{
				Authority: "",
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin("ario", 1000000),
				),
			},
			setAuthority: true,
			expectErr:    true,
			errString:    "denom: ario: coin already in register list",
		},
		{
			name: "invalid MsgRegisterNewCoins; duplicated denom bar",
			msg: types.MsgRegisterNewCoins{
				Authority: "",
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin("bar", 1000000),
				),
			},
			setAuthority: true,
			expectErr:    true,
			errString:    "denom: bar: coin already in register list",
		},
		{
			name: "invalid MsgRegisterNewCoins; unauthorized",
			msg: types.MsgRegisterNewCoins{
				Authority: testAccAddress,
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin("eth", 1000000),
				),
			},
			setAuthority: false,
			expectErr:    true,
			errString:    "invalid authority",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			srv := keeper.NewMsgServerImpl(suite.app.BridgeKeeper)

			// we register a denom "bar" here
			// there's a test case to make sure we CANNOT register new coin with the same symbol
			_, err := srv.RegisterNewCoins(suite.ctx, &types.MsgRegisterNewCoins{
				Authority: suite.admin,
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin("bar", 9999999),
				),
			})
			suite.Require().NoError(err)

			if tc.setAuthority {
				tc.msg.Authority = suite.admin
			}

			_, err = srv.RegisterNewCoins(suite.ctx, &tc.msg)
			if tc.expectErr {
				suite.Require().ErrorContains(err, tc.errString)
			} else {
				suite.Require().NoError(err)

				for _, coin := range tc.msg.Coins {
					registeredCoin, err := suite.app.BridgeKeeper.RegisteredCoins.Get(suite.ctx,
						coin.Denom,
					)
					suite.Require().NoError(err)
					suite.Require().Equal(registeredCoin.CurrentInflow, math.ZeroInt())
					suite.Require().Equal(registeredCoin.Ratelimit, coin.Amount)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDeregisterCoins() {
	testAccAddress := testutil.GenAddress().String()
	testCases := []struct {
		name         string
		msg          types.MsgDeregisterCoins
		setAuthority bool
		expectErr    bool
		errString    string
	}{
		{
			name: "valid MsgDeregisterCoins",
			msg: types.MsgDeregisterCoins{
				Authority: "",
				Denoms:    []string{"ario"},
			},
			setAuthority: true,
			expectErr:    false,
		},
		{
			name: "invalid MsgDeregisterCoins; coin not in register list",
			msg: types.MsgDeregisterCoins{
				Authority: "",
				Denoms:    []string{"eth"},
			},
			setAuthority: true,
			expectErr:    true,
			errString:    "denom: eth: coin not in register list",
		},
		{
			name: "invalid MsgDeregisterCoins; unauthorized",
			msg: types.MsgDeregisterCoins{
				Authority: testAccAddress,
				Denoms:    []string{"ario"},
			},
			setAuthority: false,
			expectErr:    true,
			errString:    "invalid authority",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			srv := keeper.NewMsgServerImpl(suite.app.BridgeKeeper)

			if tc.setAuthority {
				tc.msg.Authority = suite.admin
			}

			_, err := srv.DeregisterCoins(suite.ctx, &tc.msg)
			if tc.expectErr {
				suite.Require().ErrorContains(err, tc.errString)
			} else {
				suite.Require().NoError(err)

				for _, denom := range tc.msg.Denoms {
					_, err := suite.app.BridgeKeeper.RegisteredCoins.Get(suite.ctx,
						denom,
					)
					suite.Require().Error(err)
				}
			}
		})
	}
}
