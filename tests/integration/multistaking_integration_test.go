package integration

import (
	"encoding/json"
	"fmt"

	// "fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/evm/testutil/integration/os/factory"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	// cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// gov1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	// integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
	// stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/evm/contracts"
	commonfactory "github.com/cosmos/evm/testutil/integration/common/factory"

	// erc20types "github.com/cosmos/evm/x/erc20/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
	// "github.com/realiotech/realio-network/app"
	"github.com/cosmos/evm/testutil/integration/os/grpc"
)

var (
	multistakingPrecompileAddr       = common.HexToAddress("0x0000000000000000000000000000000000000900")
	multistakingMintAmount     int64 = 10_000_000 // 10M tokens
	multistakingStakeAmount    int64 = 1_000_000  // 1M tokens for staking
	multistakingBondWeight           = "1.0"      // 1:1 bond weight
	amount                           = big.NewInt(multistakingStakeAmount).String()
)

func (suite *EVMTestSuite) TestMultistakingCreateValidator() {
	// Deploy ERC20 contract
	senderPriv := suite.keyring.GetPrivKey(0)
	senderKey := suite.keyring.GetKey(0)
	delPriv := suite.keyring.GetPrivKey(1)
	delKey := suite.keyring.GetKey(1)
	val2Priv := suite.keyring.GetPrivKey(2)
	val2Key := suite.keyring.GetKey(2)
	constructorArgs := []interface{}{"StakeToken", "STAKE", uint8(18)}
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract

	var err error
	factoryy := factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	contractAddr, err := factoryy.DeployContract(
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

	// Mint tokens to sender and delegator
	mintTxArgs := evmtypes.EvmTxArgs{To: &contractAddr}
	amountToMint := big.NewInt(multistakingMintAmount)
	mintArgs := factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{senderKey.Addr, amountToMint},
	}
	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	mintResponse, err := factoryy.ExecuteContractCall(senderPriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintResponse.IsOK(), "mint should have succeeded", mintResponse.GetLog())

	err = suite.network.NextBlock()
	suite.Require().NoError(err)

	mintArgs = factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{delKey.Addr, amountToMint},
	}

	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	mintResponse, err = factoryy.ExecuteContractCall(senderPriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintResponse.IsOK(), "mint should have succeeded", mintResponse.GetLog())

	err = suite.network.NextBlock()
	suite.Require().NoError(err)

	mintArgs = factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{val2Key.Addr, amountToMint},
	}
	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	mintResponse, err = factoryy.ExecuteContractCall(senderPriv, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintResponse.IsOK(), "mint should have succeeded", mintResponse.GetLog())

	err = suite.network.NextBlock()
	suite.Require().NoError(err)

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

	multistakingClient := suite.network.GetMultistakingClient()
	// stakingClient := suite.network.GetStakingClient()
	coinsInf, err := multistakingClient.MultiStakingCoinInfos(suite.network.GetContext(), &multistakingtypes.QueryMultiStakingCoinInfosRequest{})
	suite.Require().NoError(err)

	validatorAddress := sdk.ValAddress(senderKey.AccAddr).String()
	pk := secp256k1.GenPrivKey().PubKey()
	pkAny, err := codectypes.NewAnyWithValue(pk)
	suite.Require().NoError(err)

	// Create validator
	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	res, err := factoryy.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgCreateEVMValidator{
			Description:       stakingtypes.NewDescription("Test Validator", "test-identity", "https://test-validator.com", "security@test-validator.com", "Test validator for multistaking"),
			Commission:        stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(1, 2), math.LegacyNewDecWithPrec(2, 1), math.LegacyNewDecWithPrec(1, 2)),
			MinSelfDelegation: math.NewInt(1),
			ValidatorAddress:  validatorAddress,
			Pubkey:            pkAny,
			ContractAddress:   contractAddr.Hex(),
			Value:             math.NewInt(1_000_000),
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "register ERC20 should have succeeded", res.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify validator was created
	validatorRes, err := multistakingClient.Validator(suite.network.GetContext(), &multistakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddress,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(validatorRes.Validator)
	suite.Require().Equal(coinsInf.Infos[2].Denom, validatorRes.Validator.BondDenom)

	// Delegate more tokens to the validator
	delegateAmount := math.NewInt(500_000) // 500K tokens
	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	delegateResponse, err := factoryy.ExecuteCosmosTx(delPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgDelegateEVM{
			DelegatorAddress: delKey.AccAddr.String(),
			ValidatorAddress: validatorAddress,
			ContractAddress:  contractAddr.Hex(),
			Amount:           delegateAmount,
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(delegateResponse.IsOK(), "delegate should have succeeded", delegateResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify delegation
	delRes, err := suite.network.GetStakingClient().Delegation(suite.network.GetContext(), &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delKey.AccAddr.String(),
		ValidatorAddr: validatorAddress,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(delRes.DelegationResponse)

	// Create a new validator then BeginRedelegate
	pk2 := secp256k1.GenPrivKey().PubKey()
	pkAny2, err := codectypes.NewAnyWithValue(pk2)
	suite.Require().NoError(err)

	suite.Require().NoError(suite.network.NextBlock())

	// Create second validator
	validator2Address := sdk.ValAddress(val2Key.AccAddr).String()
	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	res2, err := factoryy.ExecuteCosmosTx(val2Priv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgCreateEVMValidator{
			Description:       stakingtypes.NewDescription("Test Validator 2", "test-identity-2", "https://test-validator-2.com", "security@test-validator-2.com", "Second test validator for multistaking"),
			Commission:        stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(15, 3), math.LegacyNewDecWithPrec(25, 2), math.LegacyNewDecWithPrec(15, 3)),
			MinSelfDelegation: math.NewInt(1),
			ValidatorAddress:  validator2Address,
			Pubkey:            pkAny2,
			ContractAddress:   contractAddr.Hex(),
			Value:             math.NewInt(1_000_000),
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(res2.IsOK(), "create second validator should have succeeded", res2.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Redelegate tokens from first validator to second validator
	redelegateAmount := math.NewInt(200_000) // 200K tokens
	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	reDelegateResponse, err := factoryy.ExecuteCosmosTx(delPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgBeginRedelegateEVM{
			DelegatorAddress:    delKey.AccAddr.String(),
			ValidatorSrcAddress: validatorAddress,
			ValidatorDstAddress: validator2Address,
			ContractAddress:     contractAddr.Hex(),
			Amount:              redelegateAmount,
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(reDelegateResponse.IsOK(), "redelegate should have succeeded", reDelegateResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify delegation
	delRes, err = suite.network.GetStakingClient().Delegation(suite.network.GetContext(), &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delKey.AccAddr.String(),
		ValidatorAddr: validator2Address,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(delRes.DelegationResponse)

	// Unbond token from the validator
	unbondAmount := math.NewInt(200_000)
	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	undelResponse, err := factoryy.ExecuteCosmosTx(delPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgUndelegateEVM{
			DelegatorAddress: delKey.AccAddr.String(),
			ValidatorAddress: validator2Address,
			ContractAddress:  contractAddr.Hex(),
			Amount:           unbondAmount,
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(undelResponse.IsOK(), "undelegate should have succeeded", undelResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Get unbonding delegation info to get creation height
	ubdRes, err := suite.network.GetStakingClient().UnbondingDelegation(suite.network.GetContext(), &stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: delKey.AccAddr.String(),
		ValidatorAddr: validator2Address,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(ubdRes.Unbond.Entries)
	suite.Require().NoError(suite.network.NextBlock())

	// CancelUnbondingDelegation
	currentHeight := ubdRes.Unbond.Entries[0].CreationHeight
	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	cancelResponse, err := factoryy.ExecuteCosmosTx(delPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgCancelUnbondingEVMDelegation{
			DelegatorAddress: delKey.AccAddr.String(),
			ValidatorAddress: validator2Address,
			ContractAddress:  contractAddr.Hex(),
			Amount:           unbondAmount,
			CreationHeight:   currentHeight,
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(cancelResponse.IsOK(), "cancelUnbondingDelegation should have succeeded", cancelResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	_, err = suite.network.GetStakingClient().UnbondingDelegation(suite.network.GetContext(), &stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: delKey.AccAddr.String(),
		ValidatorAddr: validator2Address,
	})
	suite.Require().Error(err)

	factoryy = factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	undelResponse, err = factoryy.ExecuteCosmosTx(delPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgUndelegateEVM{
			DelegatorAddress: delKey.AccAddr.String(),
			ValidatorAddress: validator2Address,
			ContractAddress:  contractAddr.Hex(),
			Amount:           unbondAmount,
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(undelResponse.IsOK(), "undelegate should have succeeded", undelResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	paramsRes, err := suite.network.GetStakingClient().Params(suite.network.GetContext(), &stakingtypes.QueryParamsRequest{})
	suite.Require().NoError(err)

	fmt.Println("before", suite.network.GetContext().BlockHeight())
	suite.network.NextBlock()
	fmt.Println("after", suite.network.GetContext().BlockHeight())

	suite.Require().NoError(suite.network.NextBlockAfter(paramsRes.Params.UnbondingTime))
	ubdResNew, err := suite.network.GetStakingClient().UnbondingDelegation(suite.network.GetContext(), &stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: delKey.AccAddr.String(),
		ValidatorAddr: validator2Address,
	})
	fmt.Println("ubdResNew", ubdResNew)
	suite.Require().Error(err)
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

// func (suite *EVMTestSuite) createValidator(privKey cryptotypes.PrivKey, validatorAddr string, contractAddr string) {
// 	// Create a sample Ed25519 public key (base64 encoded)
// 	// In a real scenario, this would be the validator's actual consensus public key
// 	samplePubKey := `{"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="}`

// 	amount := big.NewInt(multistakingStakeAmount).String()
// 	moniker := "Test Validator"
// 	identity := "test-identity"
// 	website := "https://test-validator.com"
// 	security := "security@test-validator.com"
// 	details := "Test validator for multistaking"
// 	commissionRate := "0.10"          // 10%
// 	commissionMaxRate := "0.20"       // 20%
// 	commissionMaxChangeRate := "0.01" // 1%
// 	minSelfDelegation := "1"

// 	// Call createValidator through multistaking precompile
// 	createValidatorTxArgs := evmtypes.EvmTxArgs{
// 		To: &multistakingPrecompileAddr,
// 	}
// 	createValidatorArgs := factory.CallArgs{
// 		ContractABI: suite.getMultistakingABI(),
// 		MethodName:  "createValidator",
// 		Args: []interface{}{
// 			validatorAddr,
// 			samplePubKey,
// 			contractAddr, // contractAddress
// 			amount,       // amount
// 			moniker,
// 			identity,
// 			website,
// 			security,
// 			details,
// 			commissionRate,
// 			commissionMaxRate,
// 			commissionMaxChangeRate,
// 			minSelfDelegation,
// 		},
// 	}

// 	createValidatorResponse, err := factoryy.ExecuteContractCall(privKey, createValidatorTxArgs, createValidatorArgs)
// 	suite.Require().NoError(err)
// 	suite.Require().True(createValidatorResponse.IsOK(), "createValidator should have succeeded", createValidatorResponse.GetLog())

// 	err = suite.network.NextBlock()
// 	suite.Require().NoError(err)
// }
