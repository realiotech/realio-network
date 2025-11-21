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

// TestBankPrecompileTransfer tests the bank precompile transfer functionality
func (suite *EVMTestSuite) TestBankPrecompileTransfer() {
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

	coins := sdk.NewCoins()
	for _, bal := range bals {
		coins = coins.Add(sdk.NewCoin(bal.Denom, math.NewIntFromBigInt(bal.Amount)))
	}

	return coins
}

// // TestBankPrecompileMultiTransfer tests transferring multiple coin types through bank precompile
// func (suite *EVMTestSuite) TestBankPrecompileMultiTransfer() {
// 	// Get test accounts
// 	senderPriv := suite.keyring.GetPrivKey(0)
// 	senderKey := suite.keyring.GetKey(0)
// 	recipientKey := suite.keyring.GetKey(1)

// 	// Get initial balances
// 	bankClient := suite.network.GetBankClient()
// 	ctx := suite.network.GetContext()
// 	baseDenom := suite.network.GetBaseDenom()

// 	senderInitialBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
// 		Address: senderKey.AccAddr.String(),
// 		Denom:   baseDenom,
// 	})
// 	suite.Require().NoError(err)

// 	// Load bank precompile ABI
// 	abi, err := precompileBank.LoadABI()
// 	suite.Require().NoError(err)

// 	// Get bank precompile address
// 	bankPrecompileAddr := common.HexToAddress(precompileBank.BankPrecompileAddress)

// 	// Execute multi-coin transfer through bank precompile
// 	transferAmount1 := math.NewInt(100_000)
// 	transferAmount2 := math.NewInt(50_000)

// 	res, err := suite.factory.ExecuteContractCall(
// 		senderPriv,
// 		evmtypes.EvmTxArgs{
// 			To: &bankPrecompileAddr,
// 		},
// 		testutiltypes.CallArgs{
// 			ContractABI: abi,
// 			MethodName:  "send",
// 			Args: []interface{}{
// 				senderKey.Addr,
// 				recipientKey.Addr,
// 				[]interface{}{
// 					map[string]interface{}{
// 						"denom":  baseDenom,
// 						"amount": transferAmount1.String(),
// 					},
// 					map[string]interface{}{
// 						"denom":  "stake",
// 						"amount": transferAmount2.String(),
// 					},
// 				},
// 			},
// 		},
// 	)
// 	suite.Require().NoError(err)
// 	suite.Require().True(res.IsOK(), "multi-coin transfer should have succeeded", res.GetLog())

// 	// Move to next block
// 	err = suite.network.NextBlock()
// 	suite.Require().NoError(err)

// 	// Verify balances after transfer
// 	senderFinalBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
// 		Address: senderKey.AccAddr.String(),
// 		Denom:   baseDenom,
// 	})
// 	suite.Require().NoError(err)

// 	recipientFinalBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
// 		Address: recipientKey.AccAddr.String(),
// 		Denom:   baseDenom,
// 	})
// 	suite.Require().NoError(err)

// 	// Verify sender balance decreased
// 	expectedSenderBalance := senderInitialBalance.Balance.Amount.Sub(transferAmount1)
// 	suite.Require().Equal(
// 		expectedSenderBalance.String(),
// 		senderFinalBalance.Balance.Amount.String(),
// 		"sender balance should have decreased by transfer amount",
// 	)

// 	// Verify recipient balance increased
// 	expectedRecipientBalance := recipientInitialBalance.Balance.Amount.Add(transferAmount1)
// 	suite.Require().Equal(
// 		expectedRecipientBalance.String(),
// 		recipientFinalBalance.Balance.Amount.String(),
// 		"recipient balance should have increased by transfer amount",
// 	)

// 	fmt.Printf("✅ Bank precompile multi-transfer test passed!\n")
// 	fmt.Printf("   Transferred: %s %s + %s stake\n", transferAmount1.String(), baseDenom, transferAmount2.String())
// 	fmt.Printf("   From: %s\n", senderKey.AccAddr.String())
// 	fmt.Printf("   To: %s\n", recipientKey.AccAddr.String())
// }
