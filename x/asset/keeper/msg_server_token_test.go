package keeper_test

import (
	"strconv"
	"strings"

	"cosmossdk.io/math"
	realionetworktypes "github.com/realiotech/realio-network/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func (suite *KeeperTestSuite) TestTokenMsgServerCreate() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	expected := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000",
	}
	_, err := srv.CreateToken(wctx, expected)
	suite.Require().NoError(err)
	lowercased := strings.ToLower(expected.Symbol)
	rio, found := suite.app.AssetKeeper.GetToken(suite.ctx,
		strings.ToLower(lowercased),
	)
	suite.Require().True(found)
	suite.Require().Equal(expected.Manager, rio.Manager)
}

func (suite *KeeperTestSuite) TestTokenMsgServerCreateInvalidSender() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := "invalid"
	expected := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000",
	}
	_, err := srv.CreateToken(wctx, expected)
	suite.Require().ErrorIs(err, sdkerrors.ErrInvalidAddress)
}

func (suite *KeeperTestSuite) TestTokenMsgServerCreateAuthorizationDefaultFalse() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	expected := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000",
	}
	_, err := srv.CreateToken(wctx, expected)
	suite.Require().NoError(err)
	rio, _ := suite.app.AssetKeeper.GetToken(suite.ctx,
		expected.Symbol,
	)
	suite.Require().False(rio.AuthorizationRequired)
}

func (suite *KeeperTestSuite) TestTokenMsgServerCreateErrorDupIndex() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000",
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	t2 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000",
	}
	_, err2 := srv.CreateToken(wctx, t2)
	suite.Require().Error(err2)
	suite.Require().ErrorIs(err2, sdkerrors.ErrInvalidRequest)
}

func (suite *KeeperTestSuite) TestTokenMsgServerCreateVerifyDistribution() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RST", Total: "1000",
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	totalInt, _ := math.NewIntFromString("1000")
	canonicalAmount := totalInt.Mul(realionetworktypes.PowerReduction)

	account, _ := sdk.AccAddressFromBech32(manager)
	managerBalance := suite.app.BankKeeper.GetBalance(suite.ctx, account, "arst")
	suite.Require().Equal(managerBalance.Amount, canonicalAmount)

	totalbalance := suite.app.BankKeeper.GetSupply(suite.ctx, "arst")
	suite.Require().Equal(totalbalance.Amount, canonicalAmount)
}

func (suite *KeeperTestSuite) TestTokenMsgServerUpdate() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000",
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	rio, _ := suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Symbol,
	)
	suite.Require().False(rio.AuthorizationRequired)

	updateMsg := &types.MsgUpdateToken{
		Manager: manager,
		Symbol:  "RIO", AuthorizationRequired: true,
	}

	_, err = srv.UpdateToken(wctx, updateMsg)

	rio, _ = suite.app.AssetKeeper.GetToken(suite.ctx,
		strings.ToLower(t1.Symbol),
	)
	suite.Require().NoError(err)
	suite.Require().True(rio.AuthorizationRequired)
	suite.Require().Equal(rio.Total, "1000")
}

func (suite *KeeperTestSuite) TestTokenMsgServerUpdateNotFound() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000",
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	updateMsg := &types.MsgUpdateToken{
		Manager: manager,
		Symbol:  "RST", AuthorizationRequired: true,
	}

	_, err = srv.UpdateToken(wctx, updateMsg)
	suite.Require().ErrorIs(err, sdkerrors.ErrKeyNotFound)
}

func (suite *KeeperTestSuite) TestTokenMsgServerAuthorizeAddress() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	testUser := suite.testUser2Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	rio, _ := suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Symbol,
	)
	suite.Require().Equal(rio.Authorized[manager].TokenSymbol, "rio")

	authUserMsg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "RIO", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)
	suite.Require().NoError(err)

	rio, _ = suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Symbol,
	)
	suite.Require().NotNil(rio.Authorized)
	suite.Require().Equal(rio.Authorized[testUser].TokenSymbol, "rio")
	suite.Require().Equal(rio.Authorized[testUser].Authorized, true)
}

func (suite *KeeperTestSuite) TestTokenMsgServerAuthorizeTokenNotFound() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	testUser := suite.testUser2Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	authUserMsg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "RST", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)

	suite.Require().ErrorIs(err, sdkerrors.ErrKeyNotFound)
}

func (suite *KeeperTestSuite) TestTokenMsgServerAuthorizeAddressSenderUnauthorized() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	manager2 := suite.testUser2Address
	testUser := suite.testUser3Address

	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	authUserMsg := &types.MsgAuthorizeAddress{
		Manager: manager2,
		Symbol:  "RIO", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)

	suite.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
}

func (suite *KeeperTestSuite) TestTokenMsgServerUnAuthorizeAddress() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	testUser := suite.testUser2Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	rio, _ := suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Symbol,
	)
	suite.Require().Equal(rio.Authorized[manager].TokenSymbol, "rio")

	authUserMsg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "RIO", Address: testUser,
	}

	_, err = srv.AuthorizeAddress(wctx, authUserMsg)
	suite.Require().NoError(err)

	rio, _ = suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Symbol,
	)
	suite.Require().Equal(rio.Authorized[testUser].TokenSymbol, "rio")

	unAuthUserMsg := &types.MsgUnAuthorizeAddress{
		Manager: manager,
		Symbol:  "RIO", Address: testUser,
	}

	_, err = srv.UnAuthorizeAddress(wctx, unAuthUserMsg)
	suite.Require().NoError(err)

	rio, _ = suite.app.AssetKeeper.GetToken(suite.ctx,
		t1.Symbol,
	)
	suite.Require().Nil(rio.Authorized[testUser])
}

func (suite *KeeperTestSuite) TestTokenMsgServerUnAuthorizeTokenNotFound() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	testUser := suite.testUser2Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	unAuthUserMsg := &types.MsgUnAuthorizeAddress{
		Manager: manager,
		Symbol:  "RST", Address: testUser,
	}

	_, err = srv.UnAuthorizeAddress(wctx, unAuthUserMsg)

	suite.Require().ErrorIs(err, sdkerrors.ErrKeyNotFound)
}

func (suite *KeeperTestSuite) TestTokenMsgServerUnAuthorizeAddressSenderUnauthorized() {
	suite.SetupTest()

	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	manager2 := suite.testUser2Address
	testUser := suite.testUser3Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RIO", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	unAuthUserMsg := &types.MsgUnAuthorizeAddress{
		Manager: manager2,
		Symbol:  "RIO", Address: testUser,
	}

	_, err = srv.UnAuthorizeAddress(wctx, unAuthUserMsg)

	suite.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
}
