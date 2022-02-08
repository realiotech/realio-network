package rststaking_test

import (
	"testing"

	keepertest "github.com/realiotech/realio-network/testutil/keeper"
	"github.com/realiotech/realio-network/x/rststaking"
	"github.com/realiotech/realio-network/x/rststaking/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		RstStakeList: []types.RstStake{
			{
				Index: "0",
			},
			{
				Index: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.RststakingKeeper(t)
	rststaking.InitGenesis(ctx, *k, genesisState)
	got := rststaking.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	require.Len(t, got.RstStakeList, len(genesisState.RstStakeList))
	require.Subset(t, genesisState.RstStakeList, got.RstStakeList)
	// this line is used by starport scaffolding # genesis/test/assert
}
