package keeper_test

import (
	"context"

	"github.com/realiotech/realio-network/testutil"
	"github.com/realiotech/realio-network/x/claim/types"
)

func (suite *KeeperTestSuite) TestQueryAdmin() {
	res, err := suite.queryClient.Admin(context.Background(), &types.QueryAdminRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.admin, res.Admin)
}

func (suite *KeeperTestSuite) TestQueryLinkedAddress() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()

	res, err := suite.queryClient.LinkedAddress(context.Background(), &types.QueryLinkedAddressRequest{OldAddress: oldAddr.String()})
	suite.Require().NoError(err)
	suite.Require().False(res.Found)

	suite.Require().NoError(suite.app.ClaimKeeper.SetLink(suite.ctx, oldAddr, newAddr))

	res, err = suite.queryClient.LinkedAddress(context.Background(), &types.QueryLinkedAddressRequest{OldAddress: oldAddr.String()})
	suite.Require().NoError(err)
	suite.Require().True(res.Found)
	suite.Require().Equal(newAddr.String(), res.NewAddress)
}

func (suite *KeeperTestSuite) TestQueryLinkedAddressInvalid() {
	_, err := suite.queryClient.LinkedAddress(context.Background(), &types.QueryLinkedAddressRequest{OldAddress: "not-an-address"})
	suite.Require().Error(err)
}
