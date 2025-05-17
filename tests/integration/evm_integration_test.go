package integration

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/evm/testutil/integration/os/factory"
	"github.com/cosmos/evm/testutil/integration/os/grpc"
	testkeyring "github.com/cosmos/evm/testutil/integration/os/keyring"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/realiotech/realio-network/testutil/integration/network"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EVMTestSuite struct {
	suite.Suite
	network     network.Network
	grpcHandler grpc.Handler
	factory     factory.TxFactory
	keyring     testkeyring.Keyring
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EVMTestSuite) SetupTest() {
	keyring := testkeyring.New(4)
	integrationNetwork := network.New(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	factory := factory.New(integrationNetwork, grpcHandler)

	suite.grpcHandler = grpcHandler
	suite.factory = factory
	suite.network = integrationNetwork
	suite.keyring = keyring
}

type transferTestCase struct {
	name      string
	getTxArgs func() evmtypes.EvmTxArgs
}

func (suite *EVMTestSuite) TestNativeTransfers() {
	testCases := []transferTestCase{
		{
			name: "DynamicFeeTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
		},
		{
			name: "AccessListTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					Accesses: &ethtypes.AccessList{{
						Address:     suite.keyring.GetKey(1).Addr,
						StorageKeys: []common.Hash{{0}},
					}},
				}
			},
		},
		{
			name: "LegacyTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					GasPrice: big.NewInt(1e9),
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			senderKey := suite.keyring.GetKey(0)
			receiverKey := suite.keyring.GetKey(1)
			denom := suite.network.GetBaseDenom()

			txArgs := tc.getTxArgs()

			senderPrevBalanceResponse, err := suite.grpcHandler.GetBalanceFromBank(senderKey.AccAddr, denom)
			suite.NoError(err, "Error should not exists")

			senderPrevBalance := senderPrevBalanceResponse.GetBalance().Amount

			receiverPrevBalanceResponse, err := suite.grpcHandler.GetBalanceFromBank(receiverKey.AccAddr, denom)
			suite.Require().NoError(err)
			receiverPrevBalance := receiverPrevBalanceResponse.GetBalance().Amount

			transferAmount := int64(1000)

			txArgs.Amount = big.NewInt(transferAmount)
			txArgs.To = &receiverKey.Addr

			res, err := suite.factory.ExecuteEthTx(senderKey.Priv, txArgs)
			suite.Require().NoError(err)
			suite.Require().True(res.IsOK(), "transaction should have succeeded: %s", res.GetLog())

			err = suite.network.NextBlock()
			suite.Require().NoError(err)

			// Check sender balance after transaction
			senderBalanceResultBeforeFees := senderPrevBalance.Sub(math.NewInt(transferAmount))
			senderAfterBalance, err := suite.grpcHandler.GetBalanceFromBank(senderKey.AccAddr, denom)
			suite.Require().NoError(err)
			suite.Require().True(senderAfterBalance.GetBalance().Amount.LTE(senderBalanceResultBeforeFees))

			// Check receiver balance after transaction
			receiverBalanceResult := receiverPrevBalance.Add(math.NewInt(transferAmount))
			receverAfterBalanceResponse, err := suite.grpcHandler.GetBalanceFromBank(receiverKey.AccAddr, denom)
			suite.Require().NoError(err)
			suite.Require().Equal(receiverBalanceResult, receverAfterBalanceResponse.GetBalance().Amount)
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEVMTestSuite(t *testing.T) {
	suite.Run(t, new(EVMTestSuite))
}
