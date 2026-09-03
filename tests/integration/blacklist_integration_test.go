package integration

import (
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/evm/contracts"
	testconstants "github.com/cosmos/evm/testutil/constants"
	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/testutil/integration/network"
	blacklistmoduletypes "github.com/realiotech/realio-network/x/blacklist/types"
)

const (
	recoverGenesisPath = "../../recover_genesis.json"
	leakedDSTRXHolder  = "realio1pp6a83mfyzyza0kxgsual66zxvqu598mpau406"
)

var dstrxContract = common.HexToAddress("0xb841F365D5221Bed66d60E69094418D8C2aa5A44")

// TestBlacklistDecoratorEthTx verifies the ante-handler blacklist — backed
// by the x/blacklist module's on-chain state, seeded here via genesis — also
// rejects raw Ethereum-style transactions (MsgEthereumTx), not just plain
// Cosmos SDK messages. It submits a real signed tx end to end through a
// properly EVM-wired network (not a direct keeper/msgServer call).
func TestBlacklistDecoratorEthTx(t *testing.T) {
	keyring := testkeyring.New(3)
	blacklistedKey := keyring.GetKey(0)
	cleanKey := keyring.GetKey(1)
	recipientKey := keyring.GetKey(2)

	evmtypes.NewEVMConfigurator().ResetTestConfig()
	integrationNetwork := network.New(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(network.CustomGenesisState{
			blacklistmoduletypes.ModuleName: &blacklistmoduletypes.GenesisState{
				Addresses: []string{blacklistedKey.AccAddr.String()},
			},
		}),
	)

	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	// A blacklisted key's tx must be rejected before it touches state.
	_, err := txFactory.ExecuteEthTx(blacklistedKey.Priv, evmtypes.EvmTxArgs{
		To:     &cleanKey.Addr,
		Amount: nil,
	})
	require.Error(t, err, "expected the blacklisted address's tx to be rejected")
	require.Contains(t, err.Error(), "blacklisted")

	// A non-blacklisted key's tx must NOT just be "not rejected" — it must
	// actually go through and move real value, end to end.
	sendAmount := big.NewInt(1000)
	balBefore, err := grpcHandler.GetBalanceFromEVM(recipientKey.AccAddr)
	require.NoError(t, err)

	res, err := txFactory.ExecuteEthTx(cleanKey.Priv, evmtypes.EvmTxArgs{
		To:     &recipientKey.Addr,
		Amount: sendAmount,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(0), res.Code)
	require.NoError(t, integrationNetwork.NextBlock())

	balAfter, err := grpcHandler.GetBalanceFromEVM(recipientKey.AccAddr)
	require.NoError(t, err)

	before, ok := new(big.Int).SetString(balBefore.Balance, 10)
	require.True(t, ok)
	after, ok := new(big.Int).SetString(balAfter.Balance, 10)
	require.True(t, ok)
	require.Equal(t, new(big.Int).Add(before, sendAmount), after,
		"expected the non-blacklisted sender's transfer to actually move value, not just get accepted")
}

// TestEVMTokenBlacklistHookTransferFromRecoverGenesis imports the deployed
// DSTRX contract and a real leaked holder from the recovery snapshot, then
// executes the stale-allowance attack with a deterministic standard ERC-20 in
// that recovered EVM environment. The clean spender is a legitimate tx signer,
// so ante cannot reject it; the post-tx hook must detect the Transfer log and
// revert every tentative state change made by transferFrom.
func TestEVMTokenBlacklistHookTransferFromRecoverGenesis(t *testing.T) {
	raw, err := os.ReadFile(recoverGenesisPath)
	if os.IsNotExist(err) {
		t.Skipf("recovery genesis fixture not present at %s", recoverGenesisPath)
	}
	require.NoError(t, err)

	var recovered struct {
		AppState map[string]json.RawMessage `json:"app_state"`
	}
	require.NoError(t, json.Unmarshal(raw, &recovered))

	var evmGenesis evmtypes.GenesisState
	encodingConfig := app.MakeEncodingConfig(app.MainnetEVMChainID)
	encodingConfig.Codec.MustUnmarshalJSON(recovered.AppState[evmtypes.ModuleName], &evmGenesis)

	leakedOwnerAcc, err := sdk.AccAddressFromBech32(leakedDSTRXHolder)
	require.NoError(t, err)
	leakedOwner := common.BytesToAddress(leakedOwnerAcc.Bytes())

	keys := testkeyring.New(3)
	spender := keys.GetKey(0)
	owner := keys.GetKey(1)
	recipient := keys.GetKey(2)
	transferAmount := big.NewInt(1_000)
	prefundedAccounts := keys.GetAllAccAddrs()

	foundDSTRX := false
	for i := range evmGenesis.Accounts {
		account := &evmGenesis.Accounts[i]
		if !strings.EqualFold(account.Address, dstrxContract.Hex()) {
			continue
		}

		foundDSTRX = true
		prefundedAccounts = append(prefundedAccounts, sdk.AccAddress(dstrxContract.Bytes()))
		require.NotEmpty(t, account.Code, "recovery genesis must contain deployed DSTRX bytecode")
		// This regression only needs DSTRX. Keeping the other recovered
		// contracts would also require importing their matching x/auth module
		// accounts, which are unrelated to the transferFrom behavior under test.
		evmGenesis.Accounts = []evmtypes.GenesisAccount{*account}
		break
	}
	require.True(t, foundDSTRX, "DSTRX contract missing from recovery genesis")
	// Contract creation is permissioned in the recovery snapshot. Authorize
	// the clean test wallet to deploy the deterministic ERC-20 used below.
	evmGenesis.Params.AccessControl.Create.AccessControlList = append(
		evmGenesis.Params.AccessControl.Create.AccessControlList,
		spender.Addr.Hex(),
	)

	evmtypes.NewEVMConfigurator().ResetTestConfig()
	integrationNetwork := network.New(
		// x/evm requires the recovered contract to have a matching auth
		// account at InitGenesis. Funding it is harmless for this token test.
		network.WithPreFundedAccounts(prefundedAccounts...),
		network.WithChainID(testconstants.ChainID{
			ChainID:    "realionetwork_3301-1",
			EVMChainID: app.MainnetEVMChainID,
		}),
		network.WithCustomGenesis(network.CustomGenesisState{
			evmtypes.ModuleName: &evmGenesis,
			blacklistmoduletypes.ModuleName: &blacklistmoduletypes.GenesisState{
				Addresses: []string{leakedDSTRXHolder},
			},
		}),
	)

	evmKeeper := integrationNetwork.GetApp().EvmKeeper
	ctx := integrationNetwork.GetContext()
	require.NotEmpty(t, evmKeeper.GetCode(ctx, evmKeeper.GetCodeHash(ctx, dstrxContract)))
	require.True(t, integrationNetwork.GetApp().BlacklistKeeper.IsBlacklisted(ctx, leakedOwnerAcc))
	recoveredOwnerBalance := evmKeeper.GetState(ctx, dstrxContract, solidityMappingKey(leakedOwner, 0))
	require.Positive(t, recoveredOwnerBalance.Big().Sign(),
		"recovery genesis must retain the leaked holder's nonzero DSTRX position")

	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)
	tokenAddress, err := txFactory.DeployContract(spender.Priv, evmtypes.EvmTxArgs{
		GasLimit: 5_000_000,
	}, testutiltypes.ContractDeploymentData{
		Contract:        contracts.ERC20MinterBurnerDecimalsContract,
		ConstructorArgs: []interface{}{"Blacklist Regression Token", "BRT", uint8(18)},
	})
	require.NoError(t, err)
	require.NoError(t, integrationNetwork.NextBlock())

	_, err = txFactory.ExecuteContractCall(spender.Priv, evmtypes.EvmTxArgs{
		To:       &tokenAddress,
		GasLimit: 2_000_000,
	}, testutiltypes.CallArgs{
		ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{owner.Addr, transferAmount},
	})
	require.NoError(t, err)
	require.NoError(t, integrationNetwork.NextBlock())

	_, err = txFactory.ExecuteContractCall(owner.Priv, evmtypes.EvmTxArgs{
		To:       &tokenAddress,
		GasLimit: 2_000_000,
	}, testutiltypes.CallArgs{
		ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
		MethodName:  "approve",
		Args:        []interface{}{spender.Addr, transferAmount},
	})
	require.NoError(t, err)
	require.NoError(t, integrationNetwork.NextBlock())

	// Export the initialized token after mint+approve, then restart from a
	// genesis where the owner is already blacklisted. This models the recovery
	// binary starting with a pre-existing allowance; no blacklisted owner has to
	// sign an approval transaction after the protection is active.
	ctx = integrationNetwork.GetContext()
	tokenCodeHash := evmKeeper.GetCodeHash(ctx, tokenAddress)
	tokenGenesisAccount := evmtypes.GenesisAccount{
		Address: tokenAddress.Hex(),
		Code:    common.Bytes2Hex(evmKeeper.GetCode(ctx, tokenCodeHash)),
		Storage: evmKeeper.GetAccountStorage(ctx, tokenAddress),
	}
	require.NotEmpty(t, tokenGenesisAccount.Code)
	require.NotEmpty(t, tokenGenesisAccount.Storage)
	evmGenesis.Accounts = append(evmGenesis.Accounts, tokenGenesisAccount)
	prefundedAccounts = append(prefundedAccounts, sdk.AccAddress(tokenAddress.Bytes()))

	evmtypes.NewEVMConfigurator().ResetTestConfig()
	integrationNetwork = network.New(
		network.WithPreFundedAccounts(prefundedAccounts...),
		network.WithChainID(testconstants.ChainID{
			ChainID:    "realionetwork_3301-1",
			EVMChainID: app.MainnetEVMChainID,
		}),
		network.WithCustomGenesis(network.CustomGenesisState{
			evmtypes.ModuleName: &evmGenesis,
			blacklistmoduletypes.ModuleName: &blacklistmoduletypes.GenesisState{
				Addresses: []string{leakedDSTRXHolder, owner.AccAddr.String()},
			},
		}),
	)
	grpcHandler = grpc.NewIntegrationHandler(integrationNetwork)
	txFactory = factory.New(integrationNetwork, grpcHandler)
	require.True(t, integrationNetwork.GetApp().EvmKeeper.HasHooks())
	require.True(t, integrationNetwork.GetApp().BlacklistKeeper.IsBlacklisted(
		integrationNetwork.GetContext(), owner.AccAddr,
	))

	queryTokenUint := func(method string, args ...interface{}) *big.Int {
		t.Helper()
		response, queryErr := txFactory.QueryContract(evmtypes.EvmTxArgs{
			To: &tokenAddress,
		}, testutiltypes.CallArgs{
			ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
			MethodName:  method,
			Args:        args,
		}, 300_000)
		require.NoError(t, queryErr)
		return new(big.Int).SetBytes(response.Ret)
	}

	ownerBalanceBefore := queryTokenUint("balanceOf", owner.Addr)
	allowanceBefore := queryTokenUint("allowance", owner.Addr, spender.Addr)
	recipientBalanceBefore := queryTokenUint("balanceOf", recipient.Addr)
	require.Equal(t, transferAmount, ownerBalanceBefore)
	require.Equal(t, transferAmount, allowanceBefore)

	callArgs := testutiltypes.CallArgs{
		ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
		MethodName:  "transferFrom",
		Args:        []interface{}{owner.Addr, recipient.Addr, transferAmount},
	}
	res, err := txFactory.ExecuteContractCall(spender.Priv, evmtypes.EvmTxArgs{
		To:       &tokenAddress,
		GasLimit: 2_000_000,
	}, callArgs)
	// The integration helper only turns ABI-encoded contract reverts into Go
	// errors. A post-tx hook failure can therefore return nil here; VmError is
	// the authoritative EVM execution status asserted below.
	if err != nil {
		require.Contains(t, err.Error(), "blacklisted")
	}
	require.True(t, res.IsOK(), "EVM reverts are encoded in MsgEthereumTxResponse, not the ABCI code")
	ethRes, decodeErr := evmtypes.DecodeTxResponse(res.Data)
	require.NoError(t, decodeErr)
	require.Contains(t, ethRes.VmError, "blacklisted")
	require.Empty(t, ethRes.Logs, "reverted EVM logs must not survive post processing")

	// Read from the finalized block state. All three values must match their
	// pre-transaction values, including the allowance that transferFrom tried
	// to decrement.
	require.NoError(t, integrationNetwork.NextBlock())
	require.Equal(t, ownerBalanceBefore, queryTokenUint("balanceOf", owner.Addr))
	require.Equal(t, recipientBalanceBefore, queryTokenUint("balanceOf", recipient.Addr))
	require.Equal(t, allowanceBefore, queryTokenUint("allowance", owner.Addr, spender.Addr))
}

func solidityMappingKey(address common.Address, slot uint64) common.Hash {
	return crypto.Keccak256Hash(
		common.LeftPadBytes(address.Bytes(), common.HashLength),
		common.LeftPadBytes(new(big.Int).SetUint64(slot).Bytes(), common.HashLength),
	)
}
