package integration

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/realiotech/realio-network/testutil/integration/network"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/contracts"
	integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EVMTestSuite struct {
	suite.Suite
	network     network.Network
	grpcHandler grpc.Handler
	factory     factory.TxFactory
	keyring     testkeyring.Keyring
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EVMTestSuite) SetupTest() {
	keyring := testkeyring.New(4)
	integrationNetwork := network.New(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	factory := factory.New(integrationNetwork, grpcHandler)

	suite.grpcHandler = grpcHandler
	suite.factory = factory
	suite.network = integrationNetwork
	suite.keyring = keyring
}

type evmTestCase struct {
	name      string
	getTxArgs func() evmtypes.EvmTxArgs
}

type evmPermissionTestCase struct {
	name                     string
	getTxArgs                func() evmtypes.EvmTxArgs
	updateParams             func() evmtypes.Params
	createParams, callParams PermissionsTableTest
}

type PermissionsTableTest struct {
	ExpFail     bool
	SignerIndex int
}

func (suite *EVMTestSuite) TestNativeTransfers() {
	testCases := []evmTestCase{
		{
			name: "DynamicFeeTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
		},
		{
			name: "AccessListTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					Accesses: &ethtypes.AccessList{{
						Address:     suite.keyring.GetKey(1).Addr,
						StorageKeys: []common.Hash{{0}},
					}},
				}
			},
		},
		{
			name: "LegacyTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					GasPrice: big.NewInt(1e9),
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			senderKey := suite.keyring.GetKey(0)
			receiverKey := suite.keyring.GetKey(1)
			denom := suite.network.GetBaseDenom()

			txArgs := tc.getTxArgs()

			senderPrevBalanceResponse, err := suite.grpcHandler.GetBalanceFromBank(senderKey.AccAddr, denom)
			suite.NoError(err, "Error should not exists")

			senderPrevBalance := senderPrevBalanceResponse.GetBalance().Amount

			receiverPrevBalanceResponse, err := suite.grpcHandler.GetBalanceFromBank(receiverKey.AccAddr, denom)
			suite.Require().NoError(err)
			receiverPrevBalance := receiverPrevBalanceResponse.GetBalance().Amount

			transferAmount := int64(1000)

			txArgs.Amount = big.NewInt(transferAmount)
			txArgs.To = &receiverKey.Addr

			res, err := suite.factory.ExecuteEthTx(senderKey.Priv, txArgs)
			suite.Require().NoError(err)
			suite.Require().True(res.IsOK(), "transaction should have succeeded: %s", res.GetLog())

			err = suite.network.NextBlock()
			suite.Require().NoError(err)

			// Check sender balance after transaction
			senderBalanceResultBeforeFees := senderPrevBalance.Sub(math.NewInt(transferAmount))
			senderAfterBalance, err := suite.grpcHandler.GetBalanceFromBank(senderKey.AccAddr, denom)
			suite.Require().NoError(err)
			suite.Require().True(senderAfterBalance.GetBalance().Amount.LTE(senderBalanceResultBeforeFees))

			// Check receiver balance after transaction
			receiverBalanceResult := receiverPrevBalance.Add(math.NewInt(transferAmount))
			receverAfterBalanceResponse, err := suite.grpcHandler.GetBalanceFromBank(receiverKey.AccAddr, denom)
			suite.Require().NoError(err)
			suite.Require().Equal(receiverBalanceResult, receverAfterBalanceResponse.GetBalance().Amount)
		})
	}
}

func (suite *EVMTestSuite) TestContractDeploymentPermissionless() {
	testCases := []evmTestCase{
		{
			name: "DynamicFeeTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
		},
		{
			name: "AccessListTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					Accesses: &ethtypes.AccessList{{
						Address:     suite.keyring.GetKey(1).Addr,
						StorageKeys: []common.Hash{{0}},
					}},
				}
			},
		},
		{
			name: "LegacyTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					GasPrice: big.NewInt(1e9),
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			senderPriv := suite.keyring.GetPrivKey(0)
			constructorArgs := []interface{}{"coin", "token", uint8(18)}
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract

			txArgs := tc.getTxArgs()
			contractAddr, err := suite.factory.DeployContract(
				senderPriv,
				txArgs,
				factory.ContractDeploymentData{
					Contract:        compiledContract,
					ConstructorArgs: constructorArgs,
				},
			)
			suite.Require().NoError(err)
			suite.NotEqual(contractAddr, common.Address{})

			suite.Require().NoError(suite.network.NextBlock())

			// Check contract account got created correctly
			contractBechAddr := sdk.AccAddress(contractAddr.Bytes()).String()
			contractAccount, err := suite.grpcHandler.GetAccount(contractBechAddr)
			suite.Require().NoError(err)
			suite.Require().NotNil(contractAccount, "expected account to be retrievable via auth query")

			ethAccountRes, err := suite.grpcHandler.GetEvmAccount(contractAddr)
			suite.Require().NoError(err)
			suite.NotEqual(ethAccountRes.CodeHash, common.BytesToHash(evmtypes.EmptyCodeHash).Hex())
		})
	}
}

func (suite *EVMTestSuite) TestContractCall() {
	preDeploy := func() common.Address {
		// Deploy contract
		senderPriv := suite.keyring.GetPrivKey(0)
		constructorArgs := []interface{}{"coin", "token", uint8(18)}
		compiledContract := contracts.ERC20MinterBurnerDecimalsContract

		var err error // Avoid shadowing
		contractAddr, err := suite.factory.DeployContract(
			senderPriv,
			evmtypes.EvmTxArgs{}, // Default values
			factory.ContractDeploymentData{
				Contract:        compiledContract,
				ConstructorArgs: constructorArgs,
			},
		)
		suite.Require().NoError(err)
		suite.NotEqual(contractAddr, common.Address{})

		err = suite.network.NextBlock()
		suite.Require().NoError(err)
		return contractAddr
	}

	testCases := []evmTestCase{
		{
			name: "DynamicFeeTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
		},
		{
			name: "AccessListTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					Accesses: &ethtypes.AccessList{{
						Address:     suite.keyring.GetKey(1).Addr,
						StorageKeys: []common.Hash{{0}},
					}},
				}
			},
		},
		{
			name: "LegacyTx",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{
					GasPrice: big.NewInt(1e9),
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			contractAddr := preDeploy()
			senderPriv := suite.keyring.GetPrivKey(0)
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract
			recipientKey := suite.keyring.GetKey(1)

			// Execute contract call
			mintTxArgs := tc.getTxArgs()
			mintTxArgs.To = &contractAddr

			amountToMint := big.NewInt(1e18)
			mintArgs := factory.CallArgs{
				ContractABI: compiledContract.ABI,
				MethodName:  "mint",
				Args:        []interface{}{recipientKey.Addr, amountToMint},
			}
			mintResponse, err := suite.factory.ExecuteContractCall(senderPriv, mintTxArgs, mintArgs)
			suite.Require().NoError(err)
			suite.Require().True(mintResponse.IsOK(), "transaction should have succeeded", mintResponse.GetLog())

			err = checkMintTopics(mintResponse)
			suite.Require().NoError(err)

			err = suite.network.NextBlock()
			suite.Require().NoError(err)

			totalSupplyTxArgs := evmtypes.EvmTxArgs{
				To: &contractAddr,
			}
			totalSupplyArgs := factory.CallArgs{
				ContractABI: compiledContract.ABI,
				MethodName:  "totalSupply",
				Args:        []interface{}{},
			}
			totalSupplyRes, err := suite.factory.ExecuteContractCall(senderPriv, totalSupplyTxArgs, totalSupplyArgs)
			suite.Require().NoError(err)
			suite.Require().True(mintResponse.IsOK(), "transaction should have succeeded", totalSupplyRes.GetLog())

			var totalSupplyResponse *big.Int
			err = integrationutils.DecodeContractCallResponse(&totalSupplyResponse, totalSupplyArgs, totalSupplyRes)
			suite.Require().NoError(err)
			suite.Require().Equal(totalSupplyResponse, amountToMint)

			suite.Require().NoError(suite.network.NextBlock())
		})
	}
}

func (suite *EVMTestSuite) TestContractDeploymentAndCallWithPermissions() {
	testCases := []evmPermissionTestCase{
		{
			name: "Create and call is successful with create permission policy set to permissionless and address not blocked ",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				blockedSignerIndex := 1
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Create = evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissionless,
					AccessControlList: []string{suite.keyring.GetAddr(blockedSignerIndex).String()},
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: false, SignerIndex: 0},
			callParams:   PermissionsTableTest{ExpFail: false, SignerIndex: 0},
		},
		{
			name: "Create fails with create permission policy set to permissionless and signer is blocked ",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				blockedSignerIndex := 1
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Create = evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissionless,
					AccessControlList: []string{suite.keyring.GetAddr(blockedSignerIndex).String()},
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: true, SignerIndex: 1},
			callParams:   PermissionsTableTest{},
		},
		{
			name: "Create and call is successful with call permission policy set to permissionless and address not blocked ",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				blockedSignerIndex := 1
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Call = evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissionless,
					AccessControlList: []string{suite.keyring.GetAddr(blockedSignerIndex).String()},
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: false, SignerIndex: 0},
			callParams:   PermissionsTableTest{ExpFail: false, SignerIndex: 0},
		},
		{
			name: "Create is successful and call fails with call permission policy set to permissionless and address blocked ",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				blockedSignerIndex := 1
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Call = evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissionless,
					AccessControlList: []string{suite.keyring.GetAddr(blockedSignerIndex).String()},
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: false, SignerIndex: 0},
			callParams:   PermissionsTableTest{ExpFail: true, SignerIndex: 1},
		},
		{
			name: "Create fails create permission policy set to restricted",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Create = evmtypes.AccessControlType{
					AccessType: evmtypes.AccessTypeRestricted,
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: true, SignerIndex: 0},
			callParams:   PermissionsTableTest{},
		},
		{
			name: "Create succeeds and call fails when call permission policy set to restricted",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Call = evmtypes.AccessControlType{
					AccessType: evmtypes.AccessTypeRestricted,
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: false, SignerIndex: 0},
			callParams:   PermissionsTableTest{ExpFail: true, SignerIndex: 0},
		},
		{
			name: "Create and call are successful with create permission policy set to permissioned and signer whitelisted",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				whitelistedSignerIndex := 1
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Create = evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissioned,
					AccessControlList: []string{suite.keyring.GetAddr(whitelistedSignerIndex).String()},
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: false, SignerIndex: 1},
			callParams:   PermissionsTableTest{ExpFail: false, SignerIndex: 0},
		},
		{
			name: "Create fails with create permission policy set to permissioned and signer NOT whitelisted",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				whitelistedSignerIndex := 1
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Create = evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissioned,
					AccessControlList: []string{suite.keyring.GetAddr(whitelistedSignerIndex).String()},
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: true, SignerIndex: 0},
			callParams:   PermissionsTableTest{},
		},
		{
			name: "Create and call are successful with call permission policy set to permissioned and signer whitelisted",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				whitelistedSignerIndex := 1
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Call = evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissioned,
					AccessControlList: []string{suite.keyring.GetAddr(whitelistedSignerIndex).String()},
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: false, SignerIndex: 0},
			callParams:   PermissionsTableTest{ExpFail: false, SignerIndex: 1},
		},
		{
			name: "Create succeeds and call fails with call permission policy set to permissioned and signer NOT whitelisted",
			getTxArgs: func() evmtypes.EvmTxArgs {
				return evmtypes.EvmTxArgs{}
			},
			updateParams: func() evmtypes.Params {
				whitelistedSignerIndex := 1
				// Set params to default values
				defaultParams := evmtypes.DefaultParams()
				defaultParams.AccessControl.Call = evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissioned,
					AccessControlList: []string{suite.keyring.GetAddr(whitelistedSignerIndex).String()},
				}
				return defaultParams
			},
			createParams: PermissionsTableTest{ExpFail: false, SignerIndex: 0},
			callParams:   PermissionsTableTest{ExpFail: true, SignerIndex: 0},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := tc.updateParams()

			err := integrationutils.UpdateEvmParams(
				integrationutils.UpdateParamsInput{
					Tf:      suite.factory,
					Network: suite.network,
					Pk:      suite.keyring.GetPrivKey(0),
					Params:  params,
				},
			)
			suite.Require().NoError(err)
			suite.Require().NoError(suite.network.NextBlock())

			// Deploy contract
			createSigner := suite.keyring.GetPrivKey(tc.createParams.SignerIndex)
			constructorArgs := []interface{}{"coin", "token", uint8(18)}
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract

			contractAddr, err := suite.factory.DeployContract(
				createSigner,
				evmtypes.EvmTxArgs{}, // Default values
				factory.ContractDeploymentData{
					Contract:        compiledContract,
					ConstructorArgs: constructorArgs,
				},
			)
			if tc.createParams.ExpFail {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), "does not have permission to deploy contracts")
				// If contract deployment is expected to fail, we can skip the rest of the test
				return
			}

			suite.Require().NoError(err)
			suite.Require().NotEqual(contractAddr, common.Address{})
			suite.Require().NoError(suite.network.NextBlock())

			callSigner := suite.keyring.GetPrivKey(tc.callParams.SignerIndex)
			totalSupplyTxArgs := evmtypes.EvmTxArgs{
				To: &contractAddr,
			}
			totalSupplyArgs := factory.CallArgs{
				ContractABI: compiledContract.ABI,
				MethodName:  "totalSupply",
				Args:        []interface{}{},
			}
			res, err := suite.factory.ExecuteContractCall(callSigner, totalSupplyTxArgs, totalSupplyArgs)
			if tc.callParams.ExpFail {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), "does not have permission to perform a call")
			} else {
				suite.Require().NoError(err)
				suite.Require().True(res.IsOK())
			}
		})
	}
}

func checkMintTopics(res abcitypes.ExecTxResult) error {
	// Check contract call response has the expected topics for a mint
	// call within an ERC20 contract
	expectedTopics := []string{
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"0x0000000000000000000000000000000000000000000000000000000000000000",
	}
	return integrationutils.CheckTxTopics(res, expectedTopics)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEVMTestSuite(t *testing.T) {
	suite.Run(t, new(EVMTestSuite))
}
