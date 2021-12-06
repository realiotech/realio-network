package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/x/asset/keeper"
	assetSimapp "github.com/realiotech/realio-network/x/asset/simulation"
	"github.com/realiotech/realio-network/x/asset/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

type KeeperTestSuite struct {
	suite.Suite
	app              *app.App
	ctx              sdk.Context
	msgSrv           types.MsgServer
	testUser1Acc     sdk.AccAddress
	testUser1Address string
	testUser2Acc     sdk.AccAddress
	testUser2Address string
}

func (suite *KeeperTestSuite) SetupTest() {
	app := assetSimapp.New("")
	suite.app = app
	suite.msgSrv = keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{Height: 1})

	testUserAcc1 := app.AccountKeeper.NewAccountWithAddress(suite.ctx, sdk.AccAddress("addr1_______________"))
	app.AccountKeeper.SetAccount(suite.ctx, testUserAcc1)
	suite.testUser1Acc = testUserAcc1.GetAddress()
	suite.testUser1Address = testUserAcc1.GetAddress().String()

	testUserAcc2 := app.AccountKeeper.NewAccountWithAddress(suite.ctx, sdk.AccAddress("addr2_______________"))
	app.AccountKeeper.SetAccount(suite.ctx, testUserAcc2)
	suite.testUser2Acc = testUserAcc2.GetAddress()
	suite.testUser2Address = testUserAcc1.GetAddress().String()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
