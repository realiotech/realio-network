package app

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/realiotech/realio-network/app/migrations"
)

// RealGenesisPath is the actual pre-incident mainnet genesis export used by
// app/migrations' tests. Lives under app/migrations/testdata (a Go "testdata"
// directory: ignored by the go tool for build purposes), resolved relative to
// that package's directory. Tests in other packages keep their own fixture
// under their own testdata directory and call SetupWithGenesisFile directly.
const RealGenesisPath = "testdata/recover_genesis.json"

// ConsensusValidatorEntry mirrors the top-level consensus.validators[] shape
// in the exported genesis (hex address + base64 ed25519 pubkey + power),
// just enough to pick a ProposerAddress for FinalizeBlock.
type ConsensusValidatorEntry struct {
	Address string `json:"address"`
}

// SetupWithRealGenesis boots a full app against app/migrations' real genesis
// export (see SetupWithGenesisFile).
func SetupWithRealGenesis(t *testing.T) (realioApp *RealioNetwork, chainID string, initialHeight int64, proposerAddr []byte, blockTime time.Time) {
	t.Helper()
	return SetupWithGenesisFile(t, RealGenesisPath)
}

// SetupWithGenesisFile boots a full app against the genesis export at
// genesisPath (relative to the calling test's package directory),
// finalizing the first block at time.Now() — the same way a real node
// resuming this chain today would. Any unbonding delegation whose
// completion_time has already elapsed by "now" matures and pays out right
// here, as part of this very first block's EndBlock, before the caller gets
// a chance to inspect it as "pending". That's expected real-world behavior,
// not a bug: overdue unbondings are settled at genesis load, and rotation
// only needs to deal with whatever is still in flight afterwards.
// Skips the test (rather than failing) if the file isn't present.
//
// Exported (and not a _test.go file) so app/migrations' and
// app/upgrades/v1.8's tests — separate packages, since both are imported BY
// app and so can't import app back from an internal test — can call it too.
// See migrations.Keepers.
func SetupWithGenesisFile(t *testing.T, genesisPath string) (realioApp *RealioNetwork, chainID string, initialHeight int64, proposerAddr []byte, blockTime time.Time) {
	t.Helper()

	raw, err := os.ReadFile(genesisPath) //nolint:staticcheck // SA4006 false positive: raw is read at json.Unmarshal(raw, &doc) below
	if err != nil {
		t.Skipf("real genesis fixture not present at %s, skipping: %v", genesisPath, err)
		return nil, "", 0, nil, time.Time{}
	}

	// The EVM module keeps its coin-info config in a process-global
	// sync.Once (github.com/cosmos/evm/x/vm.SetGlobalConfigVariables), same
	// as Setup() in test_helpers.go deals with — must reset it before this
	// InitChain, or InitGenesis panics if any other test in this binary
	// already initialized it first.
	evmtypes.NewEVMConfigurator().ResetTestConfig()

	var doc struct {
		ChainID       string                     `json:"chain_id"`
		InitialHeight int64                      `json:"initial_height"`
		AppState      map[string]json.RawMessage `json:"app_state"`
		Consensus     struct {
			Validators []ConsensusValidatorEntry `json:"validators"`
		} `json:"consensus"`
	}
	require.NoError(t, json.Unmarshal(raw, &doc))
	require.NotEmpty(t, doc.Consensus.Validators)
	require.NotZero(t, doc.InitialHeight)

	proposerAddr, err = hex.DecodeString(doc.Consensus.Validators[0].Address)
	require.NoError(t, err)

	appStateBytes, err := json.Marshal(doc.AppState)
	require.NoError(t, err)

	db := dbm.NewMemDB()
	realioApp = New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, simtestutil.EmptyAppOptions{},
		baseapp.SetChainID(doc.ChainID))

	_, err = realioApp.InitChain(&abci.RequestInitChain{
		ChainId:         doc.ChainID,
		InitialHeight:   doc.InitialHeight,
		ConsensusParams: DefaultConsensusParams,
		AppStateBytes:   appStateBytes,
	})
	require.NoError(t, err)

	blockTime = time.Now()

	// This genesis's initial_height is the real chain's actual halt height —
	// which may legitimately equal BlacklistForkHeight's current production
	// value (it does today: both are 19573266, since this genesis IS the
	// real halt-block export). Neutralize every height-triggered fork for
	// the duration of this one FinalizeBlock call, so none of them fire
	// here by coincidence before a test gets a chance to control it
	// explicitly. Callers that want to exercise a fork opt in afterwards by
	// setting the relevant height themselves and driving
	// BeginBlocker/EndBlocker directly, same as the rest of this file does.
	origBlacklistHeight := migrations.BlacklistForkHeight
	migrations.BlacklistForkHeight = -1
	defer func() {
		migrations.BlacklistForkHeight = origBlacklistHeight
	}()

	// Deliberately not calling Commit() here: doing so tears down
	// finalizeBlockState, and BaseApp.NewContextLegacy(false, ...) — used
	// throughout this test to build ad-hoc contexts for direct
	// BeginBlocker/EndBlocker calls, the same pattern the rest of this
	// package's tests use against Setup() — reads from exactly that state.
	// Setup() in test_helpers.go follows the same convention.
	_, err = realioApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:          doc.InitialHeight,
		ProposerAddress: proposerAddr,
		Time:            blockTime,
	})
	require.NoError(t, err)

	return realioApp, doc.ChainID, doc.InitialHeight, proposerAddr, blockTime
}

// NewHeaderCtx builds an ad-hoc sdk.Context at the given height, the same
// way BeginBlocker/EndBlocker would see it. HeaderInfo is set alongside the
// block header because parts of x/staking (e.g. the redelegation queue) read
// the time from HeaderInfo, not from BlockHeader.
func NewHeaderCtx(realioApp *RealioNetwork, height int64, proposerAddr []byte, blockTime time.Time) sdk.Context {
	return realioApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Height:          height,
		ProposerAddress: proposerAddr,
		Time:            blockTime,
	}).WithHeaderInfo(coreheader.Info{
		Height: height,
		Time:   blockTime,
	}).WithBlockGasMeter(storetypes.NewInfiniteGasMeter())
}
