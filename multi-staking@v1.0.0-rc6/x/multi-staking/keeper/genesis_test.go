package keeper_test

import (
	"github.com/realio-tech/multi-staking-module/test/simapp"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func (suite *KeeperTestSuite) TestImportExportGenesis() {
	oldCommitHash := suite.app.LastCommitID().Hash
	appState, err := suite.app.ExportAppStateAndValidators(false, []string{})
	suite.NoError(err)

	encConfig := simapp.MakeTestEncodingConfig()

	emptyApp := simapp.NewSimApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		"temp",
		simapp.FlagPeriodValue,
		encConfig,
		simapp.EmptyAppOptions{},
	)

	_ = emptyApp.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   appState.AppState,
		},
	)

	newAppState, err := suite.app.ExportAppStateAndValidators(false, []string{})
	emptyApp.Commit()
	newCommitHash := emptyApp.LastCommitID().Hash
	suite.NoError(err)

	suite.Equal(appState, newAppState)
	suite.Equal(newCommitHash, oldCommitHash)
}
