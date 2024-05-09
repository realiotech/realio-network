package mint_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/x/mint/types"
	"github.com/stretchr/testify/require"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	realio := app.Setup(false, nil)
	ctx := realio.BaseApp.NewContext(false, tmproto.Header{})
	acc := realio.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
