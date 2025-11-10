package integration

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	commonfactory "github.com/cosmos/evm/testutil/integration/base/factory"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	realiotypes "github.com/realiotech/realio-network/types"

	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	"github.com/realiotech/realio-network/testutil/integration/network"

	// authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/realiotech/realio-network/x/mint/types"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type MintTestSuite struct {
	suite.Suite
	network     network.Network
	grpcHandler grpc.Handler
	factory     factory.TxFactory
	keyring     testkeyring.Keyring
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *MintTestSuite) SetupTest() {
	keyring := testkeyring.New(4)
	integrationNetwork := network.New(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithOtherDenoms([]string{testTokenDenom}),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	factory := factory.New(integrationNetwork, grpcHandler)

	suite.grpcHandler = grpcHandler
	suite.factory = factory
	suite.network = integrationNetwork
	suite.keyring = keyring
}

type endBlockTestCase struct {
	name        string
	preFundFunc func() error
	expBalances sdk.Coins
	shouldBurn  bool
}

var (
	testTokenDenom        = "test"
	EvmDeadAddr           = common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	sendAmount      int64 = 4000000000000000000
	rioSupplyCap, _       = math.NewIntFromString("175000000000000000000000000")
)

func (suite *MintTestSuite) TestMintEndBlock() {
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
			totalSupplyBefore, err := suite.grpcHandler.GetTotalSupply()
			suite.Require().NoError(err)

			err = tc.preFundFunc()
			suite.Require().NoError(err)

			// Recalculate minted amount in BeginBlocker
			mintClient := suite.network.GetMintModuleClient()
			params, err := mintClient.Params(suite.network.GetContext(), &types.QueryParamsRequest{})
			suite.Require().NoError(err)
			annualProvisions := params.Params.InflationRate.MulInt(rioSupplyCap.Sub(totalSupplyBefore.Supply.AmountOf(realiotypes.BaseDenom)))
			mintedAmount := annualProvisions.QuoInt(math.NewIntFromUint64(params.Params.BlocksPerYear)).TruncateInt()

			balances, err := suite.grpcHandler.GetAllBalances(EvmDeadAddr.Bytes())
			suite.Require().NoError(err)
			suite.Require().Equal(balances.Balances, tc.expBalances)

			totalSupplyAfter, err := suite.grpcHandler.GetTotalSupply()
			suite.Require().NoError(err)

			// supplyAfter = supplyBefore + mint - burn
			if tc.shouldBurn {
				burnedAmount := sendAmount
				expectedSupply := totalSupplyBefore.Supply.AmountOf(realiotypes.BaseDenom).Add(mintedAmount).Sub(math.NewInt(burnedAmount))
				actualSupply := totalSupplyAfter.Supply.AmountOf(realiotypes.BaseDenom)
				suite.Require().Equal(expectedSupply, actualSupply)
			}

			suite.Require().NoError(suite.network.NextBlock())
		})
	}
}

func TestMintestSuite(t *testing.T) {
	suite.Run(t, new(MintTestSuite))
}
