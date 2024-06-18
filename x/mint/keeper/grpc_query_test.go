package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/v2/x/mint/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)

	inflation, err := suite.queryClient.Inflation(ctx, &types.QueryInflationRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(inflation.Inflation, suite.app.MintKeeper.GetMinter(suite.ctx).Inflation)

	annualProvisions, err := suite.queryClient.AnnualProvisions(ctx, &types.QueryAnnualProvisionsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(annualProvisions.AnnualProvisions, suite.app.MintKeeper.GetMinter(suite.ctx).AnnualProvisions)
}

func (suite *KeeperTestSuite) TestGrpcQueryParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	actualParams := suite.app.MintKeeper.GetParams(suite.ctx)

	params, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(params.Params, actualParams)
}
