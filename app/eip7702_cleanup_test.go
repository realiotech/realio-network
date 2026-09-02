package app

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/testutil"
)

func TestClearLeakedEIP7702Delegations(t *testing.T) {
	realio := Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1})
	leaked := testutil.GenAddress()
	clean := testutil.GenAddress()
	target := common.HexToAddress("0x0000000000000000000000000000000000001234")
	delegation := ethtypes.AddressToDelegation(target)
	codeHash := crypto.Keccak256Hash(delegation)
	realio.EvmKeeper.SetCode(ctx, codeHash.Bytes(), delegation)
	realio.EvmKeeper.SetCodeHash(ctx, leaked.Bytes(), codeHash.Bytes())
	realio.EvmKeeper.SetCodeHash(ctx, clean.Bytes(), codeHash.Bytes())

	clearLeakedEIP7702Delegations(realio, ctx, []sdk.AccAddress{leaked})

	_, leakedDelegated := ethtypes.ParseDelegation(realio.EvmKeeper.GetCode(ctx, realio.EvmKeeper.GetCodeHash(ctx, common.BytesToAddress(leaked))))
	require.False(t, leakedDelegated)
	cleanTarget, cleanDelegated := ethtypes.ParseDelegation(realio.EvmKeeper.GetCode(ctx, realio.EvmKeeper.GetCodeHash(ctx, common.BytesToAddress(clean))))
	require.True(t, cleanDelegated)
	require.Equal(t, target, cleanTarget)
}
