package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/crypto/ethsecp256k1"
	"github.com/realiotech/realio-network/v2/app"
	realiotypes "github.com/realiotech/realio-network/v2/types"
	"github.com/realiotech/realio-network/v2/x/mint/keeper"
	"github.com/realiotech/realio-network/v2/x/mint/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	app              *app.RealioNetwork
	ctx              sdk.Context
	queryClient      types.QueryClient
	address          common.Address
	msgServer        types.MsgServer
	legacyQuerierCdc *codec.AminoCodec
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *KeeperTestSuite) DoSetupTest(t *testing.T) {
	t.Helper()
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
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
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
	types.RegisterQueryServer(queryHelper, suite.app.MintKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
	suite.msgServer = keeper.NewMsgServerImpl(suite.app.MintKeeper)
	suite.legacyQuerierCdc = codec.NewAminoCodec(suite.app.LegacyAmino())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestMintedCoinsEachBlock() {
	suite.DoSetupTest(suite.T())
	rioSupplyCap, _ := math.NewIntFromString("75000000000000000000000000")

	// params.MintDenom, params.BlocksPerYear and minter.Inflation are not changed when go to next block
	params := suite.app.MintKeeper.GetParams(suite.ctx)
	minter := suite.app.MintKeeper.GetMinter(suite.ctx)

	// block 1 vs block 2
	currentSupply := suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions := minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision := annualProvisions.QuoInt(math.NewInt(int64(params.BlocksPerYear))).TruncateInt()
	currentHeight := suite.app.LastBlockHeight()

	// block 2
	header := tmproto.Header{
		Height:  currentHeight + 1,
		ChainID: realiotypes.MainnetChainID + "-1",
	}
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: header})

	newSupply := suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount := newSupply.Sub(currentSupply).String()
	calculatedMintedAmount := blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)

	// block 2 vs block 3
	currentSupply = suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions = minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision = annualProvisions.QuoInt(math.NewInt(int64(params.BlocksPerYear))).TruncateInt()
	currentHeight = suite.app.LastBlockHeight()

	// block 3
	header = tmproto.Header{
		Height:  currentHeight + 1,
		ChainID: realiotypes.MainnetChainID + "-1",
	}
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: header})

	newSupply = suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount = newSupply.Sub(currentSupply).String()
	calculatedMintedAmount = blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)

	// block 3 vs block 4
	currentSupply = suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions = minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision = annualProvisions.QuoInt(math.NewInt(int64(params.BlocksPerYear))).TruncateInt()
	currentHeight = suite.app.LastBlockHeight()

	// block 4
	header = tmproto.Header{
		Height:  currentHeight + 1,
		ChainID: realiotypes.MainnetChainID + "-1",
	}
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: header})

	newSupply = suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount = newSupply.Sub(currentSupply).String()
	calculatedMintedAmount = blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)

	// block 4 vs block 5
	currentSupply = suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions = minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision = annualProvisions.QuoInt(math.NewInt(int64(params.BlocksPerYear))).TruncateInt()
	currentHeight = suite.app.LastBlockHeight()

	// block 5
	header = tmproto.Header{
		Height:  currentHeight + 1,
		ChainID: realiotypes.MainnetChainID + "-1",
	}
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: header})

	newSupply = suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount = newSupply.Sub(currentSupply).String()
	calculatedMintedAmount = blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)

	// block 5 vs block 6
	currentSupply = suite.app.BankKeeper.GetSupply(suite.ctx, params.MintDenom).Amount
	annualProvisions = minter.Inflation.MulInt(rioSupplyCap.Sub(currentSupply))
	blockProvision = annualProvisions.QuoInt(math.NewInt(int64(params.BlocksPerYear))).TruncateInt()
	currentHeight = suite.app.LastBlockHeight()

	// block 6
	header = tmproto.Header{
		Height:  currentHeight + 1,
		ChainID: realiotypes.MainnetChainID + "-1",
	}
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: header})

	newSupply = suite.app.MintKeeper.StakingTokenSupply(suite.ctx, params)
	expectedMintedAmount = newSupply.Sub(currentSupply).String()
	calculatedMintedAmount = blockProvision.String()
	suite.Require().Equal(expectedMintedAmount, calculatedMintedAmount)
}

func (suite *KeeperTestSuite) TestParam() {
	testCases := []struct {
		name      string
		input     types.Params
		expectErr bool
	}{
		{
			name: "set invalid params",
			input: types.Params{
				MintDenom:     sdk.DefaultBondDenom,
				InflationRate: sdk.NewDecWithPrec(-13, 2),
				BlocksPerYear: uint64(60 * 60 * 8766 / 5),
			},
			expectErr: true,
		},
		{
			name: "set full valid params",
			input: types.Params{
				MintDenom:     sdk.DefaultBondDenom,
				InflationRate: sdk.NewDecWithPrec(12, 2),
				BlocksPerYear: uint64(60 * 60 * 8766 / 5),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.DoSetupTest(suite.T())
			expected := suite.app.MintKeeper.GetParams(suite.ctx)
			err := suite.app.MintKeeper.SetParams(suite.ctx, tc.input)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				expected = tc.input
				suite.Require().NoError(err)
			}

			p := suite.app.MintKeeper.GetParams(suite.ctx)
			suite.Require().Equal(expected, p)
		})
	}
}
