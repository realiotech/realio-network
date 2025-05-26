package integration

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/evm/testutil/integration/os/factory"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	// gov1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	// integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/evm/contracts"
	commonfactory "github.com/cosmos/evm/testutil/integration/common/factory"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

)

var (
	multistakingPrecompileAddr       = common.HexToAddress("0x0000000000000000000000000000000000000900")
	multistakingMintAmount     int64 = 10_000_000 // 10M tokens
	multistakingStakeAmount    int64 = 1_000_000  // 1M tokens for staking
	multistakingBondWeight           = "1.0"      // 1:1 bond weight
	amount = big.NewInt(multistakingStakeAmount).String()
)

func (suite *EVMTestSuite) TestMultistakingCreateValidator() {
	// Deploy ERC20 contract
	senderPriv := suite.keyring.GetPrivKey(0)
	senderKey := suite.keyring.GetKey(0)
	constructorArgs := []interface{}{"StakeToken", "STAKE", uint8(18)}
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract

	var err error
	contractAddr, err := suite.factory.DeployContract(
		senderPriv,
		evmtypes.EvmTxArgs{},
		factory.ContractDeploymentData{
			Contract:        compiledContract,
			ConstructorArgs: constructorArgs,
		},
	)
	suite.Require().NoError(err)
	suite.NotEqual(contractAddr, common.Address{})

	err = suite.network.NextBlock()
	suite.Require().NoError(err)

	// Mint tokens to sender
	mintTxArgs := evmtypes.EvmTxArgs{To: &contractAddr}
	amountToMint := big.NewInt(multistakingMintAmount)
	mintArgs := factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{senderKey.Addr, amountToMint},
	}
	mintResponse, err := suite.factory.ExecuteContractCall(senderPriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintResponse.IsOK(), "mint should have succeeded", mintResponse.GetLog())

	err = suite.network.NextBlock()
	suite.Require().NoError(err)

	// Register ERC20 token as native token
	res, err := suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&erc20types.MsgRegisterERC20{
			Signer:         senderKey.AccAddr.String(),
			Erc20Addresses: []string{contractAddr.Hex()},
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "register ERC20 should have succeeded", res.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Query native denom of contract
	erc20Client := suite.network.GetERC20Client()
	tokenPairRes, err := erc20Client.TokenPair(suite.network.GetContext(), &erc20types.QueryTokenPairRequest{
		Token: contractAddr.Hex(),
	})
	suite.Require().NoError(err)
	contractDenom := tokenPairRes.TokenPair.Denom

	// Add multistaking coin proposal
	bondWeightDec, err := math.LegacyNewDecFromStr(multistakingBondWeight)
	suite.Require().NoError(err)

	err = integrationutils.RegisterMultistakingBondDenom(
		integrationutils.UpdateParamsInput{
			Tf:      suite.factory,
			Network: suite.network,
			Pk:      senderPriv,
		},
		contractDenom,
		bondWeightDec,
		senderKey.AccAddr,
	)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	// For testing purposes, we'll assume the proposal passes automatically
	// In a real scenario, you'd need to vote and wait for the voting period

	// Prepare validator creation parameters
	validatorAddress := sdk.ValAddress(senderKey.AccAddr).String()
	suite.createValidator(senderPriv, validatorAddress, contractAddr.Hex())

	// Verify validator was created
	multistakingClient := suite.network.GetMultistakingClient()
	validatorRes, err := multistakingClient.Validator(suite.network.GetContext(), &multistakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddress,
	})
	fmt.Println("validatorRes", validatorRes)
	suite.Require().NoError(err)
	suite.Require().NotNil(validatorRes.Validator)
	suite.Require().Equal(contractDenom, validatorRes.Validator.BondDenom)

	// Delegate more tokens to the validator
	delegateTxArgs := evmtypes.EvmTxArgs{
		To: &multistakingPrecompileAddr,
		GasLimit: 5_890_256,
	}
	delegateArgs := factory.CallArgs{
		ContractABI: suite.getMultistakingABI(),
		MethodName:  "delegate",
		Args: []interface{}{
			contractAddr.Hex(), // contractAddress
			validatorAddress,
			amount, // amount
		},
	}
	delegateResponse, err := suite.factory.ExecuteContractCall(senderPriv, delegateTxArgs, delegateArgs)
	suite.Require().NoError(err)
	suite.Require().True(delegateResponse.IsOK(), "delegate should have succeeded", delegateResponse.GetLog())

	err = suite.network.NextBlock()
	suite.Require().NoError(err)


	// Verify delegation was created
	stakingClient := suite.network.GetStakingClient()
	delegationRes, err := stakingClient.Delegation(suite.network.GetContext(), &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: senderKey.AccAddr.String(),
		ValidatorAddr: validatorAddress,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(delegationRes.DelegationResponse)
	suite.Require().Equal(math.NewInt(multistakingStakeAmount*2), delegationRes.DelegationResponse.Balance.Amount)

	// Verify ERC20 balance decreased (tokens were converted for staking)
	balanceTxArgs := evmtypes.EvmTxArgs{To: &contractAddr}
	balanceArgs := factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "balanceOf",
		Args:        []interface{}{senderKey.Addr},
	}
	balanceResponse, err := suite.factory.ExecuteContractCall(senderPriv, balanceTxArgs, balanceArgs)
	suite.Require().NoError(err)

	var finalBalance *big.Int
	err = integrationutils.DecodeContractCallResponse(&finalBalance, balanceArgs, balanceResponse)
	suite.Require().NoError(err)
	expectedBalance := big.NewInt(multistakingMintAmount - multistakingStakeAmount*2)
	suite.Require().Equal(expectedBalance, finalBalance)
}

// Helper function to get multistaking ABI
func (suite *EVMTestSuite) getMultistakingABI() abi.ABI {
	// Read the ABI file from the filesystem
	abiPath := filepath.Join("..", "..", "precompiles", "multistaking", "abi.json")
	abiBytes, err := os.ReadFile(abiPath)
	suite.Require().NoError(err)

	// Parse the ABI file structure
	var abiData struct {
		ABI json.RawMessage `json:"abi"`
	}

	err = json.Unmarshal(abiBytes, &abiData)
	suite.Require().NoError(err)

	parsedABI, err := abi.JSON(strings.NewReader(string(abiData.ABI)))
	suite.Require().NoError(err)
	return parsedABI
}

func (suite *EVMTestSuite) createValidator(privKey cryptotypes.PrivKey, validatorAddr string, contractAddr string) {
	// Create a sample Ed25519 public key (base64 encoded)
	// In a real scenario, this would be the validator's actual consensus public key
	samplePubKey := `{"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="}`

	amount := big.NewInt(multistakingStakeAmount).String()
	moniker := "Test Validator"
	identity := "test-identity"
	website := "https://test-validator.com"
	security := "security@test-validator.com"
	details := "Test validator for multistaking"
	commissionRate := "0.10"          // 10%
	commissionMaxRate := "0.20"       // 20%
	commissionMaxChangeRate := "0.01" // 1%
	minSelfDelegation := "1"

	// Call createValidator through multistaking precompile
	createValidatorTxArgs := evmtypes.EvmTxArgs{
		To: &multistakingPrecompileAddr,
	}
	createValidatorArgs := factory.CallArgs{
		ContractABI: suite.getMultistakingABI(),
		MethodName:  "createValidator",
		Args: []interface{}{
			validatorAddr,
			samplePubKey,
			contractAddr, // contractAddress
			amount,             // amount
			moniker,
			identity,
			website,
			security,
			details,
			commissionRate,
			commissionMaxRate,
			commissionMaxChangeRate,
			minSelfDelegation,
		},
	}

	createValidatorResponse, err := suite.factory.ExecuteContractCall(privKey, createValidatorTxArgs, createValidatorArgs)
	suite.Require().NoError(err)
	suite.Require().True(createValidatorResponse.IsOK(), "createValidator should have succeeded", createValidatorResponse.GetLog())

	err = suite.network.NextBlock()
	suite.Require().NoError(err)
}
