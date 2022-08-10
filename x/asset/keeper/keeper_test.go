package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/v1/simapp"
	"github.com/realiotech/realio-network/v1/x/asset/keeper"
	"github.com/realiotech/realio-network/v1/x/asset/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"testing"
	"time"
)

type KeeperTestSuite struct {
	suite.Suite
	app              *simapp.SimApp
	ctx              sdk.Context
	msgSrv           types.MsgServer
	testUser1Acc     sdk.AccAddress
	testUser1Address string
	testUser2Acc     sdk.AccAddress
	testUser2Address string
}

func (suite *KeeperTestSuite) SetupTest() {
	simApp := NewSimApp("")
	suite.app = simApp

	suite.msgSrv = keeper.NewMsgServerImpl(simApp.AssetKeeper)
	suite.ctx = simApp.BaseApp.NewContext(false, tmproto.Header{Height: 1})

	testUserAcc1 := simApp.AccountKeeper.NewAccountWithAddress(suite.ctx, sdk.AccAddress("addr1_______________"))
	simApp.AccountKeeper.SetAccount(suite.ctx, testUserAcc1)
	suite.testUser1Acc = testUserAcc1.GetAddress()
	suite.testUser1Address = testUserAcc1.GetAddress().String()

	testUserAcc2 := simApp.AccountKeeper.NewAccountWithAddress(suite.ctx, sdk.AccAddress("addr2_______________"))
	simApp.AccountKeeper.SetAccount(suite.ctx, testUserAcc2)
	suite.testUser2Acc = testUserAcc2.GetAddress()
	suite.testUser2Address = testUserAcc1.GetAddress().String()
	suite.app = simApp
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// New creates application instance with in-memory database and disabled logging.
func NewSimApp(dir string) *simapp.SimApp {
	db := tmdb.NewMemDB()
	logger := log.NewNopLogger()
	encoding := simapp.MakeTestEncodingConfig()

	a := simapp.NewSimApp(logger, db, nil, true, map[int64]bool{}, dir, 0, encoding,
		simapp.EmptyAppOptions{})
	// InitChain updates deliverState which is required when app.NewContext is called
	a.InitChain(abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: defaultConsensusParams,
		AppStateBytes:   []byte("{}"),
	})
	return a
}

var defaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}
