package integration

import (
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/math"
	feegranttypes "cosmossdk.io/x/feegrant"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/evm/contracts"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"

	precompileFeegrant "github.com/realiotech/realio-network/precompile/feegrant"
)

// TestFeeGrantPrecompile tests the full feegrant flow using the EVM precompile:
// 1. Submit governance proposal to update EVM fee payer as granter
// 2. Granter grants fee allowance to grantee via the feegrant precompile (EVM call)
// 3. Verify the grant was created in the feegrant module state
// 4. Grantee deploys an ERC20 contract using the fee grant
// 5. Grantee calls contract method (mint) using fee grant
// 6. Granter pays the gas fees (grantee balance remains the same)
func (suite *EVMTestSuite) TestFeeGrantPrecompile() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	// Gen new empty addr
	empAcc := testkeyring.NewKey().AccAddr

	baseDenom := suite.network.GetBaseDenom()

	// Submit governance proposal to set fee payer and vote on it
	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Grant fee allowance via feegrant precompile (EVM call) instead of cosmos tx
	feegrantABI, err := precompileFeegrant.LoadABI()
	suite.Require().NoError(err)
	feegrantPrecompileAddr := common.HexToAddress(precompileFeegrant.FeeGrantPrecompileAddress)

	grantRes, err := suite.factory.ExecuteContractCall(
		granterPriv,
		evmtypes.EvmTxArgs{
			To: &feegrantPrecompileAddr,
		},
		testutiltypes.CallArgs{
			ContractABI: feegrantABI,
			MethodName:  "grant",
			Args: []interface{}{
				grantee.Addr, // grantee address
				"",           // spendLimit (empty = unlimited)
				"",           // expiration (empty = no expiration)
				int64(0),     // period (0 = no periodic allowance)
				"",           // periodLimit (empty)
				[]string{},   // allowedMessages (empty = all messages)
			},
		},
	)
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant via precompile should have succeeded", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify the grant was created in the feegrant module state
	feegrantClient := suite.network.GetFeeGrantClient()
	allowanceRes, err := feegrantClient.Allowance(suite.network.GetContext(), &feegranttypes.QueryAllowanceRequest{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(allowanceRes.Allowance, "fee grant allowance should exist")

	// Grant to empty addr
	grantRes, err = suite.factory.ExecuteContractCall(
		granterPriv,
		evmtypes.EvmTxArgs{
			To: &feegrantPrecompileAddr,
		},
		testutiltypes.CallArgs{
			ContractABI: feegrantABI,
			MethodName:  "grant",
			Args: []interface{}{
				common.BytesToAddress(empAcc), // grantee address
				"",                            // spendLimit (empty = unlimited)
				"",                            // expiration (empty = no expiration)
				int64(0),                      // period (0 = no periodic allowance)
				"",                            // periodLimit (empty)
				[]string{},                    // allowedMessages (empty = all messages)
			},
		},
	)
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant via precompile should have succeeded", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify the grant was created in the feegrant module state
	feegrantClient = suite.network.GetFeeGrantClient()
	allowanceRes, err = feegrantClient.Allowance(suite.network.GetContext(), &feegranttypes.QueryAllowanceRequest{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(allowanceRes.Allowance, "fee grant allowance should exist")

	// Get balances before grantee transaction
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)
	granteeBeforeTx := suite.GetBalance(granteeAddr, baseDenom)

	// Deploy ERC20 contract for testing
	contractAddr := suite.DeployERC20Contract(granteePriv)

	// Grantee calls contract method (mint) using fee grant
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract
	amountToMint := big.NewInt(1e18)
	mintTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	mintArgs := testutiltypes.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{grantee.Addr, amountToMint},
	}

	mintRes, err := suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes.IsOK(), "mint should have succeeded", mintRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx := suite.GetBalance(granterAddr, baseDenom)
	granteeAfterTx := suite.GetBalance(granteeAddr, baseDenom)

	// Verify grantee balance did not decrease (granter paid fees)
	suite.True(
		granteeAfterTx.Equal(granteeBeforeTx),
		"Grantee balance should not decrease when granter pays fees",
	)

	// Verify granter balance decreased by fees
	suite.True(
		granterAfterTx.LT(granterBeforeTx),
		"Granter balance should decrease due to paying fees",
	)
}

// GrantFeeAllowanceByPrecompile is a helper that calls the feegrant precompile to grant an allowance.
func (suite *EVMTestSuite) GrantFeeAllowanceByPrecompile(
	granterPriv cryptotypes.PrivKey,
	granteeAddr common.Address,
	spendLimit string,
	expiration string,
	period int64,
	periodLimit string,
	allowedMessages []string,
) {
	feegrantABI, err := precompileFeegrant.LoadABI()
	suite.Require().NoError(err)
	feegrantPrecompileAddr := common.HexToAddress(precompileFeegrant.FeeGrantPrecompileAddress)

	grantRes, err := suite.factory.ExecuteContractCall(
		granterPriv,
		evmtypes.EvmTxArgs{To: &feegrantPrecompileAddr},
		testutiltypes.CallArgs{
			ContractABI: feegrantABI,
			MethodName:  "grant",
			Args: []interface{}{
				granteeAddr, spendLimit, expiration, period, periodLimit, allowedMessages,
			},
		},
	)
	suite.Require().NoError(err)
	suite.Require().True(grantRes.IsOK(), "grant via precompile should have succeeded: %s", grantRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
}

// TestFeeGrantPrecompile_BasicWithSpendLimit tests granting a BasicAllowance with a specific spend limit.
func (suite *EVMTestSuite) TestFeeGrantPrecompile_BasicWithSpendLimit() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Grant with spend limit of 500000 base tokens
	spendLimitAmount := math.NewInt(1000000000000000000)
	spendLimitStr := fmt.Sprintf("%s%s", spendLimitAmount.String(), baseDenom)
	suite.GrantFeeAllowanceByPrecompile(granterPriv, grantee.Addr, spendLimitStr, "", int64(0), "", []string{})

	// Verify the grant
	feegrantClient := suite.network.GetFeeGrantClient()
	allowanceRes, err := feegrantClient.Allowance(suite.network.GetContext(), &feegranttypes.QueryAllowanceRequest{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(allowanceRes.Allowance)

	// Get balances before grantee transaction
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)
	granteeBeforeTx := suite.GetBalance(granteeAddr, baseDenom)

	// Deploy ERC20 contract for testing
	contractAddr := suite.DeployERC20Contract(granteePriv)

	// Grantee calls contract method (mint) using fee grant
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract
	amountToMint := big.NewInt(1e18)
	mintTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	mintArgs := testutiltypes.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{grantee.Addr, amountToMint},
	}

	mintRes, err := suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes.IsOK(), "mint should have succeeded", mintRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx := suite.GetBalance(granterAddr, baseDenom)
	granteeAfterTx := suite.GetBalance(granteeAddr, baseDenom)

	// Verify grantee balance did not decrease (granter paid fees)
	suite.True(
		granteeAfterTx.Equal(granteeBeforeTx),
		"Grantee balance should not decrease when granter pays fees",
	)

	// Verify granter balance decreased by fees
	suite.True(
		granterAfterTx.LT(granterBeforeTx),
		"Granter balance should decrease due to paying fees",
	)
}

// TestFeeGrantPrecompile_BasicWithExpiration tests granting a BasicAllowance with an expiration.
func (suite *EVMTestSuite) TestFeeGrantPrecompile_BasicWithExpiration() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Grant with expiration 1 hour from now
	expirationTime := suite.network.GetContext().BlockTime().Add(time.Hour).UTC()
	expirationStr := expirationTime.Format(time.RFC3339)
	suite.GrantFeeAllowanceByPrecompile(granterPriv, grantee.Addr, "", expirationStr, int64(0), "", []string{})

	// Verify the grant
	feegrantClient := suite.network.GetFeeGrantClient()
	allowanceRes, err := feegrantClient.Allowance(suite.network.GetContext(), &feegranttypes.QueryAllowanceRequest{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(allowanceRes.Allowance)

	// Get balances before grantee transaction
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)
	granteeBeforeTx := suite.GetBalance(granteeAddr, baseDenom)

	// Deploy ERC20 contract for testing
	contractAddr := suite.DeployERC20Contract(granteePriv)

	// Grantee calls contract method (mint) using fee grant
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract
	amountToMint := big.NewInt(1e18)
	mintTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	mintArgs := testutiltypes.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{grantee.Addr, amountToMint},
	}

	mintRes, err := suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes.IsOK(), "mint should have succeeded", mintRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx := suite.GetBalance(granterAddr, baseDenom)
	granteeAfterTx := suite.GetBalance(granteeAddr, baseDenom)

	// Verify grantee balance did not decrease (granter paid fees)
	suite.True(
		granteeAfterTx.Equal(granteeBeforeTx),
		"Grantee balance should not decrease when granter pays fees",
	)

	// Verify granter balance decreased by fees
	suite.True(
		granterAfterTx.LT(granterBeforeTx),
		"Granter balance should decrease due to paying fees",
	)
}

// TestFeeGrantPrecompile_Periodic tests granting a PeriodicAllowance.
func (suite *EVMTestSuite) TestFeeGrantPrecompile_Periodic() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Grant periodic: 3600s period, 100000 base tokens per period
	periodLimitAmount := math.NewInt(1000000000000000000)
	periodLimitStr := fmt.Sprintf("%s%s", periodLimitAmount.String(), baseDenom)
	suite.GrantFeeAllowanceByPrecompile(granterPriv, grantee.Addr, "", "", int64(3600), periodLimitStr, []string{})

	// Verify the grant
	feegrantClient := suite.network.GetFeeGrantClient()
	allowanceRes, err := feegrantClient.Allowance(suite.network.GetContext(), &feegranttypes.QueryAllowanceRequest{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(allowanceRes.Allowance)

	// Get balances before grantee transaction
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)
	granteeBeforeTx := suite.GetBalance(granteeAddr, baseDenom)

	// Deploy ERC20 contract for testing
	contractAddr := suite.DeployERC20Contract(granteePriv)

	// Grantee calls contract method (mint) using fee grant
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract
	amountToMint := big.NewInt(1e18)
	mintTxArgs := evmtypes.EvmTxArgs{
		To: &contractAddr,
	}
	mintArgs := testutiltypes.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{grantee.Addr, amountToMint},
	}

	mintRes, err := suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes.IsOK(), "mint should have succeeded", mintRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx := suite.GetBalance(granterAddr, baseDenom)
	granteeAfterTx := suite.GetBalance(granteeAddr, baseDenom)

	// Verify grantee balance did not decrease (granter paid fees)
	suite.True(
		granteeAfterTx.Equal(granteeBeforeTx),
		"Grantee balance should not decrease when granter pays fees",
	)

	// Verify granter balance decreased by fees
	suite.True(
		granterAfterTx.LT(granterBeforeTx),
		"Granter balance should decrease due to paying fees",
	)
}

// TestFeeGrantPrecompile_AllowedMessages tests that:
// 1. A grant with non-EVM allowed messages (MsgSend, MsgDelegate) rejects EVM transactions
// 2. After revoking and re-granting with EVM message type, EVM transactions succeed
func (suite *EVMTestSuite) TestFeeGrantPrecompile_AllowedMessages() {
	granterPriv := suite.keyring.GetPrivKey(0)
	granter := suite.keyring.GetKey(0)
	granterAddr := granter.AccAddr

	granteePriv := suite.keyring.GetPrivKey(1)
	grantee := suite.keyring.GetKey(1)
	granteeAddr := grantee.AccAddr

	baseDenom := suite.network.GetBaseDenom()

	suite.SubmitSetFeePayerProposal(granterPriv, granterAddr)

	// Deploy ERC20 contract before restricting allowance (deploy is itself an EVM tx)
	contractAddr := suite.DeployERC20Contract(granteePriv)
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract
	amountToMint := big.NewInt(1e18)
	mintTxArgs := evmtypes.EvmTxArgs{To: &contractAddr}
	mintArgs := testutiltypes.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{grantee.Addr, amountToMint},
	}

	// --- Case 1: Grant with non-EVM allowed messages, EVM tx should fail ---
	nonEvmMsgs := []string{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.staking.v1beta1.MsgDelegate"}
	suite.GrantFeeAllowanceByPrecompile(granterPriv, grantee.Addr, "", "", int64(0), "", nonEvmMsgs)

	// Verify the grant exists
	feegrantClient := suite.network.GetFeeGrantClient()
	allowanceRes, err := feegrantClient.Allowance(suite.network.GetContext(), &feegranttypes.QueryAllowanceRequest{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(allowanceRes.Allowance)

	// Attempt EVM tx (mint) — fee grant won't cover it because allowance only permits MsgSend/MsgDelegate,
	// so the grantee pays the fee themselves.
	granterBeforeCase1 := suite.GetBalance(granterAddr, baseDenom)
	granteeBeforeCase1 := suite.GetBalance(granteeAddr, baseDenom)

	mintRes, err := suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes.IsOK(), "mint tx itself should succeed: %s", mintRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	granterAfterCase1 := suite.GetBalance(granterAddr, baseDenom)
	granteeAfterCase1 := suite.GetBalance(granteeAddr, baseDenom)

	// Granter balance should NOT decrease (fee grant did not apply)
	suite.True(
		granterAfterCase1.Equal(granterBeforeCase1),
		"Granter balance should not decrease when allowance does not cover EVM messages",
	)
	// Grantee balance should decrease (grantee paid the fee)
	suite.True(
		granteeAfterCase1.LT(granteeBeforeCase1),
		"Grantee balance should decrease because fee grant did not cover EVM tx",
	)

	// --- Revoke the grant via precompile ---
	feegrantABI, err := precompileFeegrant.LoadABI()
	suite.Require().NoError(err)
	feegrantPrecompileAddr := common.HexToAddress(precompileFeegrant.FeeGrantPrecompileAddress)

	revokeRes, err := suite.factory.ExecuteContractCall(
		granterPriv,
		evmtypes.EvmTxArgs{To: &feegrantPrecompileAddr},
		testutiltypes.CallArgs{
			ContractABI: feegrantABI,
			MethodName:  "revoke",
			Args:        []interface{}{grantee.Addr},
		},
	)
	suite.Require().NoError(err)
	suite.Require().True(revokeRes.IsOK(), "revoke via precompile should have succeeded: %s", revokeRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify the grant was revoked
	feegrantClient = suite.network.GetFeeGrantClient()
	_, err = feegrantClient.Allowance(suite.network.GetContext(), &feegranttypes.QueryAllowanceRequest{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().Error(err, "allowance should no longer exist after revoke")

	// --- Case 2: Grant with EVM message type, EVM tx should succeed ---
	evmMsgs := []string{"/cosmos.evm.vm.v1.MsgEthereumTx"}
	suite.GrantFeeAllowanceByPrecompile(granterPriv, grantee.Addr, "", "", int64(0), "", evmMsgs)

	// Get balances before grantee transaction
	granterBeforeTx := suite.GetBalance(granterAddr, baseDenom)
	granteeBeforeTx := suite.GetBalance(granteeAddr, baseDenom)

	// EVM tx (mint) — should succeed because allowance permits MsgEthereumTx
	mintRes, err = suite.factory.ExecuteContractCall(granteePriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintRes.IsOK(), "mint should have succeeded with EVM allowed msg: %s", mintRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get balances after transaction
	granterAfterTx := suite.GetBalance(granterAddr, baseDenom)
	granteeAfterTx := suite.GetBalance(granteeAddr, baseDenom)

	// Verify grantee balance did not decrease (granter paid fees)
	suite.True(
		granteeAfterTx.Equal(granteeBeforeTx),
		"Grantee balance should not decrease when granter pays fees",
	)

	// Verify granter balance decreased by fees
	suite.True(
		granterAfterTx.LT(granterBeforeTx),
		"Granter balance should decrease due to paying fees",
	)
}
