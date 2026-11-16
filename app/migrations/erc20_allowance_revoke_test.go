package migrations_test

import (
	"encoding/json"
	"math/big"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/app/migrations"
	"github.com/realiotech/realio-network/testutil"
)

func TestRevokeLeakedERC20Allowances(t *testing.T) {
	realio := app.Setup(false, nil, 1)

	origHeight, origJSON, origAssetRotations := migrations.BlacklistForkHeight, migrations.LeakedAddressesJSON, migrations.AssetManagerRotations
	t.Cleanup(func() {
		migrations.BlacklistForkHeight, migrations.LeakedAddressesJSON, migrations.AssetManagerRotations = origHeight, origJSON, origAssetRotations
	})

	token := common.HexToAddress("0x0000000000000000000000000000000000001234")
	pair := erc20types.NewTokenPair(token, "utest", erc20types.OWNER_EXTERNAL)
	leakedOwner := testutil.GenAddress()
	cleanOwner := testutil.GenAddress()
	spender := common.HexToAddress("0x0000000000000000000000000000000000005678")

	migrations.LeakedAddressesJSON, _ = json.Marshal([]string{leakedOwner.String()})
	migrations.AssetManagerRotations = nil
	migrations.BlacklistForkHeight = 12345

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: migrations.BlacklistForkHeight})
	require.NoError(t, realio.Erc20Keeper.SetToken(ctx, pair))
	require.NoError(t, realio.Erc20Keeper.SetAllowance(ctx, token, common.BytesToAddress(leakedOwner), spender, big.NewInt(10)))
	require.NoError(t, realio.Erc20Keeper.SetAllowance(ctx, token, common.BytesToAddress(cleanOwner), spender, big.NewInt(20)))

	realio.ScheduleForkUpgrade(ctx)

	leakedAllowance, err := realio.Erc20Keeper.GetAllowance(ctx, token, common.BytesToAddress(leakedOwner), spender)
	require.NoError(t, err)
	require.Zero(t, leakedAllowance.Sign())
	cleanAllowance, err := realio.Erc20Keeper.GetAllowance(ctx, token, common.BytesToAddress(cleanOwner), spender)
	require.NoError(t, err)
	require.Equal(t, int64(20), cleanAllowance.Int64())
}
