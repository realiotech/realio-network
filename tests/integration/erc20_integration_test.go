package integration

import (
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/evm/testutil/integration/os/factory"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/contracts"
	commonfactory "github.com/cosmos/evm/testutil/integration/common/factory"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
)

var (
	mintAmount       int64 = 1_000_000
	transferAmount   int64 = 1000
	compiledContract       = contracts.ERC20MinterBurnerDecimalsContract
	senderIndex            = 0
	recipientIndex         = 1
	constructorArgs        = []interface{}{"coin", "token", uint8(18)}
)

func (suite *EVMTestSuite) TestERC20RegisterAndConverting() {
	// Deploy ERC20 contract
	senderPriv := suite.keyring.GetPrivKey(senderIndex)
	recipientPriv := suite.keyring.GetPrivKey(recipientIndex)

	recipientKey := suite.keyring.GetKey(recipientIndex)
	senderKey := suite.keyring.GetKey(senderIndex)

	contractAddr, err := suite.factory.DeployContract(
		senderPriv,
		evmtypes.EvmTxArgs{}, // Default values
		factory.ContractDeploymentData{
			Contract:        compiledContract,
			ConstructorArgs: constructorArgs,
		},
	)
	suite.Require().NoError(err)
	suite.NotEqual(contractAddr, common.Address{})
	suite.Require().NoError(suite.network.NextBlock())

	// Mint token to sender
	mintTxArgs := evmtypes.EvmTxArgs{}
	mintTxArgs.To = &contractAddr
	amountToMint := big.NewInt(mintAmount)
	mintArgs := factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{suite.keyring.GetKey(0).Addr, amountToMint},
	}
	mintResponse, err := suite.factory.ExecuteContractCall(senderPriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintResponse.IsOK(), "transaction should have succeeded", mintResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
	suite.assertContractTotalSupply(contractAddr, mintAmount)
	suite.assertContractBalanceOf(contractAddr, senderKey.Addr, mintAmount)

	// Register ERC20 token as native token
	contractNativeDenom := suite.registerErc20(contractAddr)

	// Convert ERC20 to native token
	res, err := suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&erc20types.MsgConvertERC20{
			ContractAddress: contractAddr.Hex(),
			Amount:          math.NewInt(transferAmount),
			Receiver:        recipientKey.AccAddr.String(),
			Sender:          senderKey.Addr.Hex(),
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "transaction should have succeeded", mintResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
	suite.assertBankBalance(recipientKey.AccAddr, contractNativeDenom, transferAmount)
	suite.assertContractBalanceOf(contractAddr, senderKey.Addr, mintAmount-transferAmount)
	suite.assertContractTotalSupply(contractAddr, mintAmount)

	// Convert native token back to ERC20
	res, err = suite.factory.ExecuteCosmosTx(recipientPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&erc20types.MsgConvertCoin{
			Coin:     sdk.NewCoin(contractNativeDenom, math.NewInt(transferAmount)),
			Sender:   recipientKey.AccAddr.String(),
			Receiver: senderKey.Addr.Hex(),
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "transaction should have succeeded", mintResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
	suite.assertBankBalance(recipientKey.AccAddr, contractNativeDenom, 0)
	suite.assertContractBalanceOf(contractAddr, senderKey.Addr, mintAmount)
	suite.assertContractTotalSupply(contractAddr, mintAmount)
}

func (suite *EVMTestSuite) assertContractTotalSupply(contractAddr common.Address, expected int64) {
	totalSupplyTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	totalSupplyArgs := factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "totalSupply",
		Args:        []interface{}{},
	}
	totalSupplyRes, err := suite.factory.ExecuteContractCall(suite.keyring.GetPrivKey(senderIndex), totalSupplyTxArgs, totalSupplyArgs)
	suite.Require().NoError(err)
	suite.Require().True(totalSupplyRes.IsOK(), "transaction should have succeeded", totalSupplyRes.GetLog())

	var totalSupplyResponse *big.Int
	err = integrationutils.DecodeContractCallResponse(&totalSupplyResponse, totalSupplyArgs, totalSupplyRes)
	suite.Require().NoError(err)
	suite.Require().Equal(totalSupplyResponse, big.NewInt(expected))
	suite.Require().NoError(suite.network.NextBlock())
}

func (suite *EVMTestSuite) assertContractBalanceOf(contractAddr common.Address, addr common.Address, expected int64) {
	balanceTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	balanceArgs := factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "balanceOf",
		Args:        []interface{}{addr},
	}
	balanceRes, err := suite.factory.ExecuteContractCall(suite.keyring.GetPrivKey(senderIndex), balanceTxArgs, balanceArgs)
	suite.Require().NoError(err)

	var balance *big.Int
	err = integrationutils.DecodeContractCallResponse(&balance, balanceArgs, balanceRes)
	suite.Require().NoError(err)
	suite.Require().Equal(balance, big.NewInt(expected))

	suite.Require().NoError(suite.network.NextBlock())
}

func (suite *EVMTestSuite) assertBankBalance(addr []byte, denom string, expected int64) {
	balance, err := suite.grpcHandler.GetBalanceFromBank(addr, denom)
	suite.Require().NoError(err)
	suite.Require().Equal(balance.Balance.Amount, math.NewInt(expected))
}

func (suite *EVMTestSuite) registerErc20(contractAddr common.Address) string {
	res, err := suite.factory.ExecuteCosmosTx(suite.keyring.GetPrivKey(senderIndex), commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&erc20types.MsgRegisterERC20{
			Signer:         suite.keyring.GetAccAddr(0).String(),
			Erc20Addresses: []string{contractAddr.Hex()},
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "transaction should have succeeded", res.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Query native denom of contract
	erc20Client := suite.network.GetERC20Client()
	tokenPairRes, err := erc20Client.TokenPair(suite.network.GetContext(), &erc20types.QueryTokenPairRequest{
		Token: contractAddr.Hex(),
	})
	suite.Require().NoError(err)
	contractDenom := tokenPairRes.TokenPair.Denom
	return contractDenom
}
