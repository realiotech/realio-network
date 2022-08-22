package keeper_test

import (
	"testing"

	testkeeper "github.com/realiotech/realio-network/testutil/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.AssetKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
