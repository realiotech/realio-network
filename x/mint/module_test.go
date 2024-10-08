package mint_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/x/mint/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContext(false)
	acc := realio.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
