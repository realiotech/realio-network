package integration

import (
	_ "embed"
	"encoding/json"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/evm/contracts"
	"github.com/cosmos/evm/precompiles/testutil"
	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	precompileMultiStaking "github.com/realiotech/realio-network/precompile/multistaking"
	integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
)

// MultistakingCaller is a test contract that performs ERC20 transfers before
// and/or after delegating through the multistaking precompile within the same
// transaction. Source: data/MultistakingCaller.sol.
//
//go:embed data/MultistakingCaller.json
var multistakingCallerJSON []byte

// TestMultistakingDelegateWithTransfer delegates through the multistaking
// precompile from a smart contract that also transfers the same ERC20 token
// within the same transaction. The token state written by the contract call
// and by the precompile's ERC20 conversion must stay consistent across
// orderings, and a failed delegation must revert the internal transfer too.
func (suite *EVMTestSuite) TestMultistakingDelegateWithTransfer() {
	senderPriv := suite.keyring.GetPrivKey(0)
	senderKey := suite.keyring.GetKey(0)
	receiverKey := suite.keyring.GetKey(1)

	var callerContract evmtypes.CompiledContract
	err := json.Unmarshal(multistakingCallerJSON, &callerContract)
	suite.Require().NoError(err)

	// Deploy ERC20 contract
	factoryy := factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	contractAddr, err := factoryy.DeployContract(
		senderPriv,
		evmtypes.EvmTxArgs{},
		testutiltypes.ContractDeploymentData{
			Contract:        contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{"StakeToken", "STAKE", uint8(18)},
		},
	)
	suite.Require().NoError(err)
	suite.NotEqual(contractAddr, common.Address{})
	suite.Require().NoError(suite.network.NextBlock())

	// Deploy the caller contract
	callerAddr, err := factoryy.DeployContract(
		senderPriv,
		evmtypes.EvmTxArgs{},
		testutiltypes.ContractDeploymentData{Contract: callerContract},
	)
	suite.Require().NoError(err)
	suite.NotEqual(callerAddr, common.Address{})
	suite.Require().NoError(suite.network.NextBlock())

	// Mint tokens to the validator creator and to the caller contract,
	// which acts as the delegator
	suite.mintERC20(contractAddr, senderKey.Addr, mintAmount, senderPriv)
	suite.mintERC20(contractAddr, callerAddr, mintAmount, senderPriv)

	// Add multistaking evm coin proposal
	bondWeightDec, err := math.LegacyNewDecFromStr(multistakingBondWeight)
	suite.Require().NoError(err)
	err = integrationutils.RegisterMultistakingEVMBondDenom(
		integrationutils.UpdateParamsInput{
			Tf:      factoryy,
			Network: suite.network,
			Pk:      senderPriv,
		},
		contractAddr.Hex(),
		bondWeightDec,
		senderKey.AccAddr,
	)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	// Create validator
	valOut := suite.createEVMValidatorByPrecompile(contractAddr, senderPriv, senderKey.AccAddr)
	valAddr := valOut.Validator.OperatorAddress

	callerDelegateAmount := int64(1_000_000)
	transferAmount := int64(1_000)

	testCases := []struct {
		name   string
		before bool
		after  bool
	}{
		{"internal transfer before precompile call", true, false},
		{"internal transfer after precompile call", false, true},
		{"internal transfer before and after precompile call", true, true},
	}

	expCallerBalance := mintAmount
	expReceiverBalance := int64(0)
	expDelegation := int64(0)

	for _, tc := range testCases {
		res, err := suite.factory.ExecuteContractCall(
			senderPriv,
			evmtypes.EvmTxArgs{To: &callerAddr},
			testutiltypes.CallArgs{
				ContractABI: callerContract.ABI,
				MethodName:  "testDelegateWithTransfer",
				Args: []interface{}{
					contractAddr,       // _token
					contractAddr.Hex(), // _tokenHex
					receiverKey.Addr,   // _receiver
					valAddr,            // _validator
					math.NewInt(callerDelegateAmount).String(), // _delegateAmount
					big.NewInt(transferAmount),                 // _transferAmount
					tc.before,                                  // _before
					tc.after,                                   // _after
				},
			},
		)
		suite.Require().NoError(err, tc.name)
		suite.Require().True(res.IsOK(), "delegate with transfer should have succeeded", tc.name, res.GetLog())
		suite.Require().NoError(suite.network.NextBlock())

		transfers := int64(0)
		if tc.before {
			transfers++
		}
		if tc.after {
			transfers++
		}
		expCallerBalance -= callerDelegateAmount + transfers*transferAmount
		expReceiverBalance += transfers * transferAmount
		expDelegation += callerDelegateAmount

		// Balances must reflect both the internal transfers and the ERC20
		// conversion performed by the precompile within the same tx
		suite.assertContractBalanceOf(contractAddr, callerAddr, expCallerBalance)
		suite.assertContractBalanceOf(contractAddr, receiverKey.Addr, expReceiverBalance)
		suite.assertDelegationByPrecompile(callerAddr, valAddr, expDelegation)
	}

	// Delegating more than the caller's balance must revert the whole tx,
	// including the internal transfer executed before the precompile call
	execRevertedCheck := testutil.LogCheckArgs{
		ExpPass:     false,
		ErrContains: vm.ErrExecutionReverted.Error(),
	}
	_, _, err = suite.factory.CallContractAndCheckLogs(
		senderPriv,
		evmtypes.EvmTxArgs{To: &callerAddr},
		testutiltypes.CallArgs{
			ContractABI: callerContract.ABI,
			MethodName:  "testDelegateWithTransfer",
			Args: []interface{}{
				contractAddr,
				contractAddr.Hex(),
				receiverKey.Addr,
				valAddr,
				math.NewInt(mintAmount).String(), // exceeds the caller's remaining balance
				big.NewInt(transferAmount),
				true,
				false,
			},
		},
		execRevertedCheck,
	)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	// No balance nor delegation change
	suite.assertContractBalanceOf(contractAddr, callerAddr, expCallerBalance)
	suite.assertContractBalanceOf(contractAddr, receiverKey.Addr, expReceiverBalance)
	suite.assertDelegationByPrecompile(callerAddr, valAddr, expDelegation)
}

func (suite *EVMTestSuite) assertDelegationByPrecompile(delAddr common.Address, valAddr string, expected int64) {
	abi, err := precompileMultiStaking.LoadABI()
	suite.Require().NoError(err)
	multistakingPrecompileAddr := common.HexToAddress(precompileMultiStaking.MultistakingPrecompileAddress)

	_, delRes, err := suite.factory.CallContractAndCheckLogs(
		suite.keyring.GetPrivKey(0),
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		testutiltypes.CallArgs{
			ContractABI: abi,
			MethodName:  "delegation",
			Args:        []interface{}{delAddr, valAddr},
		},
		testutil.LogCheckArgs{ExpPass: true},
	)
	suite.Require().NoError(err)

	var del precompileMultiStaking.DelegationOutput
	err = abi.UnpackIntoInterface(&del, "delegation", delRes.Ret)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(expected).BigInt(), del.Balance.Amount)

	suite.Require().NoError(suite.network.NextBlock())
}
