package keeper_test

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func (suite *KeeperTestSuite) TestTransferToken() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)

	manager := suite.testUser1Address
	testUser := suite.testUser2Address

	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "rst", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	authUserMsg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "rst", Address: manager,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)
	suite.Require().NoError(err)

	authUser2Msg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "RST", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUser2Msg)
	suite.Require().NoError(err)

	amount := "50000000000000000000"
	expected := &types.MsgTransferToken{Symbol: "RST", From: manager, To: testUser, Amount: amount}

	_, err = srv.TransferToken(wctx, expected)
	suite.Require().NoError(err)

	balance := suite.app.BankKeeper.GetBalance(suite.ctx, suite.testUser2Acc, "arst")
	suite.Require().Equal(balance.String(), fmt.Sprintf("%s%s", amount, "arst"))
}

func (suite *KeeperTestSuite) TestTransferTokenInvalidAmount() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)

	manager := suite.testUser1Address
	testUser := suite.testUser2Address

	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "rst", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	authUserMsg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "rst", Address: manager,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)
	suite.Require().NoError(err)

	authUser2Msg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "rst", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUser2Msg)
	suite.Require().NoError(err)

	// amount is invalid, all amounts should be in base 10^18 amount
	amount := "50000000000000000000.00"
	expected := &types.MsgTransferToken{Symbol: "rst", From: manager, To: testUser, Amount: amount}

	_, err = srv.TransferToken(wctx, expected)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestTransferTokenSenderBalanceToSmall() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)

	manager := suite.testUser1Address
	testUser := suite.testUser2Address

	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RST", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	authUserMsg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "RST", Address: manager,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)
	suite.Require().NoError(err)

	authUser2Msg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "RST", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUser2Msg)
	suite.Require().NoError(err)

	// amount is invalid, all amounts should be in base 10^18 amount
	amount := "1001000000000000000000"
	expected := &types.MsgTransferToken{Symbol: "RST", From: manager, To: testUser, Amount: amount}

	_, err = srv.TransferToken(wctx, expected)
	suite.Require().Error(err)
	suite.Require().Equal(err.Error(), "spendable balance 1000000000000000000000arst is smaller than 1001000000000000000000arst: insufficient funds")
}
