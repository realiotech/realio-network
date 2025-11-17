package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/stretchr/testify/require"
)

func Test14UpgradeRemoveLMX(t *testing.T) {
	realioApp := SetupWithGenFiles(t)

	ctx := realioApp.BaseApp.NewContextLegacy(false, tmproto.Header{Height: realioApp.LastBlockHeight() + 1, Time: time.Now()})

	// Set block height and block time same as genesis file
	blockTime, _ := time.Parse(time.RFC3339Nano, "2025-09-26T05:43:53.576222314Z")
	ctx = ctx.WithBlockHeight(14432412).WithBlockTime(blockTime).WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	// Test Remove lmx proposal
	err := realioApp.MultiStakingKeeper.RemoveMultiStakingCoinProposal(ctx, &multistakingtypes.RemoveMultiStakingCoinProposal{
		Title:       "title",
		Description: "description",
		Denom:       "almx",
	})
	require.NoError(t, err)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	qServer := multistakingkeeper.NewQueryServerImpl(realioApp.MultiStakingKeeper)

	// Check balances BEFORE unbonding delegation finishes
	ubdRes, err := qServer.MultiStakingUnlocks(ctx, &multistakingtypes.QueryMultiStakingUnlocksRequest{})
	require.NoError(t, err)

	// Check balances for each delegator before unbonding finishes
	beforeBals := []sdk.Coin{}
	for _, unlock := range ubdRes.Unlocks {
		if unlock.Entries[0].UnlockingCoin.Denom != "almx" {
			continue
		}
		delegatorAddr, err := sdk.AccAddressFromBech32(unlock.UnlockID.MultiStakerAddr)
		require.NoError(t, err)
		balances := realioApp.BankKeeper.GetBalance(ctx, delegatorAddr, "almx")
		beforeBals = append(beforeBals, balances)
	}

	// Move time forward to complete unbonding (7 days = 604800 seconds)
	ctx = ctx.WithBlockTime(blockTime.Add(time.Second * 604800)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	// Move 1 block
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Second)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	// Check balances AFTER unbonding delegation finishes
	// Check balances for each delegator after unbonding finishes
	afterBals := []sdk.Coin{}
	for _, unlock := range ubdRes.Unlocks {
		if unlock.Entries[0].UnlockingCoin.Denom != "almx" {
			continue
		}
		delegatorAddr, err := sdk.AccAddressFromBech32(unlock.UnlockID.MultiStakerAddr)
		require.NoError(t, err)
		balances := realioApp.BankKeeper.GetBalance(ctx, delegatorAddr, "almx")
		afterBals = append(afterBals, balances)
	}

	// Check almx balance increase after unbonding complete
	for i := range afterBals {
		require.True(t, beforeBals[i].IsLT(afterBals[i]))
	}
}

func Test14UpgradeRemoveRST(t *testing.T) {
	realioApp := SetupWithGenFiles(t)

	ctx := realioApp.BaseApp.NewContextLegacy(false, tmproto.Header{Height: realioApp.LastBlockHeight() + 1, Time: time.Now()})

	// Set block height and block time same as genesis file
	blockTime, _ := time.Parse(time.RFC3339Nano, "2025-09-26T05:43:53.576222314Z")
	ctx = ctx.WithBlockHeight(14432412).WithBlockTime(blockTime).WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	// Test Remove lmx proposal
	err := realioApp.MultiStakingKeeper.RemoveMultiStakingCoinProposal(ctx, &multistakingtypes.RemoveMultiStakingCoinProposal{
		Title:       "title",
		Description: "description",
		Denom:       "arst",
	})
	require.NoError(t, err)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	qServer := multistakingkeeper.NewQueryServerImpl(realioApp.MultiStakingKeeper)

	// Check balances BEFORE unbonding delegation finishes
	ubdRes, err := qServer.MultiStakingUnlocks(ctx, &multistakingtypes.QueryMultiStakingUnlocksRequest{})
	require.NoError(t, err)

	// Check balances for each delegator before unbonding finishes
	beforeBals := []sdk.Coin{}
	for _, unlock := range ubdRes.Unlocks {
		if unlock.Entries[0].UnlockingCoin.Denom != "arst" {
			continue
		}
		delegatorAddr, err := sdk.AccAddressFromBech32(unlock.UnlockID.MultiStakerAddr)
		require.NoError(t, err)
		balances := realioApp.BankKeeper.GetBalance(ctx, delegatorAddr, "arst")
		beforeBals = append(beforeBals, balances)
	}

	// Move time forward to complete unbonding (7 days = 604800 seconds)
	ctx = ctx.WithBlockTime(blockTime.Add(time.Second * 604800)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	// Move 1 block
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Second)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	// Check balances AFTER unbonding delegation finishes
	// Check balances for each delegator after unbonding finishes
	afterBals := []sdk.Coin{}
	for _, unlock := range ubdRes.Unlocks {
		if unlock.Entries[0].UnlockingCoin.Denom != "arst" {
			continue
		}
		delegatorAddr, err := sdk.AccAddressFromBech32(unlock.UnlockID.MultiStakerAddr)
		require.NoError(t, err)
		balances := realioApp.BankKeeper.GetBalance(ctx, delegatorAddr, "arst")
		afterBals = append(afterBals, balances)
	}

	// Check arst balance increase after unbonding complete
	for i := range afterBals {
		require.True(t, beforeBals[i].IsLT(afterBals[i]))
	}
}

func Test14UpgradeRemoveRIO(t *testing.T) {
	realioApp := SetupWithGenFiles(t)

	ctx := realioApp.BaseApp.NewContextLegacy(false, tmproto.Header{Height: realioApp.LastBlockHeight() + 1, Time: time.Now()})

	// Set block height and block time same as genesis file
	blockTime, _ := time.Parse(time.RFC3339Nano, "2025-09-26T05:43:53.576222314Z")
	ctx = ctx.WithBlockHeight(14432412).WithBlockTime(blockTime).WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	// Test Remove lmx proposal
	err := realioApp.MultiStakingKeeper.RemoveMultiStakingCoinProposal(ctx, &multistakingtypes.RemoveMultiStakingCoinProposal{
		Title:       "title",
		Description: "description",
		Denom:       "ario",
	})
	require.NoError(t, err)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	qServer := multistakingkeeper.NewQueryServerImpl(realioApp.MultiStakingKeeper)

	// Check balances BEFORE unbonding delegation finishes
	ubdRes, err := qServer.MultiStakingUnlocks(ctx, &multistakingtypes.QueryMultiStakingUnlocksRequest{})
	require.NoError(t, err)

	// Check balances for each delegator before unbonding finishes
	beforeBals := []sdk.Coin{}
	for _, unlock := range ubdRes.Unlocks {
		if unlock.Entries[0].UnlockingCoin.Denom != "ario" {
			continue
		}
		delegatorAddr, err := sdk.AccAddressFromBech32(unlock.UnlockID.MultiStakerAddr)
		require.NoError(t, err)
		balances := realioApp.BankKeeper.GetBalance(ctx, delegatorAddr, "ario")
		beforeBals = append(beforeBals, balances)
	}

	// Move time forward to complete unbonding (7 days = 604800 seconds)
	ctx = ctx.WithBlockTime(blockTime.Add(time.Second * 604800 * 3)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	// Move 1 block
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Second)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	// Check balances AFTER unbonding delegation finishes
	// Check balances for each delegator after unbonding finishes
	afterBals := []sdk.Coin{}
	for _, unlock := range ubdRes.Unlocks {
		if unlock.Entries[0].UnlockingCoin.Denom != "ario" {
			continue
		}
		delegatorAddr, err := sdk.AccAddressFromBech32(unlock.UnlockID.MultiStakerAddr)
		require.NoError(t, err)
		balances := realioApp.BankKeeper.GetBalance(ctx, delegatorAddr, "ario")
		afterBals = append(afterBals, balances)
	}

	// Check ario balance increase after unbonding complete
	for i := range afterBals {
		require.True(t, beforeBals[i].IsLT(afterBals[i]))
	}
}

func SetupWithGenFiles(t *testing.T) *RealioNetwork {
	// Load the exported mainnet genesis file
	genFile := filepath.Join("upgrades", "v1.4", "testdata", "exported_mainnet_2609.json")
	genData, err := os.ReadFile(genFile)
	require.NoError(t, err, "failed to read exported genesis file")

	// Parse the genesis file
	var genesisDoc map[string]interface{}
	err = json.Unmarshal(genData, &genesisDoc)
	require.NoError(t, err, "failed to unmarshal genesis file")

	// Extract app_state and other genesis fields
	appState, ok := genesisDoc["app_state"].(map[string]interface{})
	require.True(t, ok, "failed to extract app_state from genesis")

	chainID := genesisDoc["chain_id"].(string)
	genesisTime, _ := time.Parse(time.RFC3339Nano, "2025-09-26T05:43:53.576222314Z")
	initialHeight := int64(14432412) // From initial_height in file

	// Create the app directly (similar to Setup but without default genesis)
	// encCdc := MakeEncodingConfig(MainnetChainID)
	db := dbm.NewMemDB()
	opt := baseapp.SetChainID(chainID)

	// Create app options with crisis flags
	appOpts := make(simtestutil.AppOptionsMap)
	appOpts[flags.FlagHome] = DefaultNodeHome
	appOpts[crisis.FlagSkipGenesisInvariants] = false // Skip genesis invariants

	realioApp := New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		1,
		appOpts,
		opt,
	)

	// Convert appState to proper genesis format and initialize chain
	stateBytes, err := json.MarshalIndent(appState, "", " ")
	require.NoError(t, err, "failed to marshal app state")

	// Initialize the chain with exported genesis
	_, err = realioApp.InitChain(&abci.RequestInitChain{
		Time:            genesisTime,
		InitialHeight:   initialHeight,
		ChainId:         chainID,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err, "failed to initialize chain")

	// Finalize the first block
	_, err = realioApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: initialHeight,
		Time:   genesisTime,
		Txs:    [][]byte{},
	})
	require.NoError(t, err, "failed to finalize block")

	return realioApp
}
