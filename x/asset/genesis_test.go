package asset_test

import (
	"testing"

	keepertest "github.com/realiotech/realio-network/testutil/keeper"
	"github.com/realiotech/realio-network/testutil/nullify"
	"github.com/realiotech/realio-network/v1/x/asset"
	"github.com/realiotech/realio-network/v1/x/asset/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.AssetKeeper(t)
	asset.InitGenesis(ctx, *k, genesisState)
	got := asset.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	// this line is used by starport scaffolding # genesis/test/assert
}
