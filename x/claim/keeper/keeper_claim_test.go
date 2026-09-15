package keeper_test

import (
	"github.com/realiotech/realio-network/testutil"
	"github.com/realiotech/realio-network/x/claim/types"
)

func (suite *KeeperTestSuite) TestGetSetAdmin() {
	suite.Require().Equal(suite.admin, suite.app.ClaimKeeper.GetAdmin(suite.ctx))

	newAdmin := testutil.GenAddress().String()
	suite.Require().NoError(suite.app.ClaimKeeper.SetAdmin(suite.ctx, newAdmin))
	suite.Require().Equal(newAdmin, suite.app.ClaimKeeper.GetAdmin(suite.ctx))
}

func (suite *KeeperTestSuite) TestGetSetLink() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()

	_, found := suite.app.ClaimKeeper.GetLink(suite.ctx, oldAddr)
	suite.Require().False(found)
	suite.Require().False(suite.app.ClaimKeeper.IsLinked(suite.ctx, oldAddr))

	suite.Require().NoError(suite.app.ClaimKeeper.SetLink(suite.ctx, oldAddr, newAddr.String()))

	got, found := suite.app.ClaimKeeper.GetLink(suite.ctx, oldAddr)
	suite.Require().True(found)
	suite.Require().Equal(newAddr.String(), got)
	suite.Require().True(suite.app.ClaimKeeper.IsLinked(suite.ctx, oldAddr))
}

func (suite *KeeperTestSuite) TestGetAllLinks() {
	oldA, newA := testutil.GenAddress(), testutil.GenAddress()
	oldB, newB := testutil.GenAddress(), testutil.GenAddress()

	suite.Require().NoError(suite.app.ClaimKeeper.SetLink(suite.ctx, oldA, newA.String()))
	suite.Require().NoError(suite.app.ClaimKeeper.SetLink(suite.ctx, oldB, newB.String()))

	links, err := suite.app.ClaimKeeper.GetAllLinks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().ElementsMatch([]types.AddressLink{
		{OldAddress: oldA.String(), NewAddress: newA.String()},
		{OldAddress: oldB.String(), NewAddress: newB.String()},
	}, links)
}
