package app

import (
	"math/big"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/testutil"
)

func TestRevokeLeakedERC20Allowances(t *testing.T) {
	realio := Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1})
	token := common.HexToAddress("0x0000000000000000000000000000000000001234")
	pair := erc20types.NewTokenPair(token, "utest", erc20types.OWNER_EXTERNAL)
	require.NoError(t, realio.Erc20Keeper.SetToken(ctx, pair))

	leakedOwner := testutil.GenAddress()
	cleanOwner := testutil.GenAddress()
	spender := common.HexToAddress("0x0000000000000000000000000000000000005678")
	require.NoError(t, realio.Erc20Keeper.SetAllowance(ctx, token, common.BytesToAddress(leakedOwner), spender, big.NewInt(10)))
	require.NoError(t, realio.Erc20Keeper.SetAllowance(ctx, token, common.BytesToAddress(cleanOwner), spender, big.NewInt(20)))

	revokeLeakedERC20Allowances(realio, ctx, []sdk.AccAddress{leakedOwner})

	leakedAllowance, err := realio.Erc20Keeper.GetAllowance(ctx, token, common.BytesToAddress(leakedOwner), spender)
	require.NoError(t, err)
	require.Zero(t, leakedAllowance.Sign())
	cleanAllowance, err := realio.Erc20Keeper.GetAllowance(ctx, token, common.BytesToAddress(cleanOwner), spender)
	require.NoError(t, err)
	require.Equal(t, int64(20), cleanAllowance.Int64())
}
