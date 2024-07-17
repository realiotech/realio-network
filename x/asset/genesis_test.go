package asset_test

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
	sdk "github.com/cosmos/cosmos-sdk/types"
	feemarkettypes "github.com/evmos/evmos/v18/x/feemarket/types"
	"github.com/realiotech/realio-network/v2/app"
	"github.com/realiotech/realio-network/v2/testutil"
	realiotypes "github.com/realiotech/realio-network/v2/types"
	"github.com/realiotech/realio-network/v2/x/asset"
	"github.com/realiotech/realio-network/v2/x/asset/types"
	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app     *app.RealioNetwork
	genesis types.GenesisState
}

func (suite *GenesisTestSuite) SetupTest() {
	// consensus key
	consAddress := sdk.ConsAddress(testutil.GenAddress().Bytes())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState(), 1)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         realiotypes.MainnetChainID,
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	suite.genesis = *types.DefaultGenesis()
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestGenesis() {
	asset.InitGenesis(suite.ctx, suite.app.AssetKeeper, suite.genesis)
	got := asset.ExportGenesis(suite.ctx, suite.app.AssetKeeper)
	suite.Require().NotNil(got)

	suite.Require().Equal(len(suite.genesis.Tokens), len(got.Tokens))
}
