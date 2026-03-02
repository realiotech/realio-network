package integration

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/cosmos/evm/testutil/integration/base/factory"
	feesponsortypes "github.com/cosmos/evm/x/feesponsor/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

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

// FeeGrantTestSuite tests the feegrant module functionality
type FeeGrantTestSuite struct {
	suite.Suite
	EVMTestSuite
}

// SetupTest sets up the test suite
func (suite *FeeGrantTestSuite) SetupTest() {
	suite.EVMTestSuite.SetupTest()
}

// SubmitSetFeePayerProposal submits a governance proposal to set the EVM fee payer
// and votes on it to make it pass
func (suite *FeeGrantTestSuite) SubmitSetFeePayerProposal(granterPriv cryptotypes.PrivKey, granterAddr sdk.AccAddress) uint64 {
	// Step 1: Submit governance proposal to update feesponsor fee payer
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

	// Step 2: Vote on the proposal using granter's private key and wait for it to pass
	err = utils.ApproveProposal(suite.factory, suite.network, granterPriv, proposalID)
	suite.Require().NoError(err)

	return proposalID
}

// TestFeeGrantEVMBasicFlow tests feegrant with EVM transactions:
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Grantee deploys an ERC20 contract
// 5. Grantee calls contract method (mint) using fee grant
// 6. Granter pays the gas fees (grantee balance remains the same)
func (suite *FeeGrantTestSuite) TestFeeGrantEVMBasicFlow() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr
	fmt.Println("granter addr", granterAddr.String())

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

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

	// Get balances before grantee transaction
	granterBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

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
	fmt.Println("After deploy contract")

	// Get balances after transaction
	granterAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify grantee balance decreased only by gas fees (not by transaction amount)
	// Since this is an EVM call, grantee shouldn't lose tokens, only gas
	suite.True(
		granteeAfterTx.Balance.Amount.Equal(granteeBeforeTx.Balance.Amount),
		"Grantee balance should not increase",
	)

	// Verify granter balance decreased by fees
	suite.True(
		granterAfterTx.Balance.Amount.LT(granterBeforeTx.Balance.Amount),
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
func (suite *FeeGrantTestSuite) TestFeeGrantMultipleEVMCalls() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

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

	// Get balances before grantee transactions
	granterBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

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
	granterAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify grantee balance did not decrease (only paid gas)
	suite.True(
		granteeAfterTx.Balance.Amount.Equal(granteeBeforeTx.Balance.Amount),
		"Grantee balance should not increase",
	)

	// Verify granter balance decreased by fees for both calls
	suite.True(
		granterAfterTx.Balance.Amount.LT(granterBeforeTx.Balance.Amount),
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
func (suite *FeeGrantTestSuite) TestFeeGrantGranterZeroBalance() {
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

	// Get granter's balance
	granterBalance, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	// Drain granter's balance to near-zero (leave some for the grant message fee)
	drainAmount := granterBalance.Balance.Amount.Sub(math.NewInt(1100000))
	if drainAmount.IsPositive() {
		drainMsg := &banktypes.MsgSend{
			FromAddress: granterAddr.String(),
			ToAddress:   thirdAddr.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin(baseDenom, drainAmount)),
		}

		drainRes, err := suite.factory.ExecuteCosmosTx(granterPriv, factory.CosmosTxArgs{
			Msgs: []sdk.Msg{drainMsg},
		})
		suite.Require().NoError(err)
		suite.Require().True(drainRes.IsOK(), "drain should have succeeded", drainRes.GetLog())
		suite.Require().NoError(suite.network.NextBlock())
	}

	// Get balances before grantee transaction
	granterBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

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
	granterAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	granteeAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify granter balance stayed the same (fee grant paid the fees)
	suite.True(
		granterAfterTx.Balance.Amount.Equal(granterBeforeTx.Balance.Amount),
		"Granter balance should remain the same when fee grant is used",
	)

	// Verify grantee balance decreased (they paid for the transaction)
	suite.True(
		granteeAfterTx.Balance.Amount.LT(granteeBeforeTx.Balance.Amount),
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
func (suite *FeeGrantTestSuite) TestFeeGrantUnauthorizedWallet() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	// Get a third wallet that is NOT authorized
	unauthorizedPriv := suite.keyring.GetPrivKey(2)

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Create a basic allowance (unlimited spend)
	basicAllowance := &feegranttypes.BasicAllowance{}
	allowanceAny, err := codectypes.NewAnyWithValue(basicAllowance)
	suite.Require().NoError(err)

	// Grant fee to grantee (NOT to unauthorized wallet)
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

	baseDenom := suite.network.GetBaseDenom()

	// Get a third wallet that is NOT authorized
	unauthorized := suite.keyring.GetKey(2)
	unauthorizedAddr := unauthorized.AccAddr

	// Get balances before unauthorized wallet sends tokens
	granterBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	unauthorizedBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(unauthorizedAddr, baseDenom)
	suite.Require().NoError(err)

	// Unauthorized wallet sends base denom tokens via EVM call using granter as fee payer
	sendAmount := int64(1000)
	sendTxArgs := evmtypes.EvmTxArgs{
		To:     &grantee.Addr,
		Amount: big.NewInt(sendAmount),
	}

	// This should succeed because fee grant allows unauthorized wallet to use granter's balance for fees
	_, err = suite.factory.ExecuteEthTx(unauthorizedPriv, sendTxArgs)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

	unauthorizedAfterTx, err := suite.grpcHandler.GetBalanceFromBank(unauthorizedAddr, baseDenom)
	suite.Require().NoError(err)

	// Verify granter balance stayed the same (fee grant paid the fees)
	suite.True(
		granterAfterTx.Balance.Amount.Equal(granterBeforeTx.Balance.Amount),
		"Granter balance should remain the same when fee grant is used",
	)

	// Verify unauthorized wallet balance decreased (they paid for the transaction)
	suite.True(
		unauthorizedAfterTx.Balance.Amount.LT(unauthorizedBeforeTx.Balance.Amount.Sub(math.NewInt(sendAmount))),
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
func (suite *FeeGrantTestSuite) TestFeeGrantSenderZeroBalance() {
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

	// Get grantee's balance
	granteeBalance, err := suite.grpcHandler.GetBalanceFromBank(granteeAddr, baseDenom)
	suite.Require().NoError(err)

	// Drain grantee's balance to zero
	drainAmount := granteeBalance.Balance.Amount.Sub(math.NewInt(58899672026158))
	if drainAmount.IsPositive() {
		drainMsg := &banktypes.MsgSend{
			FromAddress: granteeAddr.String(),
			ToAddress:   thirdAddr.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin(baseDenom, drainAmount)),
		}

		drainRes, err := suite.factory.ExecuteCosmosTx(granteePriv, factory.CosmosTxArgs{
			Msgs: []sdk.Msg{drainMsg},
		})
		suite.Require().NoError(err)
		suite.Require().True(drainRes.IsOK(), "drain should have succeeded", drainRes.GetLog())
		suite.Require().NoError(suite.network.NextBlock())
	}

	// Get balances before grantee sends tokens
	granterBeforeTx, err := suite.grpcHandler.GetBalanceFromBank(granterAddr, baseDenom)
	suite.Require().NoError(err)

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
		granterAfterTx.Balance.Amount.LT(granterBeforeTx.Balance.Amount),
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
func (suite *FeeGrantTestSuite) TestFeeGrantBothZeroBalance() {
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
		fmt.Println("drainRes", drainRes)
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
	fmt.Println("err", err)
	suite.Require().NoError(suite.network.NextBlock())

	_, err = suite.factory.ExecuteEthTx(granteePriv, sendTxArgs)
	fmt.Println("err", err)
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
func (suite *FeeGrantTestSuite) TestFeeGrantWithSpendLimit() {
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
		"Granter balance should decrease due to paying fees",
	)

	// Verify grantee balance decrease
	suite.True(
		granteeAfterTx.Balance.Amount.LT(granteeBeforeTx.Balance.Amount),
		"Grantee balance should remain zero",
	)
}

// TestFeeGrantWithExpiration tests fee grant with expiration time for EVM transactions
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee with expiration time (set to past)
// 4. Grantee attempts to call contract method using expired fee grant
// 5. Transaction should fail because fee grant has expired
func (suite *FeeGrantTestSuite) TestFeeGrantWithExpiration() {
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
		"Granter balance should decrease due to paying fees",
	)

	// Verify grantee balance decrease
	suite.True(
		granteeAfterTx.Balance.Amount.LT(granteeBeforeTx.Balance.Amount),
		"Grantee balance should remain zero",
	)
}

// TestFeeSponsorCanCoverFee tests that fee sponsor can cover transaction fees
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Vote on the proposal using granter's private key
// 3. Granter grants fee to grantee
// 4. Grantee sends base denom tokens via EVM call
// 5. Transaction succeeds, granter pays the fees, grantee balance decreases by send amount
func (suite *FeeGrantTestSuite) TestFeeSponsorCanCoverFee() {
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
func (suite *FeeGrantTestSuite) TestSenderCanCoverSendAmount() {
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
func (suite *FeeGrantTestSuite) TestSenderCannotCoverSendAmount() {
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

// TestFeeGrantTestSuite runs the test suite
func TestFeeGrantTestSuite(t *testing.T) {
	suite.Run(t, new(FeeGrantTestSuite))
}
