package integration

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/realiotech/realio-network/testutil/integration/network"
	blacklistmoduletypes "github.com/realiotech/realio-network/x/blacklist/types"
)

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
