package migrations

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"

	assetmodulekeeper "github.com/realiotech/realio-network/x/asset/keeper"
	blacklistmodulekeeper "github.com/realiotech/realio-network/x/blacklist/keeper"
	bridgemodulekeeper "github.com/realiotech/realio-network/x/bridge/keeper"
)

// Keepers bundles every keeper (and raw-store accessor) the fork/migration
// functions in this package need. Constructed and passed in by app.go rather
// than this package taking *app.RealioNetwork directly: this package is
// imported BY app (app.go calls ScheduleForkUpgrade below), so it can't
// import app back — same reason app/upgrades/vX takes individual
// keepers/*module.Manager instead of the app struct.
type Keepers struct {
	StakingKeeper   *stakingkeeper.Keeper
	AssetKeeper     assetmodulekeeper.Keeper
	BlacklistKeeper blacklistmodulekeeper.Keeper
	BridgeKeeper    bridgemodulekeeper.Keeper
	AuthzKeeper     authzkeeper.Keeper
	Erc20Keeper     erc20keeper.Keeper

	Codec           codec.Codec
	StakingStoreKey *storetypes.KVStoreKey
}
