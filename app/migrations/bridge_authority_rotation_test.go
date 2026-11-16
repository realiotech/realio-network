package migrations_test

import (
	"encoding/json"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/app/migrations"
	"github.com/realiotech/realio-network/testutil"
	bridgetypes "github.com/realiotech/realio-network/x/bridge/types"
)

// TestRotateBridgeAuthority exercises ScheduleForkUpgrade end-to-end at
// migrations.BlacklistForkHeight: seeds a real bridge Params record with an old
// authority, runs the real dispatch path, and checks Authority actually
// moved to the replacement address.
func TestRotateBridgeAuthority(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	bk := realio.BridgeKeeper

	origHeight, origAuthority, origJSON, origAssetRotations := migrations.BlacklistForkHeight, migrations.BridgeAuthority, migrations.LeakedAddressesJSON, migrations.AssetManagerRotations
	t.Cleanup(func() {
		migrations.BlacklistForkHeight, migrations.BridgeAuthority, migrations.LeakedAddressesJSON, migrations.AssetManagerRotations = origHeight, origAuthority, origJSON, origAssetRotations
	})

	// isolate from the real leaked_addresses.json embed and from
	// rotateAssetManagers, which also run at migrations.BlacklistForkHeight (see
	// forks.go) - this test is about the bridge authority, not those.
	migrations.LeakedAddressesJSON, _ = json.Marshal([]string{})
	migrations.AssetManagerRotations = nil
	migrations.BlacklistForkHeight = 12345

	oldAuthority := testutil.GenAddress().String()
	newAuthority := testutil.GenAddress().String()
	migrations.BridgeAuthority = newAuthority

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: migrations.BlacklistForkHeight})

	require.NoError(t, bk.Params.Set(ctx, bridgetypes.NewParams(oldAuthority)))

	realio.ScheduleForkUpgrade(ctx)

	params, err := bk.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, newAuthority, params.Authority, "authority must move to the replacement address")
}

// TestRotateBridgeAuthorityAgainstRealGenesis is the end-to-end test against
// the real pre-incident genesis export: real x/bridge Params, real old
// authority (confirmed to be on leaked_addresses.json), run through the
// real BeginBlocker path.
func TestRotateBridgeAuthorityAgainstRealGenesis(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := app.SetupWithRealGenesis(t)
	bk := realioApp.BridgeKeeper

	origHeight := migrations.BlacklistForkHeight
	t.Cleanup(func() { migrations.BlacklistForkHeight = origHeight })

	rotationHeight := initialHeight + 1
	migrations.BlacklistForkHeight = rotationHeight

	baseCtx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)
	before, err := bk.Params.Get(baseCtx)
	require.NoError(t, err)
	require.NotEqual(t, migrations.BridgeAuthority, before.Authority,
		"sanity: the real genesis's authority should not already be the replacement address")

	ctx := app.NewHeaderCtx(realioApp, rotationHeight, proposerAddr, baseCtx.BlockTime())
	_, err = realioApp.BeginBlocker(ctx)
	require.NoError(t, err)

	after, err := bk.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, migrations.BridgeAuthority, after.Authority, "authority must have moved to the replacement address")
	require.NotEqual(t, before.Authority, after.Authority)
}
