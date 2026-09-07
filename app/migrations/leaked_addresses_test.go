package migrations_test

import (
	"encoding/json"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/app/migrations"
	"github.com/realiotech/realio-network/testutil"
)

// TestLeakedAddressesJSONIsValid parses leaked_addresses.json exactly the
// way seedLeakedAddressBlacklist does. A malformed entry here would
// otherwise only surface as a panic inside BeginBlocker at
// BlacklistForkHeight — i.e. a chain halt on every validator at the worst
// possible time. Catching it here turns that into a normal CI failure
// before the binary is ever built.
func TestLeakedAddressesJSONIsValid(t *testing.T) {
	addrs := migrations.ParseLeakedAddresses()

	seen := make(map[string]int, len(addrs))
	for i, addr := range addrs {
		_, err := sdk.AccAddressFromBech32(addr)
		require.NoErrorf(t, err, "leaked_addresses.json entry %d: %q is not a valid bech32 address", i, addr)

		seen[addr]++
	}

	for addr, count := range seen {
		require.Equalf(t, 1, count, "leaked_addresses.json contains %q %d times", addr, count)
	}
}

// TestBlacklistFork exercises ScheduleForkUpgrade end-to-end at
// BlacklistForkHeight: it swaps in a throwaway JSON list/height, runs the
// real dispatch path, and checks the address actually lands in
// x/blacklist.
func TestBlacklistFork(t *testing.T) {
	realio := app.Setup(false, nil, 1)

	origJSON, origHeight, origAssetRotations := migrations.LeakedAddressesJSON, migrations.BlacklistForkHeight, migrations.AssetManagerRotations
	t.Cleanup(func() {
		migrations.LeakedAddressesJSON, migrations.BlacklistForkHeight, migrations.AssetManagerRotations = origJSON, origHeight, origAssetRotations
	})

	// isolate from rotateAssetManagers, which also runs at BlacklistForkHeight
	// (see forks.go) - this test is about the blacklist seeding, not that.
	migrations.AssetManagerRotations = nil

	leaked := testutil.GenAddress()
	untouched := testutil.GenAddress()

	migrations.LeakedAddressesJSON, _ = json.Marshal([]string{leaked.String()})
	migrations.BlacklistForkHeight = 12345

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: migrations.BlacklistForkHeight})
	require.False(t, realio.BlacklistKeeper.IsBlacklisted(ctx, leaked))

	realio.ScheduleForkUpgrade(ctx)

	require.True(t, realio.BlacklistKeeper.IsBlacklisted(ctx, leaked))
	require.False(t, realio.BlacklistKeeper.IsBlacklisted(ctx, untouched))

	// Re-running at a different height (the normal, non-fork case) must
	// not re-trigger the seed logic or otherwise touch the blacklist.
	otherCtx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: migrations.BlacklistForkHeight + 1})
	extra := testutil.GenAddress()
	migrations.LeakedAddressesJSON, _ = json.Marshal([]string{extra.String()})
	realio.ScheduleForkUpgrade(otherCtx)
	require.False(t, realio.BlacklistKeeper.IsBlacklisted(otherCtx, extra))
}
