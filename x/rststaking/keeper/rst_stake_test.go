package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/realiotech/realio-network/testutil/keeper"
	"github.com/realiotech/realio-network/x/rststaking/keeper"
	"github.com/realiotech/realio-network/x/rststaking/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNRstStake(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.RstStake {
	items := make([]types.RstStake, n)
	for i := range items {
		items[i].Index = strconv.Itoa(i)

		keeper.SetRstStake(ctx, items[i])
	}
	return items
}

func TestRstStakeGet(t *testing.T) {
	keeper, ctx := keepertest.RststakingKeeper(t)
	items := createNRstStake(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetRstStake(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t, item, rst)
	}
}
func TestRstStakeRemove(t *testing.T) {
	keeper, ctx := keepertest.RststakingKeeper(t)
	items := createNRstStake(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveRstStake(ctx,
			item.Index,
		)
		_, found := keeper.GetRstStake(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestRstStakeGetAll(t *testing.T) {
	keeper, ctx := keepertest.RststakingKeeper(t)
	items := createNRstStake(keeper, ctx, 10)
	require.ElementsMatch(t, items, keeper.GetAllRstStake(ctx))
}
