package keeper_test

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/evm/crypto/ethsecp256k1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"
	"github.com/realiotech/realio-network/app"
	realiotypes "github.com/realiotech/realio-network/types"
	"github.com/realiotech/realio-network/x/mint/keeper"
	"github.com/realiotech/realio-network/x/mint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *app.RealioNetwork
	ctx         sdk.Context
	queryClient types.QueryClient
	address     common.Address
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *KeeperTestSuite) DoSetupTest(t *testing.T) {
	checkTx := false

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	consAddress := sdk.ConsAddress(priv.PubKey().Address())

	// init app
	suite.app = app.Setup(checkTx, nil, 1)

	// Set Context
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
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.app.MintKeeper))
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestMintedCoinsEachBlock() {
	suite.DoSetupTest(suite.T())
	rioSupplyCap, _ := math.NewIntFromString("175000000000000000000000000")

	// params.MintDenom, params.BlocksPerYear and minter.Inflation are not changed when go to next block
	params, err := suite.app.MintKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	minter, err := suite.app.MintKeeper.Minter.Get(suite.ctx)
	suite.Require().NoError(err)

	// block 1 vs block 2
	currentSupply := suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions := minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision := annualProvisions.QuoInt(math.NewIntFromUint64(params.BlocksPerYear)).TruncateInt()
	currentHeight := suite.app.LastBlockHeight()

	// block 2
	header := tmproto.Header{Height: currentHeight + 1}
	suite.ctx = suite.ctx.WithBlockHeader(header)
	_, err = suite.app.BeginBlocker(suite.ctx)
	suite.Require().NoError(err)

	newSupply := suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount := newSupply.Sub(currentSupply).String()
	calculatedMintedAmount := blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)

	// block 2 vs block 3
	currentSupply = suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions = minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision = annualProvisions.QuoInt(math.NewIntFromUint64(params.BlocksPerYear)).TruncateInt()
	currentHeight = suite.app.LastBlockHeight()

	// block 3
	header = tmproto.Header{Height: currentHeight + 1}
	suite.ctx = suite.ctx.WithBlockHeader(header)
	_, err = suite.app.BeginBlocker(suite.ctx)
	suite.Require().NoError(err)

	newSupply = suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount = newSupply.Sub(currentSupply).String()
	calculatedMintedAmount = blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)

	// block 3 vs block 4
	currentSupply = suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions = minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision = annualProvisions.QuoInt(math.NewIntFromUint64(params.BlocksPerYear)).TruncateInt()
	currentHeight = suite.app.LastBlockHeight()

	// block 4
	header = tmproto.Header{Height: currentHeight + 1}
	suite.ctx = suite.ctx.WithBlockHeader(header)
	_, err = suite.app.BeginBlocker(suite.ctx)
	suite.Require().NoError(err)

	newSupply = suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount = newSupply.Sub(currentSupply).String()
	calculatedMintedAmount = blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)

	// block 4 vs block 5
	currentSupply = suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions = minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision = annualProvisions.QuoInt(math.NewIntFromUint64(params.BlocksPerYear)).TruncateInt()
	currentHeight = suite.app.LastBlockHeight()

	// block 5
	header = tmproto.Header{Height: currentHeight + 1}
	suite.ctx = suite.ctx.WithBlockHeader(header)
	_, err = suite.app.BeginBlocker(suite.ctx)
	suite.Require().NoError(err)

	newSupply = suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount = newSupply.Sub(currentSupply).String()
	calculatedMintedAmount = blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)

	// block 5 vs block 6
	currentSupply = suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions = minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision = annualProvisions.QuoInt(math.NewIntFromUint64(params.BlocksPerYear)).TruncateInt()
	currentHeight = suite.app.LastBlockHeight()

	// block 6
	header = tmproto.Header{Height: currentHeight + 1}
	suite.ctx = suite.ctx.WithBlockHeader(header)
	_, err = suite.app.BeginBlocker(suite.ctx)
	suite.Require().NoError(err)

	newSupply = suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount = newSupply.Sub(currentSupply).String()
	calculatedMintedAmount = blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)
}
