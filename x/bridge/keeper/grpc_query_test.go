package keeper_test

import (
	"github.com/realiotech/realio-network/x/bridge/types"
)

func (suite *KeeperTestSuite) TestGRPCQuery() {
	ratelimits, err := suite.queryClient.RateLimits(suite.ctx, &types.QueryRateLimitsRequest{})
	suite.Require().NoError(err)

	expectedRateLimits := []types.RateLimit{}
	err = suite.app.BridgeKeeper.RegisteredCoins.Walk(suite.ctx, nil, func(_ string, ratelimit types.RateLimit) (stop bool, err error) {
		expectedRateLimits = append(expectedRateLimits, ratelimit)
		return false, nil
	})
	suite.Require().NoError(err)
	suite.Require().Equal(len(ratelimits.Ratelimits), 2)
	suite.Require().Equal(ratelimits.Ratelimits, expectedRateLimits)

	epochInfo, err := suite.queryClient.EpochInfo(suite.ctx, &types.QueryEpochInfoRequest{})
	suite.Require().NoError(err)

	expectedEpochInfo, err := suite.app.BridgeKeeper.EpochInfo.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(epochInfo.EpochInfo, expectedEpochInfo)
}

func (suite *KeeperTestSuite) TestGrpcQueryParams() {
	actualParams, err := suite.app.BridgeKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)

	params, err := suite.queryClient.Params(suite.ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(params.Params, actualParams)
}
