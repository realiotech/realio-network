package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/realiotech/realio-network/app/migrations"
	v6 "github.com/realiotech/realio-network/app/upgrades/v1.6"
	blacklisttypes "github.com/realiotech/realio-network/x/blacklist/types"
)

// Exercise the actual registration path with skip heights that previously
// returned before installing the hardcoded blacklist loader.
func TestBlacklistStoreLoaderIgnoresSkipHeight(t *testing.T) {
	const forkHeight = int64(5)
	originalHeight := migrations.BlacklistForkHeight
	migrations.BlacklistForkHeight = forkHeight
	t.Cleanup(func() { migrations.BlacklistForkHeight = originalHeight })

	for _, tc := range []struct {
		name       string
		skipHeight int64
		writeInfo  bool
	}{
		{name: "no upgrade info with zero skip height", skipHeight: 0},
		{name: "stale skipped v6 upgrade", skipHeight: 3, writeInfo: true},
		{name: "skipped upgrade at fork height", skipHeight: forkHeight, writeInfo: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			evmtypes.NewEVMConfigurator().ResetTestConfig()
			home := t.TempDir()
			if tc.writeInfo {
				info, err := json.Marshal(upgradetypes.Plan{Name: v6.UpgradeName, Height: tc.skipHeight})
				require.NoError(t, err)
				require.NoError(t, os.MkdirAll(filepath.Join(home, "data"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(home, "data", upgradetypes.UpgradeInfoFilename), info, 0o600))
			}

			db := dbm.NewMemDB()
			logger := log.NewTestLogger(t)
			// Construct without loading so the old database can be seeded first.
			app := New(logger, db, nil, false, map[int64]bool{tc.skipHeight: true}, home, 0, simtestutil.EmptyAppOptions{})
			oldApp := baseapp.NewBaseApp(t.Name(), logger, db, nil)
			for name := range app.keys {
				if name != blacklisttypes.StoreKey {
					oldApp.MountStores(storetypes.NewKVStoreKey(name))
				}
			}
			require.NoError(t, oldApp.LoadLatestVersion())
			for height := int64(1); height < forkHeight; height++ {
				_, err := oldApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: height})
				require.NoError(t, err)
				_, err = oldApp.Commit()
				require.NoError(t, err)
			}

			require.NoError(t, app.LoadLatestVersion())
			require.Equal(t, forkHeight-1, app.LastBlockHeight())
		})
	}
}

// TestNewStoreLoaderAddsNewStore reproduces the exact scenario
// newStoreLoader exists for: a chain database that predates x/blacklist
// (i.e. never had that store), restarting into a binary that has it
// registered. Without a matching StoreUpgrades entry, LoadLatestVersion
// panics ("version of store blacklist mismatch..." — the error this was
// built to fix). This builds the old/new app pair directly against a shared
// MemDB, the same way cosmos-sdk's own storeloader_test.go verifies
// UpgradeStoreLoader, rather than going through app.Setup() (which always
// starts from a fresh genesis that already has every store, and so never
// actually exercises this code path).
func TestNewStoreLoaderAddsNewStore(t *testing.T) {
	const upgradeHeight = int64(5)
	const oldStoreKey = "foo"
	const newStoreKey = "blacklist"

	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	pruneOpt := baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))

	// Old binary: only knows about "foo", commits up to upgradeHeight-1.
	oldApp := baseapp.NewBaseApp(t.Name(), logger.With("instance", "old"), db, nil, pruneOpt)
	oldApp.MountStores(storetypes.NewKVStoreKey(oldStoreKey))
	require.NoError(t, oldApp.LoadLatestVersion())
	require.Equal(t, int64(0), oldApp.LastBlockHeight())

	for i := int64(1); i <= upgradeHeight-1; i++ {
		_, err := oldApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: i})
		require.NoError(t, err)
		_, err = oldApp.Commit()
		require.NoError(t, err)
	}
	require.Equal(t, upgradeHeight-1, oldApp.LastBlockHeight())

	// New binary: knows about "foo" AND "blacklist", wired through
	// newStoreLoader exactly as setupUpgradeHandlers does.
	candidates := []heightStoreUpgrade{
		{height: upgradeHeight, upgrades: storetypes.StoreUpgrades{Added: []string{newStoreKey}}},
	}
	newApp := baseapp.NewBaseApp(t.Name(), logger.With("instance", "new"), db, nil,
		pruneOpt, baseapp.SetStoreLoader(newStoreLoader(candidates)))
	newApp.MountStores(storetypes.NewKVStoreKey(oldStoreKey), storetypes.NewKVStoreKey(newStoreKey))

	// This is the line that panics today without a matching StoreUpgrades:
	// "failed to load latest version: version of store blacklist mismatch
	// root store's version; expected <N> got 0".
	require.NoError(t, newApp.LoadLatestVersion())
	require.Equal(t, upgradeHeight-1, newApp.LastBlockHeight())

	// "Execute" the upgrade block itself.
	_, err := newApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: upgradeHeight})
	require.NoError(t, err)
	_, err = newApp.Commit()
	require.NoError(t, err)
	require.Equal(t, upgradeHeight, newApp.LastBlockHeight())

	// Restarting yet again, later, at a height that matches no candidate,
	// must be an ordinary no-op (DefaultStoreLoader) — not re-trigger the
	// upgrade and not error.
	laterApp := baseapp.NewBaseApp(t.Name(), logger.With("instance", "later"), db, nil,
		pruneOpt, baseapp.SetStoreLoader(newStoreLoader(candidates)))
	laterApp.MountStores(storetypes.NewKVStoreKey(oldStoreKey), storetypes.NewKVStoreKey(newStoreKey))
	require.NoError(t, laterApp.LoadLatestVersion())
	require.Equal(t, upgradeHeight, laterApp.LastBlockHeight())
}
