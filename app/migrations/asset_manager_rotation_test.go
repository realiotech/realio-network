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
	assetmoduletypes "github.com/realiotech/realio-network/x/asset/types"
)

// TestRotateAssetManagers exercises ScheduleForkUpgrade end-to-end at
// BlacklistForkHeight: seeds a real rst Token record with an old manager,
// runs the real dispatch path, and checks the Manager field actually moved
// to the new address — with everything else about the token (name, total
// supply, authorization flag) untouched.
func TestRotateAssetManagers(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ak := realio.AssetKeeper

	origHeight, origRotations, origJSON := migrations.BlacklistForkHeight, migrations.AssetManagerRotations, migrations.LeakedAddressesJSON
	t.Cleanup(func() {
		migrations.BlacklistForkHeight, migrations.AssetManagerRotations, migrations.LeakedAddressesJSON = origHeight, origRotations, origJSON
	})

	// isolate from the real leaked_addresses.json embed - not what this test is about
	migrations.LeakedAddressesJSON, _ = json.Marshal([]string{})
	migrations.BlacklistForkHeight = 12345

	oldManager := testutil.GenAddress().String()
	newManager := testutil.GenAddress().String()
	migrations.AssetManagerRotations = []struct {
		Symbol     string
		NewManager string
	}{
		{Symbol: "rst", NewManager: newManager},
	}

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: migrations.BlacklistForkHeight})

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
// not-yet-created symbol in AssetManagerRotations fails loudly at fork time
// (chain halt, caught immediately) rather than silently skipping a manager
// rotation that was supposed to happen.
func TestRotateAssetManagersPanicsOnMissingToken(t *testing.T) {
	realio := app.Setup(false, nil, 1)

	origHeight, origRotations, origJSON := migrations.BlacklistForkHeight, migrations.AssetManagerRotations, migrations.LeakedAddressesJSON
	t.Cleanup(func() {
		migrations.BlacklistForkHeight, migrations.AssetManagerRotations, migrations.LeakedAddressesJSON = origHeight, origRotations, origJSON
	})

	migrations.LeakedAddressesJSON, _ = json.Marshal([]string{})
	migrations.BlacklistForkHeight = 1
	migrations.AssetManagerRotations = []struct {
		Symbol     string
		NewManager string
	}{
		{Symbol: "does-not-exist", NewManager: testutil.GenAddress().String()},
	}

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1})
	require.Panics(t, func() { realio.ScheduleForkUpgrade(ctx) })
}

// TestUnauthorizeLeakedAddresses exercises ScheduleForkUpgrade end-to-end at
// BlacklistForkHeight: a leaked address that's currently authorized on rst
// must come out unauthorized; a leaked address that was never authorized in
// the first place must be a no-op, not a panic; and an authorized address
// that ISN'T leaked must be left alone entirely — this only closes the
// receiving side for addresses actually on the leak list, not a blanket
// wipe of every authorization.
func TestUnauthorizeLeakedAddresses(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ak := realio.AssetKeeper

	origHeight, origJSON, origRotations := migrations.BlacklistForkHeight, migrations.LeakedAddressesJSON, migrations.AssetManagerRotations
	t.Cleanup(func() {
		migrations.BlacklistForkHeight, migrations.LeakedAddressesJSON, migrations.AssetManagerRotations = origHeight, origJSON, origRotations
	})

	leakedAuthorized := testutil.GenAddress()
	leakedNeverAuthorized := testutil.GenAddress()
	notLeakedAuthorized := testutil.GenAddress()

	migrations.LeakedAddressesJSON, _ = json.Marshal([]string{leakedAuthorized.String(), leakedNeverAuthorized.String()})
	migrations.BlacklistForkHeight = 12345
	migrations.AssetManagerRotations = []struct {
		Symbol     string
		NewManager string
	}{
		{Symbol: "rst", NewManager: testutil.GenAddress().String()},
	}

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: migrations.BlacklistForkHeight})

	token := assetmoduletypes.Token{
		Name:                  "Realio Security Token",
		Symbol:                "rst",
		Total:                 "1000000",
		AuthorizationRequired: true,
		Manager:               testutil.GenAddress().String(),
	}
	token.AuthorizeAddress(leakedAuthorized)
	token.AuthorizeAddress(notLeakedAuthorized)
	require.NoError(t, ak.Token.Set(ctx, assetmoduletypes.TokenKey("rst"), token))

	realio.ScheduleForkUpgrade(ctx)

	got, err := ak.Token.Get(ctx, assetmoduletypes.TokenKey("rst"))
	require.NoError(t, err)
	require.False(t, got.AddressIsAuthorized(leakedAuthorized), "leaked+authorized address must be unauthorized")
	require.True(t, got.AddressIsAuthorized(notLeakedAuthorized), "non-leaked authorized address must remain untouched")
	require.False(t, got.AddressIsAuthorized(leakedNeverAuthorized), "leaked-but-never-authorized address stays unauthorized (no-op, no panic)")
}

// TestUnauthorizeLeakedAddressesAgainstRealGenesis is the end-to-end test
// against the real pre-incident genesis export: real rst/lmx authorized
// lists, real leaked_addresses.json, run through the real BeginBlocker
// path. Picks one real leaked-and-authorized address per token (confirmed
// by direct inspection of recover_genesis.json / leaked_addresses.json) and
// one real authorized-but-not-leaked address per token, to prove the
// distinction actually holds against production data, not just synthetic
// fixtures.
func TestUnauthorizeLeakedAddressesAgainstRealGenesis(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := app.SetupWithRealGenesis(t)
	ak := realioApp.AssetKeeper

	origHeight := migrations.BlacklistForkHeight
	t.Cleanup(func() { migrations.BlacklistForkHeight = origHeight })

	rotationHeight := initialHeight + 1
	migrations.BlacklistForkHeight = rotationHeight

	type addrCheck struct {
		symbol    string
		leaked    sdk.AccAddress
		notLeaked sdk.AccAddress
	}
	mustAddr := func(t *testing.T, bech32 string) sdk.AccAddress {
		t.Helper()
		addr, err := sdk.AccAddressFromBech32(bech32)
		require.NoError(t, err)
		return addr
	}
	checks := []addrCheck{
		{
			symbol:    "lmx",
			leaked:    mustAddr(t, "realio1hcyuatm7p5qgqwx9g4mzyw729ugg7p5xml0m3d"),
			notLeaked: mustAddr(t, "realio16kfcdc9wgd0zjta7p67dh92twhk4lvujazjs8w"),
		},
		{
			symbol:    "rst",
			leaked:    mustAddr(t, "realio1lfjhzhc69m3rzprxyqjwgrem5w9vj635hj07us"),
			notLeaked: mustAddr(t, "realio1v7q0zxsal6atgpga9k8xesrpz9gd2nxg3228my"),
		},
	}

	baseCtx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)
	for _, c := range checks {
		token, err := ak.Token.Get(baseCtx, assetmoduletypes.TokenKey(c.symbol))
		require.NoErrorf(t, err, "expected token %q to exist in the real genesis", c.symbol)
		require.Truef(t, token.AddressIsAuthorized(c.leaked), "%s: expected the picked leaked address to actually be authorized pre-fork", c.symbol)
		require.Truef(t, token.AddressIsAuthorized(c.notLeaked), "%s: expected the picked control address to actually be authorized pre-fork", c.symbol)
	}

	// Full sweep, not just the two hand-picked samples above: every single
	// leaked address that's currently authorized on either token, counted
	// before the fork so the "must all be gone after" check below actually
	// means something.
	leakedAddrs := make([]sdk.AccAddress, 0, 512)
	for _, bech32 := range migrations.ParseLeakedAddresses() {
		leakedAddrs = append(leakedAddrs, mustAddr(t, bech32))
	}
	countAuthorized := func(ctx sdk.Context, symbol string) int {
		token, err := ak.Token.Get(ctx, assetmoduletypes.TokenKey(symbol))
		require.NoError(t, err)
		n := 0
		for _, addr := range leakedAddrs {
			if token.AddressIsAuthorized(addr) {
				n++
			}
		}
		return n
	}
	beforeLmx, beforeRst := countAuthorized(baseCtx, "lmx"), countAuthorized(baseCtx, "rst")
	t.Logf("leaked addresses authorized before fork: lmx=%d rst=%d", beforeLmx, beforeRst)
	require.Greater(t, beforeLmx, 100, "sanity: expected the real genesis to have a large lmx overlap")
	require.Greater(t, beforeRst, 100, "sanity: expected the real genesis to have a large rst overlap")

	ctx := app.NewHeaderCtx(realioApp, rotationHeight, proposerAddr, baseCtx.BlockTime())
	_, err := realioApp.BeginBlocker(ctx)
	require.NoError(t, err)

	for _, c := range checks {
		token, err := ak.Token.Get(ctx, assetmoduletypes.TokenKey(c.symbol))
		require.NoError(t, err)
		require.Falsef(t, token.AddressIsAuthorized(c.leaked), "%s: leaked+authorized address must be unauthorized after the fork", c.symbol)
		require.Truef(t, token.AddressIsAuthorized(c.notLeaked), "%s: non-leaked authorized address must remain authorized after the fork", c.symbol)
	}

	afterLmx, afterRst := countAuthorized(ctx, "lmx"), countAuthorized(ctx, "rst")
	require.Zerof(t, afterLmx, "every leaked address must be unauthorized on lmx after the fork, found %d still authorized", afterLmx)
	require.Zerof(t, afterRst, "every leaked address must be unauthorized on rst after the fork, found %d still authorized", afterRst)
}

// TestRotateAssetManagersAgainstRealGenesis is the end-to-end test against
// the real pre-incident genesis export: InitChain with the real state (real
// rst/lmx Token records, real old managers), run the real fork dispatch
// through BeginBlocker, and confirm the Manager field on each token
// actually moved to its replacement address.
func TestRotateAssetManagersAgainstRealGenesis(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := app.SetupWithRealGenesis(t)
	ak := realioApp.AssetKeeper

	origHeight := migrations.BlacklistForkHeight
	t.Cleanup(func() { migrations.BlacklistForkHeight = origHeight })

	rotationHeight := initialHeight + 1
	migrations.BlacklistForkHeight = rotationHeight

	baseCtx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	oldManagers := make(map[string]string, len(migrations.AssetManagerRotations))
	for _, r := range migrations.AssetManagerRotations {
		token, err := ak.Token.Get(baseCtx, assetmoduletypes.TokenKey(r.Symbol))
		require.NoErrorf(t, err, "expected token %q to exist in the real genesis", r.Symbol)
		require.NotEqualf(t, r.NewManager, token.Manager,
			"token %q's manager in the real genesis should not already be the replacement address", r.Symbol)
		oldManagers[r.Symbol] = token.Manager
	}
	require.NotEmpty(t, oldManagers, "expected at least one real rst/lmx token in the genesis to rotate")

	ctx := app.NewHeaderCtx(realioApp, rotationHeight, proposerAddr, baseCtx.BlockTime())
	_, err := realioApp.BeginBlocker(ctx)
	require.NoError(t, err)

	for _, r := range migrations.AssetManagerRotations {
		token, err := ak.Token.Get(ctx, assetmoduletypes.TokenKey(r.Symbol))
		require.NoError(t, err)
		require.Equalf(t, r.NewManager, token.Manager, "token %q's manager must have moved to the replacement address", r.Symbol)
		require.NotEqual(t, oldManagers[r.Symbol], token.Manager)
	}
}
