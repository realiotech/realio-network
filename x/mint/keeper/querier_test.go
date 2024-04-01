package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keep "github.com/realiotech/realio-network/x/mint/keeper"
	"github.com/realiotech/realio-network/x/mint/types"
)

func (suite *KeeperTestSuite) TestNewQuerier() {
	app, ctx, legacyQuerierCdc := suite.app, suite.ctx, suite.legacyQuerierCdc
	querier := keep.NewQuerier(app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{types.QueryParameters}, query)
	suite.Require().NoError(err)

	_, err = querier(ctx, []string{types.QueryInflation}, query)
	suite.Require().NoError(err)

	_, err = querier(ctx, []string{types.QueryAnnualProvisions}, query)
	suite.Require().NoError(err)

	_, err = querier(ctx, []string{"foo"}, query)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	querier := keep.NewQuerier(suite.app.MintKeeper, suite.legacyQuerierCdc.LegacyAmino)

	var params types.Params

	res, sdkErr := querier(suite.ctx, []string{types.QueryParameters}, abci.RequestQuery{})
	suite.Require().NoError(sdkErr)

	err := suite.app.LegacyAmino().UnmarshalJSON(res, &params)
	suite.Require().NoError(err)

	suite.Require().Equal(suite.app.MintKeeper.GetParams(suite.ctx), params)
}

func (suite *KeeperTestSuite) TestQueryInflation() {
	querier := keep.NewQuerier(suite.app.MintKeeper, suite.legacyQuerierCdc.LegacyAmino)

	var inflation sdk.Dec

	res, sdkErr := querier(suite.ctx, []string{types.QueryInflation}, abci.RequestQuery{})
	suite.Require().NoError(sdkErr)

	err := suite.app.LegacyAmino().UnmarshalJSON(res, &inflation)
	suite.Require().NoError(err)

	suite.Require().Equal(suite.app.MintKeeper.GetMinter(suite.ctx).Inflation, inflation)
}

func (suite *KeeperTestSuite) TestQueryAnnualProvisions() {
	querier := keep.NewQuerier(suite.app.MintKeeper, suite.legacyQuerierCdc.LegacyAmino)

	var annualProvisions sdk.Dec

	res, sdkErr := querier(suite.ctx, []string{types.QueryAnnualProvisions}, abci.RequestQuery{})
	suite.Require().NoError(sdkErr)

	err := suite.app.LegacyAmino().UnmarshalJSON(res, &annualProvisions)
	suite.Require().NoError(err)

	suite.Require().Equal(suite.app.MintKeeper.GetMinter(suite.ctx).AnnualProvisions, annualProvisions)
}
