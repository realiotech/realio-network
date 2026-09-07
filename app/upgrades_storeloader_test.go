package app

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

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
