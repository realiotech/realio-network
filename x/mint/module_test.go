package mint_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/x/mint/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContext(false)
	acc := realio.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}

func TestEndBlcoker(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContext(false)
	// For pass feemarket endblock
	ctx = ctx.WithBlockGasMeter(storetypes.NewGasMeter(1000))
	amt := sdk.NewCoins(sdk.NewCoin("ario", math.NewInt(1000)))
	err := realio.BankKeeper.MintCoins(ctx, types.ModuleName, amt)
	require.NoError(t, err)
	err = realio.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, types.EvmDeadAddr, amt)
	require.NoError(t, err)

	_, err = realio.EndBlocker(ctx)
	require.NoError(t, err)

	deadBalances := realio.BankKeeper.GetAllBalances(ctx, types.EvmDeadAddr)
	require.Empty(t, deadBalances)
}
