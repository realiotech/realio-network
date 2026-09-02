package app

import (
	"encoding/json"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/realiotech/realio-network/testutil"
)

// TestRevokeLeakedAuthzGrants proves the authz bypass is closed: a
// blacklisted account can't act via authz.MsgExec through an outstanding
// grant, because the grant itself is gone by the time ScheduleForkUpgrade
// finishes. Seeds one grant from a leaked granter and one from a clean
// granter, and checks only the leaked one is revoked.
func TestRevokeLeakedAuthzGrants(t *testing.T) {
	realio := Setup(false, nil, 1)
	ak := realio.AuthzKeeper

	origHeight, origJSON, origAssetRotations := BlacklistForkHeight, leakedAddressesJSON, assetManagerRotations
	t.Cleanup(func() {
		BlacklistForkHeight, leakedAddressesJSON, assetManagerRotations = origHeight, origJSON, origAssetRotations
	})

	leakedGranter := testutil.GenAddress()
	cleanGranter := testutil.GenAddress()
	grantee := testutil.GenAddress()

	leakedJSON, err := json.Marshal([]string{leakedGranter.String()})
	require.NoError(t, err)
	leakedAddressesJSON = leakedJSON
	assetManagerRotations = nil
	BlacklistForkHeight = 12345

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: BlacklistForkHeight})

	msgType := sdk.MsgTypeURL(&banktypes.MsgSend{})
	require.NoError(t, ak.SaveGrant(ctx, grantee, leakedGranter, authz.NewGenericAuthorization(msgType), nil))
	require.NoError(t, ak.SaveGrant(ctx, grantee, cleanGranter, authz.NewGenericAuthorization(msgType), nil))

	before, err := ak.GetAuthorizations(ctx, grantee, leakedGranter)
	require.NoError(t, err)
	require.Len(t, before, 1, "sanity: leaked granter's grant must exist before the fork runs")

	realio.ScheduleForkUpgrade(ctx)

	afterLeaked, err := ak.GetAuthorizations(ctx, grantee, leakedGranter)
	require.NoError(t, err)
	require.Empty(t, afterLeaked, "grant from a leaked (blacklisted) granter must be revoked by the fork")

	afterClean, err := ak.GetAuthorizations(ctx, grantee, cleanGranter)
	require.NoError(t, err)
	require.Len(t, afterClean, 1, "grant from a non-leaked granter must be untouched")
}

// TestRevokeLeakedAuthzGrantsAgainstRealGenesis is the end-to-end test
// against the real pre-incident genesis export: confirms the real genesis
// actually has authz grants whose granter is on leaked_addresses.json (the
// exact gap this fork closes — a leaked account's authz grantee could
// otherwise still act on its behalf via MsgExec after the halt, bypassing
// blacklist entirely), then confirms EVERY one of them individually is gone
// after the real BeginBlocker runs the fork — not just that the count
// reaches zero, which a bug that revoked the wrong grants could also
// produce — while every grant from a clean (non-leaked) granter survives
// untouched, so a check this thorough can't be satisfied by over-deleting.
func TestRevokeLeakedAuthzGrantsAgainstRealGenesis(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := SetupWithRealGenesis(t)
	ak := realioApp.AuthzKeeper

	origHeight := BlacklistForkHeight
	t.Cleanup(func() { BlacklistForkHeight = origHeight })

	rotationHeight := initialHeight + 1
	BlacklistForkHeight = rotationHeight

	leaked := make(map[string]bool, 512)
	for _, addr := range parseLeakedAddresses() {
		leaked[addr] = true
	}

	type grantKey struct {
		granter, grantee, msgType string
	}
	baseCtx := newHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)
	var leakedGrants, cleanGrants []grantKey
	ak.IterateGrants(baseCtx, func(granterAddr, granteeAddr sdk.AccAddress, grant authz.Grant) bool {
		auth, err := grant.GetAuthorization()
		require.NoError(t, err)
		k := grantKey{granterAddr.String(), granteeAddr.String(), auth.MsgTypeURL()}
		if leaked[k.granter] {
			leakedGrants = append(leakedGrants, k)
		} else {
			cleanGrants = append(cleanGrants, k)
		}
		return false
	})
	require.NotEmptyf(t, leakedGrants, "sanity: the real genesis must have at least one authz grant from a leaked granter")
	require.NotEmptyf(t, cleanGrants, "sanity: the real genesis must have at least one authz grant from a non-leaked granter, or the survival check below is vacuous")
	t.Logf("real genesis has %d authz grant(s) from a leaked granter and %d from a clean one, before the fork", len(leakedGrants), len(cleanGrants))

	ctx := newHeaderCtx(realioApp, rotationHeight, proposerAddr, baseCtx.BlockTime())
	_, err := realioApp.BeginBlocker(ctx)
	require.NoError(t, err)

	for _, k := range leakedGrants {
		granter, err := sdk.AccAddressFromBech32(k.granter)
		require.NoError(t, err)
		grantee, err := sdk.AccAddressFromBech32(k.grantee)
		require.NoError(t, err)
		auth, _ := ak.GetAuthorization(ctx, grantee, granter, k.msgType)
		require.Nilf(t, auth, "grant granter=%s grantee=%s msgType=%s must be revoked, individually, not just missing from a count",
			k.granter, k.grantee, k.msgType)
	}

	for _, k := range cleanGrants {
		granter, err := sdk.AccAddressFromBech32(k.granter)
		require.NoError(t, err)
		grantee, err := sdk.AccAddressFromBech32(k.grantee)
		require.NoError(t, err)
		auth, _ := ak.GetAuthorization(ctx, grantee, granter, k.msgType)
		require.NotNilf(t, auth, "clean grant granter=%s grantee=%s msgType=%s must survive untouched", k.granter, k.grantee, k.msgType)
	}

	var totalAfter int
	ak.IterateGrants(ctx, func(sdk.AccAddress, sdk.AccAddress, authz.Grant) bool {
		totalAfter++
		return false
	})
	require.Equalf(t, len(cleanGrants), totalAfter,
		"total grant count after the fork must equal exactly the clean count — anything else means either a leaked grant survived or a clean one was wrongly deleted")
}
