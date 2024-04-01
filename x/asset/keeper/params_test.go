package keeper_test

import (
	"github.com/realiotech/realio-network/x/asset/types"
)

func (suite *KeeperTestSuite) TestGetParams() {
	suite.SetupTest()

	k := suite.app.AssetKeeper
	ctx := suite.ctx
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	suite.Require().Equal(params, k.GetParams(ctx))
}
