package integration

import (
	"encoding/base64"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/evm/testutil/integration/os/factory"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/evm/contracts"
	commonfactory "github.com/cosmos/evm/testutil/integration/common/factory"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/evm/precompiles/testutil"
	"github.com/cosmos/evm/testutil/integration/os/grpc"
	precompileMultiStaking "github.com/realiotech/realio-network/precompile/multistaking"
)

var (
	mintAmount             int64 = 10_000_000
	delegateAmount         int64 = 5_000_000
	redelegateAmount       int64 = 2_000_000
	undelegateAmount       int64 = 200_000
	multistakingBondWeight       = "11.11" // 1:1 bond weight
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

func (suite *EVMTestSuite) TestMultistakingPrecompiles() {
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
	val1Out := suite.createEVMValidatorByPrecompile(contractAddr, senderPriv, senderKey.AccAddr)
	suite.Require().Equal(coinsInf.Infos[2].Denom, val1Out.Validator.BondDenom)

	// Delegate tokens to the 1st validator
	suite.delegateEVMByPrecompile(contractAddr, delPriv, delKey.AccAddr, val1Out.Validator.OperatorAddress)

	// Create a new validator then BeginRedelegate

	// Create second validator
	val2Out := suite.createEVMValidatorByPrecompile(contractAddr, val2Priv, val2Key.AccAddr)
	suite.Require().Equal(coinsInf.Infos[2].Denom, val2Out.Validator.BondDenom)

	// Redelegate tokens from first validator to second validator
	suite.redelegateEVMByPrecompile(contractAddr, delPriv, delKey.AccAddr, val1Out.Validator.OperatorAddress, val2Out.Validator.OperatorAddress)

	// Unbond token from the validator 2
	unDel := suite.undelegateEVMByPrecompile(contractAddr, delPriv, delKey.AccAddr, val2Out.Validator.OperatorAddress)

	// CancelUnbondingDelegation
	suite.cancelUndelegateEvmByPrecompile(contractAddr, delPriv, delKey.AccAddr, val2Out.Validator.OperatorAddress, unDel.UnbondingDelegation.Entries[0].CreationHeight)

	// Undelegate again and wait for completion
	suite.undelegateEVMByPrecompile(contractAddr, delPriv, delKey.AccAddr, val2Out.Validator.OperatorAddress)

	paramsRes, err := suite.network.GetStakingClient().Params(suite.network.GetContext(), &stakingtypes.QueryParamsRequest{})
	suite.Require().NoError(err)

	// User should get back their unbond tokens after unbonding period
	expectedBalanceBefore := mintAmount - delegateAmount
	suite.assertContractBalanceOf(contractAddr, delKey.Addr, expectedBalanceBefore)

	suite.Require().NoError(suite.network.NextBlockAfter(paramsRes.Params.UnbondingTime))
	suite.assertContractBalanceOf(contractAddr, delKey.Addr, expectedBalanceBefore+undelegateAmount)
}

func (suite *EVMTestSuite) TestMultistakingRemoveToken() {
	// Deploy ERC20 contract
	val1Priv := suite.keyring.GetPrivKey(0)
	val1Key := suite.keyring.GetKey(0)
	delPriv := suite.keyring.GetPrivKey(1)
	delKey := suite.keyring.GetKey(1)
	val2Priv := suite.keyring.GetPrivKey(2)
	val2Key := suite.keyring.GetKey(2)
	constructorArgs := []interface{}{"StakeToken", "STAKE", uint8(18)}
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract

	var err error
	factoryy := factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	contractAddr, err := factoryy.DeployContract(
		val1Priv,
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
	suite.mintERC20(contractAddr, val1Key.Addr, mintAmount, val1Priv)
	suite.mintERC20(contractAddr, delKey.Addr, mintAmount, val1Priv)
	suite.mintERC20(contractAddr, val2Key.Addr, mintAmount, val1Priv)

	// Add multistaking evm coin proposal
	bondWeightDec, err := math.LegacyNewDecFromStr(multistakingBondWeight)
	suite.Require().NoError(err)
	err = integrationutils.RegisterMultistakingEVMBondDenom(
		integrationutils.UpdateParamsInput{
			Tf:      factoryy,
			Network: suite.network,
			Pk:      val1Priv,
		},
		contractAddr.Hex(),
		bondWeightDec,
		val1Key.AccAddr,
	)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	multistakingClient := suite.network.GetMultistakingClient()
	coinsInf, err := multistakingClient.MultiStakingCoinInfos(suite.network.GetContext(), &multistakingtypes.QueryMultiStakingCoinInfosRequest{})
	suite.Require().NoError(err)

	// Create 1st validator
	validator1Address := sdk.ValAddress(val1Key.AccAddr).String()
	suite.createEVMValidator(contractAddr, val1Priv, validator1Address)

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

	// Create max undelegations
	paramsRes, err := suite.network.GetStakingClient().Params(suite.network.GetContext(), &stakingtypes.QueryParamsRequest{})
	suite.Require().NoError(err)
	maxEntry := paramsRes.Params.MaxEntries

	for i := 0; i < int(maxEntry); i++ {
		undelResponse, err := suite.factory.ExecuteCosmosTx(delPriv, commonfactory.CosmosTxArgs{
			Msgs: []sdk.Msg{&multistakingtypes.MsgUndelegateEVM{
				DelegatorAddress: delKey.AccAddr.String(),
				ValidatorAddress: validator2Address,
				ContractAddress:  contractAddr.Hex(),
				Amount:           math.NewInt(undelegateAmount),
			}},
		})
		suite.Require().NoError(err)
		suite.Require().True(undelResponse.IsOK(), "undelegate should have succeeded", undelResponse.GetLog())
		suite.Require().NoError(suite.network.NextBlock())
	}

	// Should be faild undelegate
	_, err = suite.factory.ExecuteCosmosTx(delPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgUndelegateEVM{
			DelegatorAddress: delKey.AccAddr.String(),
			ValidatorAddress: validator2Address,
			ContractAddress:  contractAddr.Hex(),
			Amount:           math.NewInt(undelegateAmount),
		}},
	})
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "too many unbonding delegation entries")
	suite.Require().NoError(suite.network.NextBlock())

	// Try to remove multistaking token
	err = integrationutils.RemoveMultistakingBondDenom(
		integrationutils.UpdateParamsInput{
			Tf:      factoryy,
			Network: suite.network,
			Pk:      val1Priv,
		},
		coinsInf.Infos[2].Denom,
		val1Key.AccAddr,
	)
	suite.Require().NoError(err)

	// Make sure denom was remove from multistaking coin list
	coinsInf, err = multistakingClient.MultiStakingCoinInfos(suite.network.GetContext(), &multistakingtypes.QueryMultiStakingCoinInfosRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(len(coinsInf.Infos), 2)

	// Make sure max entry was increased
	paramsRes, err = suite.network.GetStakingClient().Params(suite.network.GetContext(), &stakingtypes.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(paramsRes.Params.MaxEntries, maxEntry+1)

	// Get unbonding delegation

	ubdRes1, err := suite.network.GetStakingClient().ValidatorUnbondingDelegations(suite.network.GetContext(), &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
		ValidatorAddr: validator1Address,
	})
	suite.Require().NoError(err)
	// should have 2 undels: 1 from validator 1 self delegation, 1 from delKey => validator 1
	suite.Require().Equal(len(ubdRes1.UnbondingResponses), 2)

	ubdRes2, err := suite.network.GetStakingClient().ValidatorUnbondingDelegations(suite.network.GetContext(), &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
		ValidatorAddr: validator2Address,
	})
	suite.Require().NoError(err)
	// should have 2 undels: 1 from validator 2 self delegation, 1 from delKey => validator 2
	suite.Require().Equal(len(ubdRes2.UnbondingResponses), 2)
	val1Res, err := suite.network.GetStakingClient().Validator(suite.network.GetContext(), &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validator1Address,
	})

	suite.Require().NoError(err)
	suite.Require().Equal(val1Res.Validator.Jailed, true)
	suite.Require().Equal(val1Res.Validator.Status, stakingtypes.Unbonding)
	suite.Require().Equal(val1Res.Validator.Tokens, math.ZeroInt())
	suite.Require().Equal(val1Res.Validator.DelegatorShares, math.LegacyZeroDec())

	val2Res, err := suite.network.GetStakingClient().Validator(suite.network.GetContext(), &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validator2Address,
	})

	suite.Require().NoError(err)
	suite.Require().Equal(val2Res.Validator.Jailed, true)
	suite.Require().Equal(val2Res.Validator.Status, stakingtypes.Unbonding)
	suite.Require().Equal(val2Res.Validator.Tokens, math.ZeroInt())
	suite.Require().Equal(val2Res.Validator.DelegatorShares, math.LegacyZeroDec())

	// Since we removed bonded token and validator was jailed
	// attacker not cancel these force undelegations
	_, err = suite.factory.ExecuteCosmosTx(delPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{&multistakingtypes.MsgCancelUnbondingEVMDelegation{
			DelegatorAddress: delKey.AccAddr.String(),
			ValidatorAddress: validator2Address,
			ContractAddress:  contractAddr.Hex(),
			Amount:           math.NewInt(undelegateAmount),
			CreationHeight:   ubdRes2.UnbondingResponses[0].Entries[0].CreationHeight,
		}},
	})
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "validator for this address is currently jailed")

	// After unbonding period, all delegated tokens should be returned to users
	suite.Require().NoError(suite.network.NextBlockAfter(paramsRes.Params.UnbondingTime))
	suite.assertContractBalanceOf(contractAddr, delKey.Addr, mintAmount)
	suite.assertContractBalanceOf(contractAddr, val1Key.Addr, mintAmount)
	suite.assertContractBalanceOf(contractAddr, val2Key.Addr, mintAmount)

	// Also validator should be removed after that since all tokens are unbonded
	_, err = suite.network.GetStakingClient().Validator(suite.network.GetContext(), &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validator1Address,
	})
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not found")

	_, err = suite.network.GetStakingClient().Validator(suite.network.GetContext(), &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validator2Address,
	})
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not found")
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

func (suite *EVMTestSuite) createEVMValidatorByPrecompile(contractAddr common.Address, senderPriv cryptotypes.PrivKey, senderAddr sdk.AccAddress) precompileMultiStaking.ValidatorOutput {
	base64Pk := base64.StdEncoding.EncodeToString(ed25519.GenPrivKey().PubKey().(*ed25519.PubKey).Bytes())
	abi, err := precompileMultiStaking.LoadABI()
	suite.Require().NoError(err)
	multistakingPrecompileAddr := common.HexToAddress(precompileMultiStaking.MultistakingPrecompileAddress)

	// Create validator
	res, err := suite.factory.ExecuteContractCall(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "createValidator",
			Args: []interface{}{
				base64Pk,           // pubkey base64 format
				contractAddr.Hex(), // erc20 contract address
				"1000000",          // amount
				"moniker",
				"identity",
				"website",
				"security",
				"details",
				"0.1",  // commission-rate
				"0.2",  // commission-max-rate
				"0.01", // commission-max-change-rate
				"1",    // min-self-delegation
			},
		},
	)

	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "create ERC20 validator should have succeeded", res.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify validator was created
	res, balanceRes, err := suite.factory.CallContractAndCheckLogs(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "validator",
			Args: []interface{}{
				common.BytesToAddress(senderAddr.Bytes()),
			},
		},
		testutil.LogCheckArgs{ExpPass: true},
	)
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "create ERC20 validator should have succeeded", res.GetLog())

	var val precompileMultiStaking.ValidatorOutput
	err = abi.UnpackIntoInterface(&val, "validator", balanceRes.Ret)

	suite.Require().NoError(err)
	suite.Require().NotEqual(val.Validator, precompileMultiStaking.ValidatorInfo{})
	suite.Require().NoError(suite.network.NextBlock())

	return val
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

func (suite *EVMTestSuite) delegateEVMByPrecompile(contractAddr common.Address, senderPriv cryptotypes.PrivKey, senderAddr sdk.AccAddress, valAddr string) {
	abi, err := precompileMultiStaking.LoadABI()
	suite.Require().NoError(err)
	multistakingPrecompileAddr := common.HexToAddress(precompileMultiStaking.MultistakingPrecompileAddress)

	// Create delegation
	res, err := suite.factory.ExecuteContractCall(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "delegate",
			Args: []interface{}{
				contractAddr.Hex(),
				valAddr,
				math.NewInt(delegateAmount).String(),
			},
		},
	)

	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "create ERC20 delegation should have succeeded", res.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify delegation was created

	_, delRes, err := suite.factory.CallContractAndCheckLogs(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "delegation",
			Args: []interface{}{
				common.BytesToAddress(senderAddr.Bytes()),
				valAddr,
			},
		},
		testutil.LogCheckArgs{ExpPass: true},
	)
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "create ERC20 delegation should have succeeded", res.GetLog())

	var del precompileMultiStaking.DelegationOutput
	err = abi.UnpackIntoInterface(&del, "delegation", delRes.Ret)

	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())
	suite.Require().Equal(del.Balance.Amount, math.NewInt(delegateAmount).BigInt())
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

func (suite *EVMTestSuite) redelegateEVMByPrecompile(contractAddr common.Address, senderPriv cryptotypes.PrivKey, delAddr sdk.AccAddress, oldValAddr, newValAddr string) {
	abi, err := precompileMultiStaking.LoadABI()
	suite.Require().NoError(err)
	multistakingPrecompileAddr := common.HexToAddress(precompileMultiStaking.MultistakingPrecompileAddress)

	// Create redelegation
	res, err := suite.factory.ExecuteContractCall(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "redelegate",
			Args: []interface{}{
				contractAddr.Hex(),
				oldValAddr,
				newValAddr,
				math.NewInt(redelegateAmount).String(),
			},
		},
	)

	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "create ERC20 redelegation should have succeeded", res.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify new delegation was created
	_, delRes, err := suite.factory.CallContractAndCheckLogs(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "delegation",
			Args: []interface{}{
				common.BytesToAddress(delAddr.Bytes()),
				newValAddr,
			},
		},
		testutil.LogCheckArgs{ExpPass: true},
	)
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "create ERC20 delegation should have succeeded", res.GetLog())

	var del precompileMultiStaking.DelegationOutput
	err = abi.UnpackIntoInterface(&del, "delegation", delRes.Ret)

	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())
	suite.Require().Equal(del.Balance.Amount, math.NewInt(redelegateAmount).BigInt())
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

func (suite *EVMTestSuite) undelegateEVMByPrecompile(contractAddr common.Address, senderPriv cryptotypes.PrivKey, delAddr sdk.AccAddress, valAddr string) precompileMultiStaking.UnbondingDelegationOutput {
	abi, err := precompileMultiStaking.LoadABI()
	suite.Require().NoError(err)
	multistakingPrecompileAddr := common.HexToAddress(precompileMultiStaking.MultistakingPrecompileAddress)

	// Create redelegation
	res, err := suite.factory.ExecuteContractCall(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "undelegate",
			Args: []interface{}{
				contractAddr.Hex(),
				valAddr,
				math.NewInt(undelegateAmount).String(),
			},
		},
	)

	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "create ERC20 undelegation should have succeeded", res.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify new undelegation was created
	_, delRes, err := suite.factory.CallContractAndCheckLogs(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "unbondingDelegation",
			Args: []interface{}{
				common.BytesToAddress(delAddr.Bytes()),
				valAddr,
			},
		},
		testutil.LogCheckArgs{ExpPass: true},
	)
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "create ERC20 delegation should have succeeded", res.GetLog())

	var unDel precompileMultiStaking.UnbondingDelegationOutput
	err = abi.UnpackIntoInterface(&unDel, "unbondingDelegation", delRes.Ret)
	suite.Require().NoError(err)
	suite.Require().Equal(unDel.UnbondingDelegation.Entries[0].Balance, math.NewInt(undelegateAmount).BigInt())

	suite.Require().NoError(suite.network.NextBlock())

	return unDel
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

func (suite *EVMTestSuite) cancelUndelegateEvmByPrecompile(contractAddr common.Address, senderPriv cryptotypes.PrivKey, delAddr sdk.AccAddress, valAddr string, creationHeight int64) {
	abi, err := precompileMultiStaking.LoadABI()
	suite.Require().NoError(err)
	multistakingPrecompileAddr := common.HexToAddress(precompileMultiStaking.MultistakingPrecompileAddress)

	// Create redelegation
	res, err := suite.factory.ExecuteContractCall(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "cancelUnbondingDelegation",
			Args: []interface{}{
				contractAddr.Hex(),
				valAddr,
				math.NewInt(undelegateAmount).String(),
				math.NewInt(creationHeight).String(),
			},
		},
	)
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "cancel ERC20 undelegation should have succeeded", res.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	// Verify no undelegation
	_, unDelRes, err := suite.factory.CallContractAndCheckLogs(
		senderPriv,
		evmtypes.EvmTxArgs{
			To: &multistakingPrecompileAddr,
		},
		factory.CallArgs{
			ContractABI: abi,
			MethodName:  "unbondingDelegation",
			Args: []interface{}{
				common.BytesToAddress(delAddr.Bytes()),
				valAddr,
			},
		},
		testutil.LogCheckArgs{ExpPass: true},
	)
	suite.Require().NoError(err)
	suite.Require().True(res.IsOK(), "create ERC20 delegation should have succeeded", res.GetLog())

	var unDel precompileMultiStaking.UnbondingDelegationOutput
	err = abi.UnpackIntoInterface(&unDel, "unbondingDelegation", unDelRes.Ret)
	suite.Require().NoError(err)
	suite.Require().Equal(unDel.UnbondingDelegation.DelegatorAddress, "")
	suite.Require().Equal(unDel.UnbondingDelegation.ValidatorAddress, "")
	suite.Require().Equal(len(unDel.UnbondingDelegation.Entries), 0)

	suite.Require().NoError(suite.network.NextBlock())
}
