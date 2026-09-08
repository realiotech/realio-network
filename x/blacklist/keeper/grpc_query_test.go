package keeper_test

import (
	"github.com/realiotech/realio-network/testutil"
	"github.com/realiotech/realio-network/x/blacklist/types"
)

func (suite *KeeperTestSuite) TestGRPCQueryIsBlacklisted() {
	blacklisted := testutil.GenAddress()
	clean := testutil.GenAddress()

	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, blacklisted))

	res, err := suite.queryClient.IsBlacklisted(suite.ctx, &types.QueryIsBlacklistedRequest{Address: blacklisted.String()})
	suite.Require().NoError(err)
	suite.Require().True(res.Blacklisted)

	res, err = suite.queryClient.IsBlacklisted(suite.ctx, &types.QueryIsBlacklistedRequest{Address: clean.String()})
	suite.Require().NoError(err)
	suite.Require().False(res.Blacklisted)

	_, err = suite.queryClient.IsBlacklisted(suite.ctx, &types.QueryIsBlacklistedRequest{Address: "not-a-valid-address"})
	suite.Require().Error(err)
}
