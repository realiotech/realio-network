package apptesting

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
)

type KeeperTestSuite struct {
	suite.Suite
	app *app.RealioNetwork
	ctx sdk.Context

	assetKeeper *keeper.Keeper
	govkeeper   govkeeper.Keeper
	msgServer   types.MsgServer
	bankKeeper  bankkeeper.Keeper
}

func (suite *KeeperTestSuite) SetupTest() {
	app := app.Setup(false, nil, 3)

	suite.app = app
	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{Height: app.LastBlockHeight() + 1})
	suite.assetKeeper = keeper.NewKeeper(
		app.AppCodec(), app.InterfaceRegistry(), app.GetKey(types.StoreKey),
		app.GetMemKey(types.MemStoreKey), app.GetSubspace(types.ModuleName), app.BankKeeper, app.AccountKeeper,
	)
	suite.govkeeper = app.GovKeeper
	suite.bankKeeper = app.BankKeeper
}

func (suite *KeeperTestSuite) 

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
