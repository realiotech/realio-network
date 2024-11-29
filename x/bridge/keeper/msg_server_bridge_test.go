package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/testutil"
	"github.com/realiotech/realio-network/x/bridge/keeper"
	"github.com/realiotech/realio-network/x/bridge/types"
)

func (suite *KeeperTestSuite) TestBridgeIn() {
	testAccAddress := testutil.GenAddress().String()
	testCases := []struct {
		name         string
		msg          *types.MsgBridgeIn
		setAuthority bool
		expectErr    bool
		errString    string
	}{
		{
			name: "valid MsgBridgeIn",
			msg: &types.MsgBridgeIn{
				Authority: "",
				Coin:      sdk.NewInt64Coin("ario", 1000000),
			},
			setAuthority: true,
			expectErr:    false,
		},
		{
			name: "invalid MsgBridgeIn; coin not in register list",
			msg: &types.MsgBridgeIn{
				Authority: "",
				Coin:      sdk.NewInt64Coin("eth", 1000000),
			},
			setAuthority: true,
			expectErr:    true,
			errString:    "denom: eth: coin not in register list",
		},
		{
			name: "invalid MsgBridgeIn; unauthorized",
			msg: &types.MsgBridgeIn{
				Authority: testAccAddress,
				Coin:      sdk.NewInt64Coin("ario", 1000000),
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
				tc.msg.Reciever = suite.admin
			}

			addr, err := suite.app.AccountKeeper.AddressCodec().StringToBytes(tc.msg.Authority)
			suite.Require().NoError(err)
			prevBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, tc.msg.Coin.Denom)

			_, err = srv.BridgeIn(suite.ctx, tc.msg)
			if tc.expectErr {
				suite.Require().ErrorContains(err, tc.errString)
			} else {
				suite.Require().NoError(err)

				afterBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, tc.msg.Coin.Denom)
				suite.Require().Equal(prevBalance.Add(tc.msg.Coin), afterBalance)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBridgeOut() {
	testAcc := testutil.GenAddress()
	testAccAddress := testAcc.String()
	prevBalance := sdk.NewInt64Coin("ario", 2000000)
	testCases := []struct {
		name      string
		msg       *types.MsgBridgeOut
		expectErr bool
		errString string
	}{
		{
			name: "valid MsgBridgeOut",
			msg: &types.MsgBridgeOut{
				Signer: testAccAddress,
				Coin:   sdk.NewInt64Coin("ario", 1000000),
			},
			expectErr: false,
		},
		{
			name: "invalid MsgBridgeOut; coin not in register list",
			msg: &types.MsgBridgeOut{
				Signer: testAccAddress,
				Coin:   sdk.NewInt64Coin("eth", 1000000),
			},
			expectErr: true,
			errString: "denom: eth: coin not in register list",
		},
		{
			name: "invalid MsgBridgeOut; insufficient funds",
			msg: &types.MsgBridgeOut{
				Signer: testAccAddress,
				Coin:   sdk.NewInt64Coin("arst", 1000000),
			},
			expectErr: true,
			errString: "spendable balance 0arst is smaller than 1000000arst: insufficient funds",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			srv := keeper.NewMsgServerImpl(suite.app.BridgeKeeper)

			adminAcc, err := suite.app.AccountKeeper.AddressCodec().StringToBytes(suite.admin)
			suite.Require().NoError(err)
			err = suite.app.BankKeeper.SendCoins(suite.ctx, adminAcc, testAcc, sdk.NewCoins(prevBalance))
			suite.Require().NoError(err)

			_, err = srv.BridgeOut(suite.ctx, tc.msg)
			if tc.expectErr {
				suite.Require().EqualError(err, tc.errString)
			} else {
				suite.Require().NoError(err)

				newBalance := suite.app.BankKeeper.GetBalance(suite.ctx, testAcc, tc.msg.Coin.Denom)
				suite.Require().Equal(prevBalance.Sub(tc.msg.Coin), newBalance)
			}
		})
	}
}
