package keeper_test

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"

	"github.com/cosmos/evm/crypto/ethsecp256k1"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/app"
	realiotypes "github.com/realiotech/realio-network/types"
	"github.com/realiotech/realio-network/x/blacklist/keeper"
	"github.com/realiotech/realio-network/x/blacklist/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *app.RealioNetwork
	ctx         sdk.Context
	queryClient types.QueryClient
	authority   string
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false

	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	consAddress := sdk.ConsAddress(priv.PubKey().Address())

	suite.app = app.Setup(checkTx, nil, 1)

	suite.ctx = suite.app.BaseApp.NewContextLegacy(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         realiotypes.MainnetChainID + "-1",
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

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.app.BlacklistKeeper))
	suite.queryClient = types.NewQueryClient(queryHelper)

	suite.authority = suite.app.BlacklistKeeper.GetAuthority()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
