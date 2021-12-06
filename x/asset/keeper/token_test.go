package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/realiotech/realio-network/testutil/keeper"
	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNToken(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Token {
	items := make([]types.Token, n)
	for i := range items {
		items[i].Index = strconv.Itoa(i)

		keeper.SetToken(ctx, items[i])
	}
	return items
}

func TestTokenGet(t *testing.T) {
	keeper, ctx := keepertest.AssetKeeper(t)
	items := createNToken(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetToken(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t, item, rst)
	}
}
func TestTokenRemove(t *testing.T) {
	keeper, ctx := keepertest.AssetKeeper(t)
	items := createNToken(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveToken(ctx,
			item.Index,
		)
		_, found := keeper.GetToken(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestTokenGetAll(t *testing.T) {
	keeper, ctx := keepertest.AssetKeeper(t)
	items := createNToken(keeper, ctx, 10)
	require.ElementsMatch(t, items, keeper.GetAllToken(ctx))
}

func (suite *KeeperTestSuite) TestIsAddressAuthorized() {
	suite.SetupTest()

	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"

	createMsg := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000, AuthorizationRequired: true,
	}
	_, _ = suite.msgSrv.CreateToken(wctx, createMsg)

	suite.Require().False(suite.app.AssetKeeper.IsAddressAuthorizedToSend(suite.ctx, "1", suite.testUser1Acc))

	authUserMsg := &types.MsgAuthorizeAddress{Creator: creator,
		Index: "1", Address: suite.testUser1Address,
	}
	_, _ = suite.msgSrv.AuthorizeAddress(wctx, authUserMsg)

	suite.Require().True(suite.app.AssetKeeper.IsAddressAuthorizedToSend(suite.ctx, "1", suite.testUser1Acc))
}
