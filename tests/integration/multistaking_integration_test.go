package integration

import (

	// "fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/evm/testutil/integration/os/factory"
	evmtypes "github.com/cosmos/evm/x/vm/types"
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
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/evm/testutil/integration/os/grpc"
)

var (
	mintAmount             int64 = 1_000_000
	delegateAmount         int64 = 500_000
	redelegateAmount       int64 = 200_000
	undelegateAmount       int64 = 200_000
	multistakingBondWeight       = "1.0" // 1:1 bond weight
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

	// Mint tokens to sender, delegator and validator
	suite.mintERC20(contractAddr, senderKey.Addr, mintAmount, senderPriv)
	suite.mintERC20(contractAddr, delKey.Addr, mintAmount, senderPriv)
	suite.mintERC20(contractAddr, val2Key.Addr, mintAmount, senderPriv)

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
	coinsInf, err := multistakingClient.MultiStakingCoinInfos(suite.network.GetContext(), &multistakingtypes.QueryMultiStakingCoinInfosRequest{})
	suite.Require().NoError(err)

	// Create 1st validator
	validator1Address := sdk.ValAddress(senderKey.AccAddr).String()
	suite.createEVMValidator(contractAddr, senderPriv, validator1Address)

	// Verify validator was created
	validatorRes, err := multistakingClient.Validator(suite.network.GetContext(), &multistakingtypes.QueryValidatorRequest{
		ValidatorAddr: validator1Address,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(validatorRes.Validator)
	suite.Require().Equal(coinsInf.Infos[2].Denom, validatorRes.Validator.BondDenom)

	suite.delegateEVM(contractAddr, delPriv, delKey.AccAddr.String(), validator1Address)

	// Verify delegation
	delRes, err := suite.network.GetStakingClient().Delegation(suite.network.GetContext(), &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delKey.AccAddr.String(),
		ValidatorAddr: validator1Address,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(delRes.DelegationResponse)

	// Create a new validator then BeginRedelegate

	// Create second validator
	validator2Address := sdk.ValAddress(val2Key.AccAddr).String()
	suite.createEVMValidator(contractAddr, val2Priv, validator2Address)

	// Redelegate tokens from first validator to second validator
	suite.redelegateEVM(contractAddr, delPriv, delKey.AccAddr.String(), validator1Address, validator2Address)

	// Verify delegation
	delRes, err = suite.network.GetStakingClient().Delegation(suite.network.GetContext(), &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delKey.AccAddr.String(),
		ValidatorAddr: validator2Address,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(delRes.DelegationResponse)

	// Unbond token from the validator 2
	suite.undelegateEVM(contractAddr, delPriv, delKey.AccAddr.String(), validator2Address)

	// Get unbonding delegation info to get creation height
	ubdRes, err := suite.network.GetStakingClient().UnbondingDelegation(suite.network.GetContext(), &stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: delKey.AccAddr.String(),
		ValidatorAddr: validator2Address,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(ubdRes.Unbond.Entries)
	suite.Require().NoError(suite.network.NextBlock())

	// CancelUnbondingDelegation
	creationHeight := ubdRes.Unbond.Entries[0].CreationHeight
	suite.cancelUndelegateEvm(contractAddr, delPriv, delKey.AccAddr.String(), validator2Address, creationHeight)

	_, err = suite.network.GetStakingClient().UnbondingDelegation(suite.network.GetContext(), &stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: delKey.AccAddr.String(),
		ValidatorAddr: validator2Address,
	})
	suite.Require().Error(err) // empty

	suite.undelegateEVM(contractAddr, delPriv, delKey.AccAddr.String(), validator2Address)

	paramsRes, err := suite.network.GetStakingClient().Params(suite.network.GetContext(), &stakingtypes.QueryParamsRequest{})
	suite.Require().NoError(err)

	// User should get back their unbond tokens after unbonding period
	expectedBalanceBefore := mintAmount - delegateAmount
	suite.assertContractBalanceOf(contractAddr, delKey.Addr, expectedBalanceBefore)

	suite.Require().NoError(suite.network.NextBlockAfter(paramsRes.Params.UnbondingTime))
	suite.assertContractBalanceOf(contractAddr, delKey.Addr, expectedBalanceBefore+undelegateAmount)
}

func (suite *EVMTestSuite) mintERC20(contractAddr common.Address, to common.Address, amount int64, privKey cryptotypes.PrivKey) {
	mintTxArgs := evmtypes.EvmTxArgs{To: &contractAddr}
	mintArgs := factory.CallArgs{
		ContractABI: compiledContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{to, big.NewInt(amount)},
	}
	mintResponse, err := suite.factory.ExecuteContractCall(privKey, mintTxArgs, mintArgs)
	suite.Require().NoError(err)
	suite.Require().True(mintResponse.IsOK(), "mint should have succeeded", mintResponse.GetLog())

	err = suite.network.NextBlock()
	suite.Require().NoError(err)
}

func (suite *EVMTestSuite) createEVMValidator(contractAddr common.Address, senderPriv cryptotypes.PrivKey, valAddr string) {
	pk := secp256k1.GenPrivKey().PubKey()
	pkAny, err := codectypes.NewAnyWithValue(pk)
	suite.Require().NoError(err)

	// Create validator
	res, err := suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgCreateEVMValidator{
			Description:       stakingtypes.NewDescription("Test Validator", "test-identity", "https://test-validator.com", "security@test-validator.com", "Test validator for multistaking"),
			Commission:        stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(1, 2), math.LegacyNewDecWithPrec(2, 1), math.LegacyNewDecWithPrec(1, 2)),
			MinSelfDelegation: math.NewInt(1),
			ValidatorAddress:  valAddr,
			Pubkey:            pkAny,
			ContractAddress:   contractAddr.Hex(),
			Value:             math.NewInt(1_000_000),
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "register ERC20 should have succeeded", res.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
}

func (suite *EVMTestSuite) delegateEVM(contractAddr common.Address, senderPriv cryptotypes.PrivKey, delAddr, valAddr string) {
	// delegateAmount := math.NewInt(500_000) // 500K tokens
	delegateResponse, err := suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgDelegateEVM{
			DelegatorAddress: delAddr,
			ValidatorAddress: valAddr,
			ContractAddress:  contractAddr.Hex(),
			Amount:           math.NewInt(delegateAmount),
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(delegateResponse.IsOK(), "delegate should have succeeded", delegateResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
}

func (suite *EVMTestSuite) redelegateEVM(contractAddr common.Address, senderPriv cryptotypes.PrivKey, delAddr, oldValAddr, newValAddr string) {
	reDelegateResponse, err := suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgBeginRedelegateEVM{
			DelegatorAddress:    delAddr,
			ValidatorSrcAddress: oldValAddr,
			ValidatorDstAddress: newValAddr,
			ContractAddress:     contractAddr.Hex(),
			Amount:              math.NewInt(redelegateAmount),
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(reDelegateResponse.IsOK(), "redelegate should have succeeded", reDelegateResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
}

func (suite *EVMTestSuite) undelegateEVM(contractAddr common.Address, senderPriv cryptotypes.PrivKey, delAddr, valAddr string) {
	undelResponse, err := suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgUndelegateEVM{
			DelegatorAddress: delAddr,
			ValidatorAddress: valAddr,
			ContractAddress:  contractAddr.Hex(),
			Amount:           math.NewInt(undelegateAmount),
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(undelResponse.IsOK(), "undelegate should have succeeded", undelResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
}

func (suite *EVMTestSuite) cancelUndelegateEvm(contractAddr common.Address, senderPriv cryptotypes.PrivKey, delAddr, valAddr string, creationHeight int64) {
	cancelResponse, err := suite.factory.ExecuteCosmosTx(senderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgCancelUnbondingEVMDelegation{
			DelegatorAddress: delAddr,
			ValidatorAddress: valAddr,
			ContractAddress:  contractAddr.Hex(),
			Amount:           math.NewInt(undelegateAmount),
			CreationHeight:   creationHeight,
		}},
	})
	suite.Require().NoError(err)
	suite.Require().True(cancelResponse.IsOK(), "cancelUnbondingDelegation should have succeeded", cancelResponse.GetLog())
	suite.Require().NoError(suite.network.NextBlock())
}
