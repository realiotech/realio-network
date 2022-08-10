package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/realiotech/realio-network/testutil/keeper"
	"github.com/realiotech/realio-network/v1/x/asset/keeper"
	"github.com/realiotech/realio-network/v1/x/asset/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.AssetKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
