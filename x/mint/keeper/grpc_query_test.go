package keeper_test

import (
	"github.com/realiotech/realio-network/x/mint/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	ctx := suite.ctx

	inflation, err := suite.queryClient.Inflation(ctx, &types.QueryInflationRequest{})
	suite.Require().NoError(err)
	minter, err := suite.app.MintKeeper.Minter.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(inflation.Inflation, minter.Inflation)

	annualProvisions, err := suite.queryClient.AnnualProvisions(ctx, &types.QueryAnnualProvisionsRequest{})
	suite.Require().NoError(err)
	minter, err = suite.app.MintKeeper.Minter.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(annualProvisions.AnnualProvisions, minter.AnnualProvisions)
}

func (suite *KeeperTestSuite) TestGrpcQueryParams() {
	actualParams, err := suite.app.MintKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)

	params, err := suite.queryClient.Params(suite.ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(params.Params, actualParams)
}
