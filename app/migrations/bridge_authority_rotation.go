package migrations

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BridgeAuthority replaces x/bridge's on-chain Params.Authority — the one
// and only signer MsgBridgeIn and MsgBridgeOut will accept. Runs alongside
// seedLeakedAddressBlacklist at BlacklistForkHeight (see forks.go): the old
// authority, realio1yxmgj2rp86xt4lrw4xzuszqqzff2lelewm99ft, is in
// leaked_addresses.json. That alone would make it unusable (blacklisting
// stops it from ever signing another tx), but leaves nobody able to
// legitimately call MsgBridgeIn/MsgBridgeOut going forward — the same shape
// of problem asset-manager rotation solves for a compromised token manager.
// Rotating Params.Authority restores that.
//
// MsgBridgeIn in particular is why this is higher-stakes than it looks: its
// only gate is msg.Authority == param.Authority, and on success it mints
// new coins outright (x/bridge/keeper/msg_server_bridge.go) — so the old,
// leaked authority is a live unlimited-mint key until this rotates, not
// just an inconvenience.
var BridgeAuthority = "realio1usws4ccu3m7ln6m5gesc3m8jqcl0fqagq32kv6"

func rotateBridgeAuthority(k Keepers, ctx sdk.Context) {
	params, err := k.BridgeKeeper.Params.Get(ctx)
	if err != nil {
		panic(fmt.Errorf("bridge authority rotation: failed to get params: %w", err))
	}

	params.Authority = BridgeAuthority
	if err := params.Validate(); err != nil {
		panic(fmt.Errorf("bridge authority rotation: invalid new authority %q: %w", BridgeAuthority, err))
	}

	if err := k.BridgeKeeper.Params.Set(ctx, params); err != nil {
		panic(fmt.Errorf("bridge authority rotation: failed to set params: %w", err))
	}
}
