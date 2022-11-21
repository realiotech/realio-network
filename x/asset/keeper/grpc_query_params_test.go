package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/asset/types"
)

func (suite *KeeperTestSuite) TestParamsQuery() {
	suite.SetupTest()

	k := suite.app.AssetKeeper
	wctx := sdk.WrapSDKContext(suite.ctx)

	params := types.DefaultParams()
	k.SetParams(suite.ctx, params)

	response, err := k.Params(wctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}
