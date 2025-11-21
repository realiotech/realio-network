package integration

import (
	"cosmossdk.io/math"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/precompiles/testutil"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	precompileBank "github.com/realiotech/realio-network/precompile/bank"
)

var (
	bankPrecompileAddr = common.HexToAddress(evmtypes.BankPrecompileAddress)
	// Load bank precompile ABI
	abi = precompileBank.ABI
)

func (suite *EVMTestSuite) TestBankPrecompileSend() {
	// Get test accounts
	senderPriv := suite.keyring.GetPrivKey(0)
	senderKey := suite.keyring.GetKey(0)
	recipientKey := suite.keyring.GetKey(1)

	senderInitialBalance := suite.GetBalancesByBankPrecompile(senderPriv, senderKey.Addr)
	recipientInitialBalance := suite.GetBalancesByBankPrecompile(senderPriv, recipientKey.Addr)

	// Execute transfer through bank precompile
	transferAmt := math.NewInt(transferAmount)
	res, err := suite.factory.ExecuteContractCall(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &bankPrecompileAddr,
		},
		testutiltypes.CallArgs{
			ContractABI: abi,
			MethodName:  "send",
			Args: []interface{}{
				recipientKey.Addr,
				sdk.NewCoins(sdk.NewCoin("ario", transferAmt)).String(),
			},
		},
	)
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "bank transfer should have succeeded", res.GetLog())

	// Move to next block
	err = suite.network.NextBlock()
	suite.Require().NoError(err)

	// Verify balances after transfer

	senderFinalBalance := suite.GetBalancesByBankPrecompile(senderPriv, senderKey.Addr)
	recipientFinalBalance := suite.GetBalancesByBankPrecompile(senderPriv, recipientKey.Addr)

	suite.Require().True(recipientFinalBalance.AmountOf("ario").Equal(recipientInitialBalance.AmountOf("ario").Add(transferAmt)))
	// sender balance less cause of gas
	suite.Require().True(senderFinalBalance.AmountOf("ario").LT(senderInitialBalance.AmountOf("ario").Sub(transferAmt)))
}

func (suite *EVMTestSuite) TestBankPrecompileMultiSend() {
	// Get test accounts
	senderPriv := suite.keyring.GetPrivKey(0)
	senderKey := suite.keyring.GetKey(0)
	recipient1Key := suite.keyring.GetKey(1)
	recipient2Key := suite.keyring.GetKey(2)

	senderInitialBalance := suite.GetBalancesByBankPrecompile(senderPriv, senderKey.Addr)
	recipient1InitialBalance := suite.GetBalancesByBankPrecompile(senderPriv, recipient1Key.Addr)
	recipient2InitialBalance := suite.GetBalancesByBankPrecompile(senderPriv, recipient2Key.Addr)

	// Execute transfer through bank precompile
	transferAmt := math.NewInt(transferAmount)
	transferAmt1 := math.NewInt(500)
	transferAmt2 := math.NewInt(500)
	res, err := suite.factory.ExecuteContractCall(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &bankPrecompileAddr,
		},
		testutiltypes.CallArgs{
			ContractABI: abi,
			MethodName:  "multiSend",
			Args: []interface{}{
				sdk.NewCoins(sdk.NewCoin("ario", transferAmt)).String(),
				[]precompileBank.Output{
					{
						Addr:   recipient1Key.Addr,
						Amount: sdk.NewCoins(sdk.NewCoin("ario", transferAmt1)).String(),
					},
					{
						Addr:   recipient2Key.Addr,
						Amount: sdk.NewCoins(sdk.NewCoin("ario", transferAmt2)).String(),
					},
				},
			},
		},
	)
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "bank transfer should have succeeded", res.GetLog())

	// Move to next block
	err = suite.network.NextBlock()
	suite.Require().NoError(err)

	// Verify balances after transfer

	senderFinalBalance := suite.GetBalancesByBankPrecompile(senderPriv, senderKey.Addr)
	recipient1FinalBalance := suite.GetBalancesByBankPrecompile(senderPriv, recipient1Key.Addr)
	recipient2FinalBalance := suite.GetBalancesByBankPrecompile(senderPriv, recipient2Key.Addr)

	suite.Require().True(recipient1FinalBalance.AmountOf("ario").Equal(recipient1InitialBalance.AmountOf("ario").Add(transferAmt1)))
	suite.Require().True(recipient2FinalBalance.AmountOf("ario").Equal(recipient2InitialBalance.AmountOf("ario").Add(transferAmt2)))
	// sender balance less cause of gas
	suite.Require().True(senderFinalBalance.AmountOf("ario").LT(senderInitialBalance.AmountOf("ario").Sub(transferAmt)))
}

func (suite *EVMTestSuite) GetBalancesByBankPrecompile(senderPriv cryptotypes.PrivKey, senderAddr common.Address) sdk.Coins {
	// Verify balances after transfer
	_, balRes, err := suite.factory.CallContractAndCheckLogs(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &bankPrecompileAddr,
		},
		testutiltypes.CallArgs{
			ContractABI: abi,
			MethodName:  "balances",
			Args: []interface{}{
				senderAddr,
			},
		},
		testutil.LogCheckArgs{ExpPass: true},
	)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	var bals []precompileBank.Balance
	err = abi.UnpackIntoInterface(&bals, "balances", balRes.Ret)
	suite.Require().NoError(err)

	coins := sdk.NewCoins()
	for _, bal := range bals {
		coins = coins.Add(sdk.NewCoin(bal.Denom, math.NewIntFromBigInt(bal.Amount)))
	}

	return coins
}
