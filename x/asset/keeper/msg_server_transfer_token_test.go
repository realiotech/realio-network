package keeper_test

import (
	"strconv"
	"strings"

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

	amount := "1234567890.1234567890"
	from, to := suite.testUser1Address, suite.testUser2Address
	expected := &types.MsgTransferToken{Symbol: "RST", From: from, To: to, Amount: amount}

	_, err := srv.TransferToken(wctx, expected)
	suite.Require().NoError(err)

	lowercased := strings.ToLower(expected.Symbol)
	rio, found := suite.app.AssetKeeper.GetToken(suite.ctx,
		strings.ToLower(lowercased),
	)
}
