package keeper_test

import (
	"github.com/realiotech/realio-network/testutil"
)

func (suite *KeeperTestSuite) TestSetIsRemoveBlacklisted() {
	addr := testutil.GenAddress()

	suite.Require().False(suite.app.BlacklistKeeper.IsBlacklisted(suite.ctx, addr))

	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, addr))
	suite.Require().True(suite.app.BlacklistKeeper.IsBlacklisted(suite.ctx, addr))

	suite.Require().NoError(suite.app.BlacklistKeeper.RemoveBlacklisted(suite.ctx, addr))
	suite.Require().False(suite.app.BlacklistKeeper.IsBlacklisted(suite.ctx, addr))
}

func (suite *KeeperTestSuite) TestGetAllBlacklisted() {
	addrA := testutil.GenAddress()
	addrB := testutil.GenAddress()

	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, addrA))
	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, addrB))

	all, err := suite.app.BlacklistKeeper.GetAllBlacklisted(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().ElementsMatch([]string{addrA.String(), addrB.String()}, all)
}
