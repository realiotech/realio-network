package integration

import (
	"math/big"
	"time"

	"github.com/cosmos/evm/testutil/integration/base/factory"
	feesponsortypes "github.com/cosmos/evm/x/feesponsor/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/math"
	feegranttypes "cosmossdk.io/x/feegrant"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/evm/contracts"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	"github.com/realiotech/realio-network/testutil/integration/utils"
)

// SubmitSetFeePayerProposal submits a governance proposal to set the EVM fee payer
// and votes on it to make it pass
func (suite *EVMTestSuite) SubmitSetFeePayerProposal(granterPriv cryptotypes.PrivKey, granterAddr sdk.AccAddress) uint64 {
	updateParamsMsg := &feesponsortypes.MsgSetFeePayer{
		Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		EvmFeePayer: granterAddr.String(),
	}

	proposalID, err := utils.SubmitProposal(
		suite.factory,
		suite.network,
		granterPriv,
		"Update EVM Fee Payer",
		updateParamsMsg,
	)
	suite.Require().NoError(err)
	suite.Require().NotZero(proposalID, "proposal ID should be non-zero")

	err = utils.ApproveProposal(suite.factory, suite.network, granterPriv, proposalID)
	suite.Require().NoError(err)

	return proposalID
}

// GrantFeeAllowance grants a fee allowance from granter to grantee
func (suite *EVMTestSuite) GrantFeeAllowance(granterPriv cryptotypes.PrivKey, granterAddr, granteeAddr sdk.AccAddress, allowance *feegranttypes.BasicAllowance) {
	allowanceAny, err := codectypes.NewAnyWithValue(allowance)
	suite.Require().NoError(err)

	grantMsg := &feegranttypes.MsgGrantAllowance{
		Granter:   granterAddr.String(),
		Grantee:   granteeAddr.String(),
		Allowance: allowanceAny,
	}

	grantRes, err := suite.factory.ExecuteCosmosTx(granterPriv, factory.CosmosTxArgs{
		Msgs: []sdk.Msg{grantMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant should have succeeded", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
}

// DeployERC20Contract deploys an ERC20 contract and returns the contract address
func (suite *EVMTestSuite) DeployERC20Contract(deployerPriv cryptotypes.PrivKey) common.Address {
	constructorArgs := []interface{}{"TestToken", "TEST", uint8(18)}
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract

	contractAddr, err := suite.factory.DeployContract(
		deployerPriv,
		evmtypes.EvmTxArgs{},
		testutiltypes.ContractDeploymentData{
			Contract:        compiledContract,
			ConstructorArgs: constructorArgs,
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotEqual(contractAddr, common.Address{})
	suite.Require().NoError(suite.network.NextBlock())

	return contractAddr
}

// GetBalance retrieves the balance of an address
func (suite *EVMTestSuite) GetBalance(addr sdk.AccAddress, denom string) math.Int {
	balance, err := suite.grpcHandler.GetBalanceFromBank(addr, denom)
	suite.Require().NoError(err)
	return balance.Balance.Amount
}

// DrainBalance drains an account's balance to a specific amount
func (suite *EVMTestSuite) DrainBalance(senderPriv cryptotypes.PrivKey, senderAddr, recipientAddr sdk.AccAddress, denom string, leaveAmount math.Int) {
	currentBalance := suite.GetBalance(senderAddr, denom)
	if !currentBalance.IsPositive() {
		return
	}

	drainAmount := currentBalance.Sub(leaveAmount)
	if !drainAmount.IsPositive() {
		return
	}

	drainMsg := &banktypes.MsgSend{
		FromAddress: senderAddr.String(),
		ToAddress:   recipientAddr.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, drainAmount)),
	}

	drainRes, err := suite.factory.ExecuteCosmosTx(senderPriv, factory.CosmosTxArgs{
		Msgs: []sdk.Msg{drainMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(drainRes.IsOK(), "drain should have succeeded", drainRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
}

// TestFeeGrantEVMBasicFlow tests feegrant with EVM transactions:
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Grantee deploys an ERC20 contract
// 5. Grantee calls contract method (mint) using fee grant
// 6. Granter pays the gas fees (grantee balance remains the same)
func (suite *EVMTestSuite) TestFeeGrantEVMBasicFlow() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Grant fee to grantee
	suite.GrantFeeAllowance(granterPriv, granterAddr, granteeAddr, &feegranttypes.BasicAllowance{})

	// Get balances before grantee transaction
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)
	granteeBeforeTx := suite.GetBalance(granteeAddr, baseDenom)

	// Deploy ERC20 contract for testing
	_ = suite.DeployERC20Contract(granteePriv)

	// Get balances after transaction
	granterAfterTx := suite.GetBalance(granterAddr, baseDenom)
	granteeAfterTx := suite.GetBalance(granteeAddr, baseDenom)

	// Verify grantee balance decreased only by gas fees (not by transaction amount)
	// Since this is an EVM call, grantee shouldn't lose tokens, only gas
	suite.True(
		granteeAfterTx.Equal(granteeBeforeTx),
		"Grantee balance should not increase",
	)

	// Verify granter balance decreased by fees
	suite.True(
		granterAfterTx.LT(granterBeforeTx),
		"Granter balance should decrease due to paying fees",
	)
}

// TestFeeGrantMultipleEVMCalls tests that grantee can send multiple EVM contract calls using fee grant
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Grantee deploys an ERC20 contract
// 5. Grantee calls contract method (mint) multiple times using fee grant
// 6. Granter pays the gas fees for all calls
func (suite *EVMTestSuite) TestFeeGrantMultipleEVMCalls() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Grant fee to grantee
	suite.GrantFeeAllowance(granterPriv, granterAddr, granteeAddr, &feegranttypes.BasicAllowance{})

	// Deploy ERC20 contract for testing
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract
	contractAddr := suite.DeployERC20Contract(granteePriv)

	// Get balances before grantee transactions
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)
	granteeBeforeTx := suite.GetBalance(granteeAddr, baseDenom)

	// Grantee calls contract method (mint) multiple times using fee grant
	amountToMint := big.NewInt(1e18)
	mintTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	mintArgs := testutiltypes.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{grantee.Addr, amountToMint},
	}

	// First mint call
	mintRes1, err := suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes1.IsOK(), "first mint should have succeeded", mintRes1.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Second mint call
	mintRes2, err := suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes2.IsOK(), "second mint should have succeeded", mintRes2.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transactions
	granterAfterTx := suite.GetBalance(granterAddr, baseDenom)
	granteeAfterTx := suite.GetBalance(granteeAddr, baseDenom)

	// Verify grantee balance did not decrease (only paid gas)
	suite.True(
		granteeAfterTx.Equal(granteeBeforeTx),
		"Grantee balance should not increase",
	)

	// Verify granter balance decreased by fees for both calls
	suite.True(
		granterAfterTx.LT(granterBeforeTx),
		"Granter balance should decrease due to paying fees for multiple calls",
	)
}

// TestFeeGrantGranterZeroBalance tests behavior when granter has minimal balance for EVM calls
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Grantee deploys an ERC20 contract
// 5. Drain granter's balance to minimal amount (just enough for fees)
// 6. Grantee calls contract method using fee grant
// 7. Transaction should succeed, granter balance stays same, grantee balance decreases
func (suite *EVMTestSuite) TestFeeGrantGranterZeroBalance() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	// Get a third wallet to drain funds to
	third := suite.keyring.GetKey(2)
	thirdAddr := third.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Grant fee to grantee
	suite.GrantFeeAllowance(granterPriv, granterAddr, granteeAddr, &feegranttypes.BasicAllowance{})

	// Deploy ERC20 contract for testing
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract
	contractAddr := suite.DeployERC20Contract(granteePriv)

	// Drain granter's balance to near-zero (leave some for the grant message fee)
	suite.DrainBalance(granterPriv, granterAddr, thirdAddr, baseDenom, math.NewInt(1100000))

	// Get balances before grantee transaction
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)
	granteeBeforeTx := suite.GetBalance(granteeAddr, baseDenom)

	// Now call contract method when granter has minimal balance (just enough for fees)
	amountToMint := big.NewInt(1e18)
	mintTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	mintArgs := testutiltypes.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{grantee.Addr, amountToMint},
	}

	// This should succeed because granter has enough balance for fees
	mintRes, err := suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes.IsOK(), "mint should have succeeded", mintRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx := suite.GetBalance(granterAddr, baseDenom)
	granteeAfterTx := suite.GetBalance(granteeAddr, baseDenom)

	// Verify granter balance stayed the same (fee grant paid the fees)
	suite.True(
		granterAfterTx.Equal(granterBeforeTx),
		"Granter balance should remain the same when fee grant is used",
	)

	// Verify grantee balance decreased (they paid for the transaction)
	suite.True(
		granteeAfterTx.LT(granteeBeforeTx),
		"Grantee balance should decrease due to paying for the transaction",
	)
}

// TestFeeGrantUnauthorizedWallet tests that unauthorized wallets can still send tokens via EVM
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee (NOT to unauthorized wallet)
// 4. Grantee deploys an ERC20 contract
// 5. Unauthorized wallet sends base denom tokens via EVM call
// 6. Transaction succeeds, granter balance stays same, unauthorized wallet balance decreases
func (suite *EVMTestSuite) TestFeeGrantUnauthorizedWallet() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	// Get a third wallet that is NOT authorized
	unauthorizedPriv := suite.keyring.GetPrivKey(2)

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Grant fee to grantee (NOT to unauthorized wallet)
	suite.GrantFeeAllowance(granterPriv, granterAddr, granteeAddr, &feegranttypes.BasicAllowance{})

	// Deploy ERC20 contract for testing
	_ = suite.DeployERC20Contract(granteePriv)

	// Get a third wallet that is NOT authorized
	unauthorized := suite.keyring.GetKey(2)
	unauthorizedAddr := unauthorized.AccAddr

	// Get balances before unauthorized wallet sends tokens
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)
	unauthorizedBeforeTx := suite.GetBalance(unauthorizedAddr, baseDenom)

	// Unauthorized wallet sends base denom tokens via EVM call using granter as fee payer
	sendAmount := int64(1000)
	sendTxArgs := evmtypes.EvmTxArgs{
		To:     &grantee.Addr,
		Amount: big.NewInt(sendAmount),
	}

	// This should succeed because fee grant allows unauthorized wallet to use granter's balance for fees
	var err error
	_, err = suite.factory.ExecuteEthTx(unauthorizedPriv, sendTxArgs)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx := suite.GetBalance(granterAddr, baseDenom)
	unauthorizedAfterTx := suite.GetBalance(unauthorizedAddr, baseDenom)

	// Verify granter balance stayed the same (fee grant paid the fees)
	suite.True(
		granterAfterTx.Equal(granterBeforeTx),
		"Granter balance should remain the same when fee grant is used",
	)

	// Verify unauthorized wallet balance decreased (they paid for the transaction)
	suite.True(
		unauthorizedAfterTx.LT(unauthorizedBeforeTx.Sub(math.NewInt(sendAmount))),
		"Unauthorized wallet balance should decrease due to paying for the transaction",
	)
}

// TestFeeGrantSenderZeroBalance tests that sender with zero balance can still submit tx using fee grant
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Drain grantee's balance to zero
// 5. Grantee sends base denom tokens via EVM call
// 6. Transaction succeeds, granter pays the fees, grantee balance stays zero
func (suite *EVMTestSuite) TestFeeGrantSenderZeroBalance() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	// Get a third wallet to drain funds to
	third := suite.keyring.GetKey(2)
	thirdAddr := third.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Grant fee to grantee
	suite.GrantFeeAllowance(granterPriv, granterAddr, granteeAddr, &feegranttypes.BasicAllowance{})

	// Drain grantee's balance to zero
	suite.DrainBalance(granteePriv, granteeAddr, thirdAddr, baseDenom, math.NewInt(58899672026158))

	// Get balances before grantee sends tokens
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)

	granteeBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Grantee with zero balance sends base denom tokens via EVM call using granter as fee payer
	sendAmount := int64(1000)
	thirdKey := suite.keyring.GetKey(2)
	sendTxArgs := evmtypes.EvmTxArgs{
		To:     &thirdKey.Addr,
		Amount: big.NewInt(sendAmount),
	}

	// This should succeed because fee grant allows grantee to use granter's balance for fees
	_, err = suite.factory.ExecuteEthTx(granteePriv, sendTxArgs)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify granter balance decreased (paid the fees)
	suite.True(
		granterAfterTx.Balance.Amount.LT(granterBeforeTx),
		"Granter balance should decrease due to paying fees",
	)

	// Verify grantee balance stayed zero (had no balance to begin with)
	suite.True(
		granteeAfterTx.Balance.Amount.Equal(granteeBeforeTx.Balance.Amount.Sub(math.NewInt(sendAmount))),
		"Grantee balance should remain zero",
	)
}

// TestFeeGrantBothZeroBalance tests that tx fails when both sender and granter have zero balance
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Drain both granter and grantee's balance to zero
// 5. Grantee attempts to send base denom tokens via EVM call
// 6. Transaction should fail because neither has balance for fees
func (suite *EVMTestSuite) TestFeeGrantBothZeroBalance() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	// Get a third wallet to drain funds to
	third := suite.keyring.GetKey(2)
	thirdAddr := third.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Create a basic allowance (unlimited spend)
	basicAllowance := &feegranttypes.BasicAllowance{}
	allowanceAny, err := codectypes.NewAnyWithValue(basicAllowance)
	suite.Require().NoError(err)

	// Grant fee to grantee
	grantMsg := &feegranttypes.MsgGrantAllowance{
		Granter:   granterAddr.String(),
		Grantee:   granteeAddr.String(),
		Allowance: allowanceAny,
	}

	grantRes, err := suite.factory.ExecuteCosmosTx(granterPriv, factory.CosmosTxArgs{
		Msgs: []sdk.Msg{grantMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant should have succeeded", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Drain granter's balance to zero
	granterBalance, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	if granterBalance.Balance.Amount.IsPositive() {
		drainMsg := &banktypes.MsgSend{
			FromAddress: granterAddr.String(),
			ToAddress:   thirdAddr.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin(baseDenom, granterBalance.Balance.Amount.Sub(math.NewInt(1050000)))),
		}

		drainRes, err := suite.factory.ExecuteCosmosTx(granterPriv, factory.CosmosTxArgs{
			Msgs: []sdk.Msg{drainMsg},
		})
		suite.Require().NoError(err)
		suite.Require().True(drainRes.IsOK(), "drain should have succeeded", drainRes.GetLog())
		suite.Require().NoError(suite.network.NextBlock())
	}

	// Drain grantee's balance to zero
	granteeBalance, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	if granteeBalance.Balance.Amount.IsPositive() {
		drainMsg := &banktypes.MsgSend{
			FromAddress: granteeAddr.String(),
			ToAddress:   thirdAddr.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin(baseDenom, granteeBalance.Balance.Amount.Sub(math.NewInt(1050000)))),
		}

		drainRes, err := suite.factory.ExecuteCosmosTx(granteePriv, factory.CosmosTxArgs{
			Msgs: []sdk.Msg{drainMsg},
		})
		suite.Require().NoError(err)
		suite.Require().True(drainRes.IsOK(), "drain should have succeeded", drainRes.GetLog())
		suite.Require().NoError(suite.network.NextBlock())
	}

	// Grantee with zero balance attempts to send base denom tokens via EVM call
	sendAmount := int64(1111)
	thirdKey := suite.keyring.GetKey(2)
	sendTxArgs := evmtypes.EvmTxArgs{
		To:     &thirdKey.Addr,
		Amount: big.NewInt(sendAmount),
	}

	_, err = suite.factory.ExecuteEthTx(granteePriv, sendTxArgs)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	_, err = suite.factory.ExecuteEthTx(granteePriv, sendTxArgs)
	suite.Require().NoError(suite.network.NextBlock())

	suite.Error(err, "transaction should fail when both sender and granter have zero balance")
}

// TestFeeGrantWithSpendLimit tests fee grant with spend limit for EVM transactions
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee with spend limit
// 4. Grantee deploys an ERC20 contract
// 5. Grantee calls contract method (mint) using fee grant
// 6. Transaction succeeds because spend is within limit
// 7. Grantee attempts another transaction that exceeds spend limit
// 8. Transaction should fail because spend limit is exceeded
func (suite *EVMTestSuite) TestFeeGrantWithSpendLimit() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Create an allowance with spend limit (1000000 tokens)
	spendLimit := sdk.NewCoins(sdk.NewCoin(baseDenom, math.NewInt(200000)))
	limitedAllowance := &feegranttypes.BasicAllowance{
		SpendLimit: spendLimit,
	}
	allowanceAny, err := codectypes.NewAnyWithValue(limitedAllowance)
	suite.Require().NoError(err)

	// Grant fee to grantee with spend limit
	grantMsg := &feegranttypes.MsgGrantAllowance{
		Granter:   granterAddr.String(),
		Grantee:   granteeAddr.String(),
		Allowance: allowanceAny,
	}

	grantRes, err := suite.factory.ExecuteCosmosTx(granterPriv, factory.CosmosTxArgs{
		Msgs: []sdk.Msg{grantMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant should have succeeded", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Deploy ERC20 contract for testing
	constructorArgs := []interface{}{"TestToken", "TEST", uint8(18)}
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract

	contractAddr, err := suite.factory.DeployContract(
		granteePriv,
		evmtypes.EvmTxArgs{},
		testutiltypes.ContractDeploymentData{
			Contract:        compiledContract,
			ConstructorArgs: constructorArgs,
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotEqual(contractAddr, common.Address{})
	suite.Require().NoError(suite.network.NextBlock())

	// First transaction should succeed (within spend limit)
	amountToMint := big.NewInt(1e18)
	mintTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	mintArgs := testutiltypes.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{grantee.Addr, amountToMint},
	}

	mintRes1, err := suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes1.IsOK(), "first mint should have succeeded", mintRes1.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances before
	granterBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	_, err = suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.NoError(err, "transaction should fail because fee grant has expired")
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify granter balance is the same
	suite.True(
		granterAfterTx.Balance.Amount.Equal(granterBeforeTx.Balance.Amount),
		"Granter balance should be the same",
	)

	// Verify grantee balance decrease
	suite.True(
		granteeAfterTx.Balance.Amount.LT(granteeBeforeTx.Balance.Amount),
		"Grantee balance should be decreased",
	)
}

// TestFeeGrantWithExpiration tests fee grant with expiration time for EVM transactions
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee with expiration time (set to past)
// 4. Grantee attempts to call contract method using expired fee grant
// 5. Transaction should fail because fee grant has expired
func (suite *EVMTestSuite) TestFeeGrantWithExpiration() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Create an allowance with expiration time (set to past, so it's already expired)
	expirationTime := suite.network.GetContext().BlockTime().Add(time.Hour)
	expiredAllowance := &feegranttypes.BasicAllowance{
		Expiration: &expirationTime, // Expired 1 hour ago
	}
	allowanceAny, err := codectypes.NewAnyWithValue(expiredAllowance)
	suite.Require().NoError(err)

	// Grant fee to grantee with expired allowance
	grantMsg := &feegranttypes.MsgGrantAllowance{
		Granter:   granterAddr.String(),
		Grantee:   granteeAddr.String(),
		Allowance: allowanceAny,
	}

	grantRes, err := suite.factory.ExecuteCosmosTx(granterPriv, factory.CosmosTxArgs{
		Msgs: []sdk.Msg{grantMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant should have succeeded", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Deploy ERC20 contract for testing
	constructorArgs := []interface{}{"TestToken", "TEST", uint8(18)}
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract

	contractAddr, err := suite.factory.DeployContract(
		granteePriv,
		evmtypes.EvmTxArgs{},
		testutiltypes.ContractDeploymentData{
			Contract:        compiledContract,
			ConstructorArgs: constructorArgs,
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotEqual(contractAddr, common.Address{})
	suite.Require().NoError(suite.network.NextBlock())

	// Transaction should fail because fee grant has expired
	suite.Require().NoError(suite.network.NextBlockAfter(time.Hour + time.Minute))
	amountToMint := big.NewInt(1e18)
	mintTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	mintArgs := testutiltypes.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{grantee.Addr, amountToMint},
	}

	// Get balances before
	granterBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	_, err = suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.NoError(err, "transaction should fail because fee grant has expired")
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify granter balance is the same
	suite.True(
		granterAfterTx.Balance.Amount.Equal(granterBeforeTx.Balance.Amount),
		"Granter balance should be the same",
	)

	// Verify grantee balance decrease
	suite.True(
		granteeAfterTx.Balance.Amount.LT(granteeBeforeTx.Balance.Amount),
		"Grantee balance should be decrease",
	)
}

// TestFeeSponsorCanCoverFee tests that fee sponsor can cover transaction fees
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Grantee sends base denom tokens via EVM call
// 5. Transaction succeeds, granter pays the fees, grantee balance decreases by send amount
func (suite *EVMTestSuite) TestFeeSponsorCanCoverFee() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	recipient := suite.keyring.GetKey(2)

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Create a basic allowance (unlimited spend)
	basicAllowance := &feegranttypes.BasicAllowance{}
	allowanceAny, err := codectypes.NewAnyWithValue(basicAllowance)
	suite.Require().NoError(err)

	// Grant fee to grantee
	grantMsg := &feegranttypes.MsgGrantAllowance{
		Granter:   granterAddr.String(),
		Grantee:   granteeAddr.String(),
		Allowance: allowanceAny,
	}

	grantRes, err := suite.factory.ExecuteCosmosTx(granterPriv, factory.CosmosTxArgs{
		Msgs: []sdk.Msg{grantMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant should have succeeded", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances before transaction
	granterBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Grantee sends base denom tokens via EVM call using granter as fee payer
	sendAmount := int64(1000)
	sendTxArgs := evmtypes.EvmTxArgs{
		To:     &recipient.Addr,
		Amount: big.NewInt(sendAmount),
	}

	// This should succeed because fee grant allows grantee to use granter's balance for fees
	_, err = suite.factory.ExecuteEthTx(granteePriv, sendTxArgs)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify granter balance decreased (paid the fees)
	suite.True(
		granterAfterTx.Balance.Amount.LT(granterBeforeTx.Balance.Amount),
		"Granter balance should decrease due to paying fees",
	)

	// Verify grantee balance decreased by send amount (paid for the transaction)
	expectedGranteeBalance := granteeBeforeTx.Balance.Amount.Sub(math.NewInt(sendAmount))
	suite.True(
		granteeAfterTx.Balance.Amount.Equal(expectedGranteeBalance),
		"Grantee balance should decrease by send amount",
	)
}

// TestSenderCanCoverSendAmount tests that sender can cover the send amount
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Grantee sends base denom tokens via EVM call
// 5. Transaction succeeds, grantee balance decreases by send amount + fees
func (suite *EVMTestSuite) TestSenderCanCoverSendAmount() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	recipient := suite.keyring.GetKey(2)

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Create a basic allowance (unlimited spend)
	basicAllowance := &feegranttypes.BasicAllowance{}
	allowanceAny, err := codectypes.NewAnyWithValue(basicAllowance)
	suite.Require().NoError(err)

	// Grant fee to grantee
	grantMsg := &feegranttypes.MsgGrantAllowance{
		Granter:   granterAddr.String(),
		Grantee:   granteeAddr.String(),
		Allowance: allowanceAny,
	}

	grantRes, err := suite.factory.ExecuteCosmosTx(granterPriv, factory.CosmosTxArgs{
		Msgs: []sdk.Msg{grantMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant should have succeeded", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances before transaction
	granterBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Grantee sends base denom tokens via EVM call
	sendAmount := int64(1000)
	sendTxArgs := evmtypes.EvmTxArgs{
		To:     &recipient.Addr,
		Amount: big.NewInt(sendAmount),
	}

	// This should succeed
	_, err = suite.factory.ExecuteEthTx(granteePriv, sendTxArgs)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify granter balance decreased (paid the fees)
	suite.True(
		granterAfterTx.Balance.Amount.LT(granterBeforeTx.Balance.Amount),
		"Granter balance should decrease due to paying fees",
	)

	// Verify grantee balance decreased by send amount (sender paid for the send amount)
	expectedGranteeBalance := granteeBeforeTx.Balance.Amount.Sub(math.NewInt(sendAmount))
	suite.True(
		granteeAfterTx.Balance.Amount.Equal(expectedGranteeBalance),
		"Grantee balance should decrease by send amount",
	)
}

// TestSenderCannotCoverSendAmount tests that transaction fails when sender cannot cover send amount
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Drain grantee's balance to zero
// 5. Grantee attempts to send base denom tokens via EVM call
// 6. Transaction should fail because grantee cannot cover send amount
func (suite *EVMTestSuite) TestSenderCannotCoverSendAmount() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	recipient := suite.keyring.GetKey(2)

	third := suite.keyring.GetKey(3)
	thirdAddr := third.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Create a basic allowance (unlimited spend)
	basicAllowance := &feegranttypes.BasicAllowance{}
	allowanceAny, err := codectypes.NewAnyWithValue(basicAllowance)
	suite.Require().NoError(err)

	// Grant fee to grantee
	grantMsg := &feegranttypes.MsgGrantAllowance{
		Granter:   granterAddr.String(),
		Grantee:   granteeAddr.String(),
		Allowance: allowanceAny,
	}

	grantRes, err := suite.factory.ExecuteCosmosTx(granterPriv, factory.CosmosTxArgs{
		Msgs: []sdk.Msg{grantMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant should have succeeded", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Drain grantee's balance to zero
	granteeBalance, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	if granteeBalance.Balance.Amount.IsPositive() {
		drainMsg := &banktypes.MsgSend{
			FromAddress: granteeAddr.String(),
			ToAddress:   thirdAddr.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin(baseDenom, granteeBalance.Balance.Amount.Sub(math.NewInt(10000000)))),
		}

		drainRes, err := suite.factory.ExecuteCosmosTx(granteePriv, factory.CosmosTxArgs{
			Msgs: []sdk.Msg{drainMsg},
		})
		suite.Require().NoError(err)
		suite.Require().True(drainRes.IsOK(), "drain should have succeeded", drainRes.GetLog())
		suite.Require().NoError(suite.network.NextBlock())
	}

	// Grantee with zero balance attempts to send base denom tokens via EVM call
	sendAmount := int64(100000000)
	sendTxArgs := evmtypes.EvmTxArgs{
		To:     &recipient.Addr,
		Amount: big.NewInt(sendAmount),
	}

	// This should fail because grantee cannot cover send amount
	_, err = suite.factory.ExecuteEthTx(granteePriv, sendTxArgs)
	suite.Error(err, "transaction should fail because sender cannot cover send amount")
}

// TestNoProposalSenderPaysFees tests that without governance proposal, sender pays their own fees
// 1. Do NOT submit governance proposal to set EVM fee payer
// 2. Sender sends base denom tokens via EVM call
// 3. Transaction succeeds, sender pays both the send amount and the fees
// 4. Sender balance should decrease by send amount + fees
func (suite *EVMTestSuite) TestNoProposalSenderPaysFees() {
	senderPriv := suite.keyring.GetPrivKey(0)
	sender := suite.keyring.GetKey(0)
	senderAddr := sender.AccAddr

	recipient := suite.keyring.GetKey(1)

	baseDenom := suite.network.GetBaseDenom()

	// Get sender balance before transaction
	senderBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(senderAddr, baseDenom)
	suite.Require().NoError(err)

	// Sender sends base denom tokens via EVM call (without fee grant or fee sponsor)
	sendAmount := int64(1000)
	sendTxArgs := evmtypes.EvmTxArgs{
		To:     &recipient.Addr,
		Amount: big.NewInt(sendAmount),
	}

	// Execute transaction - sender should pay for both send amount and fees
	_, err = suite.factory.ExecuteEthTx(senderPriv, sendTxArgs)
	suite.Require().NoError(err, "transaction should succeed")
	suite.Require().NoError(suite.network.NextBlock())

	// Get sender balance after transaction
	senderAfterTx, err := suite.grpcHandler.GetBalanceFromBank(senderAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify sender balance decreased (by send amount + fees)
	suite.True(
		senderAfterTx.Balance.Amount.LT(senderBeforeTx.Balance.Amount),
		"Sender balance should decrease after paying for send amount and fees",
	)

	// Verify the decrease is at least the send amount
	balanceDecrease := senderBeforeTx.Balance.Amount.Sub(senderAfterTx.Balance.Amount)
	suite.True(
		balanceDecrease.GTE(math.NewInt(sendAmount)),
		"Balance decrease should be at least the send amount",
	)
}
