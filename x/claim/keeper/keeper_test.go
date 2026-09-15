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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/realiotech/realio-network/app"
	realiotypes "github.com/realiotech/realio-network/types"
	"github.com/realiotech/realio-network/x/claim/keeper"
	"github.com/realiotech/realio-network/x/claim/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *app.RealioNetwork
	ctx         sdk.Context
	queryClient types.QueryClient

	// admin is the address configured as x/claim's admin for every test.
	admin string

	// validator is the operator address of the single bonded validator
	// app.Setup seeds at genesis (see app/test_helpers.go's
	// GenesisStateWithValSet) — its registered multi-staking coin denom is
	// "ario".
	validator sdk.ValAddress
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
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.app.ClaimKeeper))
	suite.queryClient = types.NewQueryClient(queryHelper)

	suite.admin = sdk.AccAddress(priv.PubKey().Address()).String()
	suite.Require().NoError(suite.app.ClaimKeeper.SetAdmin(suite.ctx, suite.admin))

	vals, err := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(vals, 1)
	valAddr, err := sdk.ValAddressFromBech32(vals[0].OperatorAddress)
	suite.Require().NoError(err)
	suite.validator = valAddr
	suite.Require().Equal(stakingtypes.Bonded, vals[0].Status)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
