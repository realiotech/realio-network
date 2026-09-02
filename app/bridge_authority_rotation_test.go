package app

import (
	"encoding/json"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/realiotech/realio-network/testutil"
	bridgetypes "github.com/realiotech/realio-network/x/bridge/types"
)

// TestRotateBridgeAuthority exercises ScheduleForkUpgrade end-to-end at
// BlacklistForkHeight: seeds a real bridge Params record with an old
// authority, runs the real dispatch path, and checks Authority actually
// moved to the replacement address.
func TestRotateBridgeAuthority(t *testing.T) {
	realio := Setup(false, nil, 1)
	bk := realio.BridgeKeeper

	origHeight, origAuthority, origJSON, origAssetRotations := BlacklistForkHeight, BridgeAuthority, leakedAddressesJSON, assetManagerRotations
	t.Cleanup(func() {
		BlacklistForkHeight, BridgeAuthority, leakedAddressesJSON, assetManagerRotations = origHeight, origAuthority, origJSON, origAssetRotations
	})

	// isolate from the real leaked_addresses.json embed and from
	// rotateAssetManagers, which also run at BlacklistForkHeight (see
	// forks.go) - this test is about the bridge authority, not those.
	leakedAddressesJSON, _ = json.Marshal([]string{})
	assetManagerRotations = nil
	BlacklistForkHeight = 12345

	oldAuthority := testutil.GenAddress().String()
	newAuthority := testutil.GenAddress().String()
	BridgeAuthority = newAuthority

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: BlacklistForkHeight})

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
	realioApp, _, initialHeight, proposerAddr, blockTime := SetupWithRealGenesis(t)
	bk := realioApp.BridgeKeeper

	origHeight := BlacklistForkHeight
	t.Cleanup(func() { BlacklistForkHeight = origHeight })

	rotationHeight := initialHeight + 1
	BlacklistForkHeight = rotationHeight

	baseCtx := newHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)
	before, err := bk.Params.Get(baseCtx)
	require.NoError(t, err)
	require.NotEqual(t, BridgeAuthority, before.Authority,
		"sanity: the real genesis's authority should not already be the replacement address")

	ctx := newHeaderCtx(realioApp, rotationHeight, proposerAddr, baseCtx.BlockTime())
	_, err = realioApp.BeginBlocker(ctx)
	require.NoError(t, err)

	after, err := bk.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, BridgeAuthority, after.Authority, "authority must have moved to the replacement address")
	require.NotEqual(t, before.Authority, after.Authority)
}
