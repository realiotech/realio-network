package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

func (suite *KeeperTestSuite) TestQueryHackedAmount() {
	oldAddr := testutil.GenAddress()

	res, err := suite.queryClient.HackedAmount(context.Background(), &types.QueryHackedAmountRequest{OldAddress: oldAddr.String()})
	suite.Require().NoError(err)
	suite.Require().True(res.Amount.IsZero(), "nothing is recorded by default")

	amount := sdk.NewCoins(sdk.NewInt64Coin("ario", 500), sdk.NewInt64Coin("arst", 7))
	suite.Require().NoError(suite.app.ClaimKeeper.SetHackedAmount(suite.ctx, oldAddr, amount))

	res, err = suite.queryClient.HackedAmount(context.Background(), &types.QueryHackedAmountRequest{OldAddress: oldAddr.String()})
	suite.Require().NoError(err)
	suite.Require().Equal(amount, res.Amount)

	// A payout deletes the record, and the query follows.
	suite.Require().NoError(suite.app.ClaimKeeper.ClearHackedAmount(suite.ctx, oldAddr))
	res, err = suite.queryClient.HackedAmount(context.Background(), &types.QueryHackedAmountRequest{OldAddress: oldAddr.String()})
	suite.Require().NoError(err)
	suite.Require().True(res.Amount.IsZero())
}

func (suite *KeeperTestSuite) TestQueryHackedAmountInvalid() {
	_, err := suite.queryClient.HackedAmount(context.Background(), &types.QueryHackedAmountRequest{OldAddress: "not-an-address"})
	suite.Require().Error(err)
}
