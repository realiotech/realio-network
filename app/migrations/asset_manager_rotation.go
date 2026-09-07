package migrations

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	assetmoduletypes "github.com/realiotech/realio-network/x/asset/types"
)

// AssetManagerRotations lists each asset token's replacement manager
// (called "issuer" by token holders), keyed by symbol. Runs alongside
// seedLeakedAddressBlacklist at BlacklistForkHeight (see forks.go) — the old
// manager keys for these tokens are believed to have leaked along with the
// validator keys. A slice, not a map: this runs in BeginBlocker and must
// stay deterministic across every validator, though in this particular case
// each entry only touches its own symbol so iteration order wouldn't
// actually change the result.
var AssetManagerRotations = []struct {
	Symbol     string
	NewManager string
}{
	{Symbol: "rst", NewManager: "realio1k8arezrrlq7zwv66cxn47kw0m9hvzufjcmtmxc"},
	{Symbol: "lmx", NewManager: "realio1ewn7ftvnjyyep9w2x2jz3k977p3ys6e22c2hzj"},
}

// rotateAssetManagers replaces the Manager field on each token in
// AssetManagerRotations. Nothing else about the token (name, symbol, total
// supply, authorization list) changes — only who can administer it going
// forward. x/asset's own MsgUpdateToken deliberately never lets the manager
// field itself change (it requires the CURRENT manager's signature and
// always re-writes Manager: existing.Manager), so this can only be done via
// a fork touching keeper state directly, the same as validator rotation.
func rotateAssetManagers(k Keepers, ctx sdk.Context) {
	for _, r := range AssetManagerRotations {
		key := assetmoduletypes.TokenKey(r.Symbol)
		token, err := k.AssetKeeper.Token.Get(ctx, key)
		if err != nil {
			panic(fmt.Errorf("asset manager rotation: token %q not found: %w", r.Symbol, err))
		}
		token.Manager = r.NewManager
		if err := k.AssetKeeper.Token.Set(ctx, key, token); err != nil {
			panic(fmt.Errorf("asset manager rotation: failed to set token %q: %w", r.Symbol, err))
		}
	}
}

// unauthorizeLeakedAddresses closes a gap blacklisting alone doesn't cover:
// x/blacklist stops a leaked address from ever signing another tx (it can't
// send anything out), but AssetSendRestriction (x/asset/keeper/restrictions.go)
// checks BOTH sides of a transfer of an AuthorizationRequired token — so as
// long as a leaked address stays "authorized", anyone else can still send
// rst/lmx TO it, and those funds land in an account whose key is
// compromised. This walks every address in leaked_addresses.json and, for
// each token in AssetManagerRotations (rst, lmx — the same two this
// incident already touches), un-authorizes it if it currently holds
// authorized status. Reuses AssetManagerRotations' symbol list rather than
// a separate hardcoded one so there's a single source of truth for "which
// tokens this incident affects".
func unauthorizeLeakedAddresses(k Keepers, ctx sdk.Context, leaked []sdk.AccAddress) {
	for _, r := range AssetManagerRotations {
		key := assetmoduletypes.TokenKey(r.Symbol)
		token, err := k.AssetKeeper.Token.Get(ctx, key)
		if err != nil {
			panic(fmt.Errorf("unauthorize leaked addresses: token %q not found: %w", r.Symbol, err))
		}

		for _, addr := range leaked {
			if token.AddressIsAuthorized(addr) {
				token.UnAuthorizeAddress(addr)
			}
		}
		if err := k.AssetKeeper.Token.Set(ctx, key, token); err != nil {
			panic(fmt.Errorf("unauthorize leaked addresses: failed to set token %q: %w", r.Symbol, err))
		}
	}
}
