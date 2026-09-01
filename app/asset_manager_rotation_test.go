package app

import (
	"encoding/json"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/realiotech/realio-network/testutil"
	assetmoduletypes "github.com/realiotech/realio-network/x/asset/types"
)

// TestRotateAssetManagers exercises ScheduleForkUpgrade end-to-end at
// BlacklistForkHeight: seeds a real rst Token record with an old manager,
// runs the real dispatch path, and checks the Manager field actually moved
// to the new address — with everything else about the token (name, total
// supply, authorization flag) untouched.
func TestRotateAssetManagers(t *testing.T) {
	realio := Setup(false, nil, 1)
	ak := realio.AssetKeeper

	origHeight, origRotations, origJSON := BlacklistForkHeight, assetManagerRotations, leakedAddressesJSON
	t.Cleanup(func() {
		BlacklistForkHeight, assetManagerRotations, leakedAddressesJSON = origHeight, origRotations, origJSON
	})

	// isolate from the real leaked_addresses.json embed - not what this test is about
	leakedAddressesJSON, _ = json.Marshal([]string{})
	BlacklistForkHeight = 12345

	oldManager := testutil.GenAddress().String()
	newManager := testutil.GenAddress().String()
	assetManagerRotations = []struct {
		Symbol     string
		NewManager string
	}{
		{Symbol: "rst", NewManager: newManager},
	}

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: BlacklistForkHeight})

	require.NoError(t, ak.Token.Set(ctx, assetmoduletypes.TokenKey("rst"), assetmoduletypes.Token{
		Name:                  "Realio Security Token",
		Symbol:                "rst",
		Total:                 "1000000",
		AuthorizationRequired: true,
		Manager:               oldManager,
	}))

	realio.ScheduleForkUpgrade(ctx)

	token, err := ak.Token.Get(ctx, assetmoduletypes.TokenKey("rst"))
	require.NoError(t, err)
	require.Equal(t, newManager, token.Manager, "manager must move to the replacement address")
	require.Equal(t, "Realio Security Token", token.Name, "everything else about the token must stay untouched")
	require.Equal(t, "1000000", token.Total)
	require.True(t, token.AuthorizationRequired)
}

// TestRotateAssetManagersPanicsOnMissingToken proves a typo'd or
// not-yet-created symbol in assetManagerRotations fails loudly at fork time
// (chain halt, caught immediately) rather than silently skipping a manager
// rotation that was supposed to happen.
func TestRotateAssetManagersPanicsOnMissingToken(t *testing.T) {
	realio := Setup(false, nil, 1)

	origRotations := assetManagerRotations
	t.Cleanup(func() { assetManagerRotations = origRotations })

	assetManagerRotations = []struct {
		Symbol     string
		NewManager string
	}{
		{Symbol: "does-not-exist", NewManager: testutil.GenAddress().String()},
	}

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1})
	require.Panics(t, func() { rotateAssetManagers(realio, ctx) })
}

// TestRotateAssetManagersAgainstRealGenesis is the end-to-end test against
// the real pre-incident genesis export (see validator_rotation_genesis_test.go
// for SetupWithRealGenesis/newHeaderCtx): InitChain with the real state
// (real rst/lmx Token records, real old managers), run the real fork
// dispatch through BeginBlocker, and confirm the Manager field on each
// token actually moved to its replacement address.
func TestRotateAssetManagersAgainstRealGenesis(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := SetupWithRealGenesis(t)
	ak := realioApp.AssetKeeper

	origHeight := BlacklistForkHeight
	t.Cleanup(func() { BlacklistForkHeight = origHeight })

	rotationHeight := initialHeight + 1
	BlacklistForkHeight = rotationHeight

	baseCtx := newHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	oldManagers := make(map[string]string, len(assetManagerRotations))
	for _, r := range assetManagerRotations {
		token, err := ak.Token.Get(baseCtx, assetmoduletypes.TokenKey(r.Symbol))
		require.NoErrorf(t, err, "expected token %q to exist in the real genesis", r.Symbol)
		require.NotEqualf(t, r.NewManager, token.Manager,
			"token %q's manager in the real genesis should not already be the replacement address", r.Symbol)
		oldManagers[r.Symbol] = token.Manager
	}
	require.NotEmpty(t, oldManagers, "expected at least one real rst/lmx token in the genesis to rotate")

	ctx := newHeaderCtx(realioApp, rotationHeight, proposerAddr, baseCtx.BlockTime())
	_, err := realioApp.BeginBlocker(ctx)
	require.NoError(t, err)

	for _, r := range assetManagerRotations {
		token, err := ak.Token.Get(ctx, assetmoduletypes.TokenKey(r.Symbol))
		require.NoError(t, err)
		require.Equalf(t, r.NewManager, token.Manager, "token %q's manager must have moved to the replacement address", r.Symbol)
		require.NotEqual(t, oldManagers[r.Symbol], token.Manager)
	}
}
