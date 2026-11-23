package integration

import (
	"math/big"

	"cosmossdk.io/math"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/evm/contracts"
	commonfactory "github.com/cosmos/evm/testutil/integration/base/factory"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
)

var (
	transferAmount   int64 = 1000
	compiledContract       = contracts.ERC20MinterBurnerDecimalsContract
	senderIndex            = 0
	recipientIndex         = 1
	constructorArgs        = []interface{}{"coin", "token", uint8(18)}
)

func (suite *EVMTestSuite) TestERC20MsgsTxBuilder() {
	// Deploy ERC20 contract
	senderPriv := suite.keyring.GetPrivKey(senderIndex)
	recipientPriv := suite.keyring.GetPrivKey(recipientIndex)

	recipientKey := suite.keyring.GetKey(recipientIndex)
	senderKey := suite.keyring.GetKey(senderIndex)

	contractAddr, err := suite.factory.DeployContract(
		senderPriv,
		evmtypes.EvmTxArgs{}, // Default values
		testutiltypes.ContractDeploymentData{
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
	mintArgs := testutiltypes.CallArgs{
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

	// Register ERC20 token as native token: erc20 msgs are exposed to users,
	// so registration must now succeed.
	coinDenom := suite.registerErc20(contractAddr)

	// Converting ERC20 to the native coin is rejected: MsgConvertERC20 is not
	// in the app codec.
	_, err = suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&erc20types.MsgConvertERC20{
			ContractAddress: contractAddr.Hex(),
			Amount:          math.NewInt(transferAmount),
			Receiver:        recipientKey.AccAddr.String(),
			Sender:          senderKey.Addr.Hex(),
		}},
	})
	suite.Require().ErrorContains(err, "unable to resolve type URL /cosmos.evm.erc20.v1.MsgConvertERC20")
	suite.Require().NoError(suite.network.NextBlock())
	suite.assertContractBalanceOf(contractAddr, senderKey.Addr, mintAmount)
	suite.assertBankBalance(recipientKey.AccAddr, coinDenom, 0)

	// MsgConvertCoin is in the app codec, so it reaches x/erc20, which fails
	// to escrow the coin because the recipient holds none.
	_, err = suite.factory.ExecuteCosmosTx(recipientPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&erc20types.MsgConvertCoin{
			Coin:     sdk.NewCoin(coinDenom, math.NewInt(transferAmount)),
			Sender:   recipientKey.AccAddr.String(),
			Receiver: senderKey.Addr.Hex(),
		}},
	})
	suite.Require().ErrorContains(err, "failed to escrow coins")
	suite.Require().NoError(suite.network.NextBlock())
	suite.assertContractBalanceOf(contractAddr, senderKey.Addr, mintAmount)
}

func (suite *EVMTestSuite) assertContractTotalSupply(contractAddr common.Address, expected int64) {
	totalSupplyTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	totalSupplyArgs := testutiltypes.CallArgs{
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
	balanceArgs := testutiltypes.CallArgs{
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

// registerErc20 registers contractAddr as a native token pair and returns
// the resulting bank denom ("erc20:0x...").
func (suite *EVMTestSuite) registerErc20(contractAddr common.Address) string {
	_, err := suite.factory.ExecuteCosmosTx(suite.keyring.GetPrivKey(senderIndex), commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&erc20types.MsgRegisterERC20{
			Signer:         suite.keyring.GetAccAddr(0).String(),
			Erc20Addresses: []string{contractAddr.Hex()},
		}},
	})
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())
	return erc20types.CreateDenom(contractAddr.Hex())
}

// assertBankBalance asserts addr's bank balance of denom equals expected.
func (suite *EVMTestSuite) assertBankBalance(addr sdk.AccAddress, denom string, expected int64) {
	res, err := suite.network.GetBankClient().Balance(suite.network.GetContext(), &banktypes.QueryBalanceRequest{
		Address: addr.String(),
		Denom:   denom,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(expected), res.Balance.Amount)
}
