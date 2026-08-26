package integration

import (
	"cosmossdk.io/math"

	bip39 "github.com/cosmos/go-bip39"

	"github.com/cosmos/evm/contracts"
	ethermintHD "github.com/cosmos/evm/crypto/hd"
	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
)

// rotatedHDPath is the same BIP44 path used by realio-networkd's default
// `keys add` (coin-type 60, eth_secp256k1) — confirmed earlier to match the
// chain binary's own key derivation for the address-rotation exercise.
const rotatedHDPath = "m/44'/60'/0'/0/0"

// generateRotatedKey derives a brand-new eth_secp256k1 key the exact same
// way the address-rotation script does (BIP39 mnemonic -> coin-type 60 HD
// path -> eth_secp256k1), entirely in-process. Unlike reading a real key
// from the rotation's mnemonic export, this needs no external file and no
// private key material to ever leave this test run, so it works identically
// for anyone who checks out the repo.
func generateRotatedKey(suite *EVMTestSuite) testkeyring.Key {
	entropy, err := bip39.NewEntropy(256) // 256 bits -> 24-word mnemonic
	suite.Require().NoError(err)
	mnemonic, err := bip39.NewMnemonic(entropy)
	suite.Require().NoError(err)

	derivedKey, err := ethermintHD.EthSecp256k1.Derive()(mnemonic, "", rotatedHDPath)
	suite.Require().NoError(err)
	privKey := ethermintHD.EthSecp256k1.Generate()(derivedKey)

	pub := privKey.PubKey()
	accAddr := sdk.AccAddress(pub.Address().Bytes())
	ethAddr := common.BytesToAddress(pub.Address().Bytes())

	return testkeyring.Key{Addr: ethAddr, AccAddr: accAddr, Priv: privKey}
}

// TestMultistakingPrecompilesRotatedKey mirrors TestMultistakingPrecompiles
// (same erc20-backed multistaking delegate/undelegate/maturity flow), but
// the delegator is a key derived through the exact same BIP39/HD scheme the
// address-rotation script uses, instead of the suite's default throwaway
// keyring account — using the shared EVMTestSuite's properly EVM-wired
// network (unlike the minimal SetupWithGenFile harness in
// app/rotation_test.go, where this same maturity step panicked with "EVM
// call unexpected error").
func (suite *EVMTestSuite) TestMultistakingPrecompilesRotatedKey() {
	rotatedKey := generateRotatedKey(suite)

	senderPriv := suite.keyring.GetPrivKey(0)
	senderKey := suite.keyring.GetKey(0)
	val2Priv := suite.keyring.GetPrivKey(2)
	val2Key := suite.keyring.GetKey(2)

	// Give the rotated account native gas funds — on a real chain it would
	// already hold "ario" from genesis; here it needs SOME balance to pay
	// for its own delegate/undelegate txs.
	err := suite.factory.FundAccount(senderKey, rotatedKey.AccAddr, sdk.NewCoins(sdk.NewCoin(suite.network.GetBaseDenom(), math.NewInt(1_000_000_000_000_000_000))))
	suite.Require().NoError(err)
	suite.Require().NoError(suite.network.NextBlock())

	// Deploy ERC20 contract
	constructorArgs := []interface{}{"StakeToken", "STAKE", uint8(18)}
	compiledContract := contracts.ERC20MinterBurnerDecimalsContract

	factoryy := factory.New(suite.network, grpc.NewIntegrationHandler(suite.network))
	contractAddr, err := factoryy.DeployContract(
		senderPriv,
		evmtypes.EvmTxArgs{},
		testutiltypes.ContractDeploymentData{
			Contract:        compiledContract,
			ConstructorArgs: constructorArgs,
		},
	)
	suite.Require().NoError(err)
	suite.NotEqual(contractAddr, common.Address{})
	suite.Require().NoError(suite.network.NextBlock())

	// Mint ERC20 tokens to the rotated address and the second validator.
	suite.mintERC20(contractAddr, rotatedKey.Addr, mintAmount, senderPriv)
	suite.mintERC20(contractAddr, val2Key.Addr, mintAmount, senderPriv)

	// Register the ERC20 contract as a multistaking bond coin.
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

	// Create a validator for the rotated key to delegate to.
	val2Out := suite.createEVMValidatorByPrecompile(contractAddr, val2Priv, val2Key.AccAddr)

	// The ROTATED key delegates.
	suite.delegateEVMByPrecompile(contractAddr, rotatedKey.Priv, rotatedKey.AccAddr, val2Out.Validator.OperatorAddress)

	// The ROTATED key undelegates.
	suite.undelegateEVMByPrecompile(contractAddr, rotatedKey.Priv, rotatedKey.AccAddr, val2Out.Validator.OperatorAddress)

	paramsRes, err := suite.network.GetStakingClient().Params(suite.network.GetContext(), &stakingtypes.QueryParamsRequest{})
	suite.Require().NoError(err)

	expectedBalanceBefore := mintAmount - delegateAmount
	suite.assertContractBalanceOf(contractAddr, rotatedKey.Addr, expectedBalanceBefore)

	// Advance past the full unbonding period, exactly like the maturity
	// check in app/rotation_test.go — except here the EVM/erc20 payout path
	// is exercised for real, with a properly wired network.
	suite.Require().NoError(suite.network.NextBlockAfter(paramsRes.Params.UnbondingTime))
	suite.assertContractBalanceOf(contractAddr, rotatedKey.Addr, expectedBalanceBefore+undelegateAmount)
}
