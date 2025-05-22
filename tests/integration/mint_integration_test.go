package integration

import (
	"math/big"

	"cosmossdk.io/math"
	commonfactory "github.com/cosmos/evm/testutil/integration/common/factory"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	realiotypes "github.com/realiotech/realio-network/types"
)

type endBlockTestCase struct {
	name        string
	preFundFunc func() error
	expBalances sdk.Coins
	shouldBurn  bool
}

var (
	testTokenDenom       = "test"
	EvmDeadAddr          = common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	sendAmount     int64 = 1000000
)

func (suite *EVMTestSuite) TestMintEndBlock() {
	senderKey := suite.keyring.GetKey(0)
	testCases := []endBlockTestCase{
		{
			name: "empty balance",
			preFundFunc: func() error {
				return nil
			},
			shouldBurn: false,
		},
		{
			name: "only RIO locked",
			preFundFunc: func() error {
				txArgs := evmtypes.EvmTxArgs{
					To:     &EvmDeadAddr,
					Amount: big.NewInt(sendAmount),
				}
				res, err := suite.factory.ExecuteEthTx(senderKey.Priv, txArgs)
				suite.Require().NoError(err)
				suite.Require().True(res.IsOK(), "transaction should have succeeded: %s", res.GetLog())

				return nil
			},
			shouldBurn:  true,
			expBalances: nil,
		},
		{
			name: "Have token locked but not RIO",
			preFundFunc: func() error {
				res, err := suite.factory.ExecuteCosmosTx(senderKey.Priv, commonfactory.CosmosTxArgs{
					Msgs: []sdk.Msg{&banktypes.MsgSend{
						FromAddress: senderKey.AccAddr.String(),
						ToAddress:   sdk.AccAddress(EvmDeadAddr.Bytes()).String(),
						Amount:      sdk.NewCoins(sdk.NewCoin("test", math.NewInt(sendAmount))),
					}},
				})
				suite.Require().NoError(err)
				suite.Require().True(res.IsOK(), "transaction should have succeeded: %s", res.GetLog())
				return nil
			},
			shouldBurn:  false,
			expBalances: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(sendAmount))),
		},
		{
			name: "Multiple coins have RIO",
			preFundFunc: func() error {
				res, err := suite.factory.ExecuteCosmosTx(senderKey.Priv, commonfactory.CosmosTxArgs{
					Msgs: []sdk.Msg{&banktypes.MsgSend{
						FromAddress: senderKey.AccAddr.String(),
						ToAddress:   sdk.AccAddress(EvmDeadAddr.Bytes()).String(),
						Amount:      sdk.NewCoins(sdk.NewCoin("test", math.NewInt(sendAmount)), sdk.NewCoin(realiotypes.BaseDenom, math.NewInt(sendAmount))),
					}},
				})
				suite.Require().NoError(err)
				suite.Require().True(res.IsOK(), "transaction should have succeeded: %s", res.GetLog())
				return nil
			},
			shouldBurn:  true,
			expBalances: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(sendAmount*2))),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.preFundFunc()
			suite.Require().NoError(err)

			err = suite.network.NextBlock()
			suite.Require().NoError(err)

			balances, err := suite.grpcHandler.GetAllBalances(EvmDeadAddr.Bytes())
			suite.Require().NoError(err)
			suite.Require().Equal(balances.Balances, tc.expBalances)
		})
	}
}
