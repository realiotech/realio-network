package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	assetmoduletypes "github.com/realiotech/realio-network/x/asset/types"
)

// assetManagerRotations lists each asset token's replacement manager
// (called "issuer" by token holders), keyed by symbol. Runs alongside
// seedLeakedAddressBlacklist at BlacklistForkHeight (see forks.go) — the old
// manager keys for these tokens are believed to have leaked along with the
// validator keys. A slice, not a map: this runs in BeginBlocker and must
// stay deterministic across every validator, though in this particular case
// each entry only touches its own symbol so iteration order wouldn't
// actually change the result — kept as a slice anyway to match the same
// discipline as validatorRotations.
var assetManagerRotations = []struct {
	Symbol     string
	NewManager string
}{
	{Symbol: "rst", NewManager: "realio1k8arezrrlq7zwv66cxn47kw0m9hvzufjcmtmxc"},
	{Symbol: "lmx", NewManager: "realio1ewn7ftvnjyyep9w2x2jz3k977p3ys6e22c2hzj"},
}

// rotateAssetManagers replaces the Manager field on each token in
// assetManagerRotations. Nothing else about the token (name, symbol, total
// supply, authorization list) changes — only who can administer it going
// forward. x/asset's own MsgUpdateToken deliberately never lets the manager
// field itself change (it requires the CURRENT manager's signature and
// always re-writes Manager: existing.Manager), so this can only be done via
// a fork touching keeper state directly, the same as validator rotation.
func rotateAssetManagers(app *RealioNetwork, ctx sdk.Context) {
	for _, r := range assetManagerRotations {
		key := assetmoduletypes.TokenKey(r.Symbol)
		token, err := app.AssetKeeper.Token.Get(ctx, key)
		if err != nil {
			panic(fmt.Errorf("asset manager rotation: token %q not found: %w", r.Symbol, err))
		}
		token.Manager = r.NewManager
		if err := app.AssetKeeper.Token.Set(ctx, key, token); err != nil {
			panic(fmt.Errorf("asset manager rotation: failed to set token %q: %w", r.Symbol, err))
		}
	}
}
