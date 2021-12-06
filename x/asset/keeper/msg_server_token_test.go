package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
	"strconv"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func (suite *KeeperTestSuite) TestTokenMsgServerCreate() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	expected := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000,
	}
	_, err := srv.CreateToken(wctx, expected)
	suite.Require().NoError(err)
	rio, found := suite.app.AssetKeeper.GetToken(suite.ctx,
		expected.Index,
	)
	suite.Require().True(found)
	suite.Require().Equal(expected.Creator, rio.Creator)
}

func (suite *KeeperTestSuite) TestTokenMsgServerCreateInvalidSender() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "invalid"
	expected := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000,
	}
	_, err := srv.CreateToken(wctx, expected)
	suite.Require().ErrorIs(err, sdkerrors.ErrInvalidAddress)
}

func (suite *KeeperTestSuite) TestTokenMsgServerCreateAuthorizationDefaultFalse() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	expected := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000,
	}
	_, err := srv.CreateToken(wctx, expected)
	suite.Require().NoError(err)
	rio, _ := suite.app.AssetKeeper.GetToken(suite.ctx,
		expected.Index,
	)
	suite.Require().False(rio.AuthorizationRequired)
}

func (suite *KeeperTestSuite) TestTokenMsgServerCreateErrorDupIndex() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	t2 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000,
	}
	_, err2 := srv.CreateToken(wctx, t2)
	suite.Require().Error(err2)
	suite.Require().ErrorIs(err2, sdkerrors.ErrInvalidRequest)

}

func (suite *KeeperTestSuite) TestTokenMsgServerCreateVerifyDistribution() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	account, _ := sdk.AccAddressFromBech32(creator)
	creatorBalance := suite.app.BankKeeper.GetBalance(suite.ctx, account, "RIO")
	suite.Require().Equal(creatorBalance.Amount, sdk.NewInt(1000))

	totalbalance := suite.app.BankKeeper.GetSupply(suite.ctx, "RIO")
	suite.Require().Equal(totalbalance.Amount, sdk.NewInt(1000))
}

func (suite *KeeperTestSuite) TestTokenMsgServerUpdate() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	rio, _ := suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Index,
	)
	suite.Require().False(rio.AuthorizationRequired)

	updateMsg := &types.MsgUpdateToken{Creator: creator,
		Index: "1", AuthorizationRequired: true,
	}

	_, err = srv.UpdateToken(wctx, updateMsg)

	rio, _ = suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Index,
	)
	suite.Require().NoError(err)
	suite.Require().True(rio.AuthorizationRequired)
	suite.Require().Equal(rio.Total, int64(1000))
}

func (suite *KeeperTestSuite) TestTokenMsgServerUpdateNotFound() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000,
	}
	_, err := srv.CreateToken(wctx, t1)

	updateMsg := &types.MsgUpdateToken{Creator: creator,
		Index: "2", AuthorizationRequired: true,
	}

	_, err = srv.UpdateToken(wctx, updateMsg)
	suite.Require().ErrorIs(err, sdkerrors.ErrKeyNotFound)
}

func (suite *KeeperTestSuite) TestTokenMsgServerAuthorizeAddress() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	testUser := "cosmos18bc0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000, AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	rio, _ := suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Index,
	)
	suite.Require().Nil(rio.Authorized)

	authUserMsg := &types.MsgAuthorizeAddress{Creator: creator,
		Index: "1", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)

	rio, _ = suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Index,
	)
	suite.Require().NotNil(rio.Authorized)
	suite.Require().Equal(rio.Authorized[testUser].TokenIndex, "1")
	suite.Require().Equal(rio.Authorized[testUser].Authorized, true)
}

func (suite *KeeperTestSuite) TestTokenMsgServerAuthorizeTokenNotFound() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	testUser := "cosmos18bc0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000, AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	authUserMsg := &types.MsgAuthorizeAddress{Creator: creator,
		Index: "2", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)

	suite.Require().ErrorIs(err, sdkerrors.ErrKeyNotFound)
}

func (suite *KeeperTestSuite) TestTokenMsgServerAuthorizeAddressSenderUnauthorized() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	creator2 := "cosmos16ds7p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	testUser := "cosmos18bc0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000, AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)

	authUserMsg := &types.MsgAuthorizeAddress{Creator: creator2,
		Index: "1", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)

	suite.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
}

func (suite *KeeperTestSuite) TestTokenMsgServerUnAuthorizeAddress() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	testUser := "cosmos18bc0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000, AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	rio, _ := suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Index,
	)
	suite.Require().Nil(rio.Authorized)

	authUserMsg := &types.MsgAuthorizeAddress{Creator: creator,
		Index: "1", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)

	rio, _ = suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Index,
	)
	suite.Require().Equal(rio.Authorized[testUser].TokenIndex, "1")

	unAuthUserMsg := &types.MsgUnAuthorizeAddress{Creator: creator,
		Index: "1", Address: testUser,
	}

	_, err = srv.UnAuthorizeAddress(wctx, unAuthUserMsg)

	rio, _ = suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Index,
	)
	suite.Require().Nil(rio.Authorized[testUser])
}

func (suite *KeeperTestSuite) TestTokenMsgServerUnAuthorizeTokenNotFound() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	testUser := "cosmos18bc0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000, AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	unAuthUserMsg := &types.MsgUnAuthorizeAddress{Creator: creator,
		Index: "2", Address: testUser,
	}

	_, err = srv.UnAuthorizeAddress(wctx, unAuthUserMsg)

	suite.Require().ErrorIs(err, sdkerrors.ErrKeyNotFound)
}

func (suite *KeeperTestSuite) TestTokenMsgServerUnAuthorizeAddressSenderUnauthorized() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	creator2 := "cosmos16ds7p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	testUser := "cosmos18bc0p4aep5j83j8d8evwhwwegepjrh9zjn030q"
	t1 := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000, AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	unAuthUserMsg := &types.MsgUnAuthorizeAddress{Creator: creator2,
		Index: "1", Address: testUser,
	}

	_, err = srv.UnAuthorizeAddress(wctx, unAuthUserMsg)

	suite.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
}
