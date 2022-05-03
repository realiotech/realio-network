package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/asset/types"
	"strconv"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func (suite *KeeperTestSuite) TestTransferTokenWithAuthorization() {
	suite.SetupTest()

	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"

	createMsg := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000, AuthorizationRequired: true,
	}
	_, _ = suite.msgSrv.CreateToken(wctx, createMsg)

	authUserMsg := &types.MsgAuthorizeAddress{Creator: creator,
		Symbol: "RIO", Address: suite.testUser1Address,
	}
	_, _ = suite.msgSrv.AuthorizeAddress(wctx, authUserMsg)

	authUserMsg2 := &types.MsgAuthorizeAddress{Creator: creator,
		Symbol: "RIO", Address: creator,
	}
	_, _ = suite.msgSrv.AuthorizeAddress(wctx, authUserMsg2)

	transferMsg := &types.MsgTransferToken{Creator: creator,
		Index: "1", Symbol: "RIO", From: creator, To: suite.testUser1Address, Amount: 10,
	}
	_, err := suite.msgSrv.TransferToken(wctx, transferMsg)
	suite.Require().NoError(err)

	balance := suite.app.BankKeeper.GetBalance(suite.ctx, suite.testUser1Acc, "RIO")
	suite.Require().Equal(balance.Amount, sdk.NewInt(10))
}

func (suite *KeeperTestSuite) TestTransferTokenWithoutAuthorization() {
	suite.SetupTest()

	wctx := sdk.WrapSDKContext(suite.ctx)
	creator := "cosmos19cm0p4aep5j83j8d8evwhwwegepjrh9zjn030q"

	createMsg := &types.MsgCreateToken{Creator: creator,
		Index: "1", Symbol: "RIO", Total: 1000, AuthorizationRequired: false,
	}
	_, _ = suite.msgSrv.CreateToken(wctx, createMsg)

	transferMsg := &types.MsgTransferToken{Creator: creator,
		Index: "1", Symbol: "RIO", From: creator, To: suite.testUser1Address, Amount: 99,
	}
	_, err := suite.msgSrv.TransferToken(wctx, transferMsg)
	suite.Require().NoError(err)

	balance := suite.app.BankKeeper.GetBalance(suite.ctx, suite.testUser1Acc, "RIO")
	suite.Require().Equal(balance.Amount, sdk.NewInt(99))
}

func (suite *KeeperTestSuite) TestTransferTokenNotFound() {
	suite.SetupTest()

	wctx := sdk.WrapSDKContext(suite.ctx)

	transferMsg := &types.MsgTransferToken{Creator: suite.testUser1Address,
		Index: "1", Symbol: "RIO", From: suite.testUser1Address, To: suite.testUser2Address, Amount: 99,
	}
	_, err := suite.msgSrv.TransferToken(wctx, transferMsg)
	suite.Require().ErrorIs(err, sdkerrors.ErrKeyNotFound)

}
