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
	mintAmount     int64 = 1_000_000
	transferAmount int64 = 1000
)

func (suite *EVMTestSuite) TestERC20RegisterAndConverting() {
	// Deploy ERC20 contract
	senderPriv := suite.keyring.GetPrivKey(0)
	recipientPriv := suite.keyring.GetPrivKey(1)
	constructorArgs := []interface{}{"coin", "token", uint8(18)}
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract
	recipientKey := suite.keyring.GetKey(1)
	senderKey := suite.keyring.GetKey(0)

	var err error // Avoid shadowing
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

	err = suite.network.NextBlock()
	suite.Require().NoError(err)

	// Mint first
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

	err = suite.network.NextBlock()
	suite.Require().NoError(err)

	totalSupplyTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	totalSupplyArgs := factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "totalSupply",
		Args:        []interface{}{},
	}
	totalSupplyRes, err := suite.factory.ExecuteContractCall(senderPriv, totalSupplyTxArgs, totalSupplyArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintResponse.IsOK(), "transaction should have succeeded", totalSupplyRes.GetLog())

	var totalSupplyResponse *big.Int
	err = integrationutils.DecodeContractCallResponse(&totalSupplyResponse, totalSupplyArgs, totalSupplyRes)
	suite.Require().NoError(err)
	suite.Require().Equal(totalSupplyResponse, amountToMint)
	suite.Require().NoError(suite.network.NextBlock())

	balanceTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	balanceArgs := factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "balanceOf",
		Args:        []interface{}{senderKey.Addr},
	}
	senderEvmBalanceRes, err := suite.factory.ExecuteContractCall(senderPriv, balanceTxArgs, balanceArgs)
	suite.Require().NoError(err)

	var initSenderEvmBalance *big.Int
	err = integrationutils.DecodeContractCallResponse(&initSenderEvmBalance, balanceArgs, senderEvmBalanceRes)
	suite.Require().NoError(err)
	suite.Require().Equal(initSenderEvmBalance, big.NewInt(mintAmount))

	suite.Require().NoError(suite.network.NextBlock())

	// Register ERC20 token as native token
	res, err := suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&erc20types.MsgRegisterERC20{
			Signer:         suite.keyring.GetAccAddr(0).String(),
			Erc20Addresses: []string{contractAddr.Hex()},
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "transaction should have succeeded", mintResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Query native denom of contract
	erc20Client := suite.network.GetERC20Client()
	tokenPairRes, err := erc20Client.TokenPair(suite.network.GetContext(), &erc20types.QueryTokenPairRequest{
		Token: contractAddr.Hex(),
	})
	suite.Require().NoError(err)
	contractDenom := tokenPairRes.TokenPair.Denom

	// Convert ERC20 to native token
	res, err = suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
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

	recipientBalance, err := suite.grpcHandler.GetBalanceFromBank(recipientKey.AccAddr, contractDenom)
	suite.Require().NoError(err)
	suite.Require().Equal(recipientBalance.Balance.Amount, math.NewInt(transferAmount))

	balanceTxArgs = evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	balanceArgs = factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "balanceOf",
		Args:        []interface{}{senderKey.Addr},
	}
	senderEvmBalanceRes, err = suite.factory.ExecuteContractCall(senderPriv, balanceTxArgs, balanceArgs)
	suite.Require().NoError(err)

	var senderEvmBalance *big.Int
	err = integrationutils.DecodeContractCallResponse(&senderEvmBalance, balanceArgs, senderEvmBalanceRes)
	suite.Require().NoError(err)
	suite.Require().Equal(senderEvmBalance, big.NewInt(mintAmount-transferAmount))

	suite.Require().NoError(suite.network.NextBlock())

	// Convert native token back to ERC20
	res, err = suite.factory.ExecuteCosmosTx(recipientPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&erc20types.MsgConvertCoin{
			Coin:     sdk.NewCoin(contractDenom, math.NewInt(transferAmount)),
			Sender:   recipientKey.AccAddr.String(),
			Receiver: senderKey.Addr.Hex(),
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "transaction should have succeeded", mintResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	recipientBalance, err = suite.grpcHandler.GetBalanceFromBank(recipientKey.AccAddr, contractDenom)
	suite.Require().NoError(err)
	suite.Require().Equal(recipientBalance.Balance.Amount, math.ZeroInt())

	balanceTxArgs = evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	balanceArgs = factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "balanceOf",
		Args:        []interface{}{senderKey.Addr},
	}
	senderEvmBalanceRes, err = suite.factory.ExecuteContractCall(senderPriv, balanceTxArgs, balanceArgs)
	suite.Require().NoError(err)

	var senderEvmBalanceAfter *big.Int
	err = integrationutils.DecodeContractCallResponse(&senderEvmBalanceAfter, balanceArgs, senderEvmBalanceRes)
	suite.Require().NoError(err)
	suite.Require().Equal(senderEvmBalanceAfter, big.NewInt(mintAmount))

	suite.Require().NoError(suite.network.NextBlock())
}
