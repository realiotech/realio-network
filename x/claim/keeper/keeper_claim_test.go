package keeper_test

import (
	"github.com/realiotech/realio-network/testutil"
	"github.com/realiotech/realio-network/x/claim/types"
)

func (suite *KeeperTestSuite) TestGetSetAdmin() {
	admin, ok := suite.app.ClaimKeeper.GetAdmin(suite.ctx)
	suite.Require().True(ok)
	suite.Require().Equal(suite.admin, admin.String())

	newAdmin := testutil.GenAddress()
	suite.Require().NoError(suite.app.ClaimKeeper.SetAdmin(suite.ctx, newAdmin))
	gotAdmin, ok := suite.app.ClaimKeeper.GetAdmin(suite.ctx)
	suite.Require().True(ok)
	suite.Require().Equal(newAdmin, gotAdmin)
}

func (suite *KeeperTestSuite) TestGetSetLink() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()

	_, found := suite.app.ClaimKeeper.GetLink(suite.ctx, oldAddr)
	suite.Require().False(found)
	suite.Require().False(suite.app.ClaimKeeper.IsLinked(suite.ctx, oldAddr))

	suite.Require().NoError(suite.app.ClaimKeeper.SetLink(suite.ctx, oldAddr, newAddr))

	got, found := suite.app.ClaimKeeper.GetLink(suite.ctx, oldAddr)
	suite.Require().True(found)
	suite.Require().Equal(newAddr, got)
	suite.Require().True(suite.app.ClaimKeeper.IsLinked(suite.ctx, oldAddr))
}

func (suite *KeeperTestSuite) TestGetAllLinks() {
	oldA, newA := testutil.GenAddress(), testutil.GenAddress()
	oldB, newB := testutil.GenAddress(), testutil.GenAddress()

	suite.Require().NoError(suite.app.ClaimKeeper.SetLink(suite.ctx, oldA, newA))
	suite.Require().NoError(suite.app.ClaimKeeper.SetLink(suite.ctx, oldB, newB))

	links, err := suite.app.ClaimKeeper.GetAllLinks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().ElementsMatch([]types.AddressLink{
		{OldAddress: oldA.String(), NewAddress: newA.String()},
		{OldAddress: oldB.String(), NewAddress: newB.String()},
	}, links)
}
