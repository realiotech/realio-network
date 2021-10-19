package asset_test

import (
	"testing"

	keepertest "github.com/realiotech/network/testutil/keeper"
	"github.com/realiotech/network/x/asset"
	"github.com/realiotech/network/x/asset/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		PortId: types.PortID,
		TokenList: []types.Token{
			{
				Index: "0",
			},
			{
				Index: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.AssetKeeper(t)
	asset.InitGenesis(ctx, *k, genesisState)
	got := asset.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	require.Equal(t, genesisState.PortId, got.PortId)
	require.Len(t, got.TokenList, len(genesisState.TokenList))
	require.Subset(t, genesisState.TokenList, got.TokenList)
	// this line is used by starport scaffolding # genesis/test/assert
}
