package v8

import (
	storetypes "cosmossdk.io/store/types"

	claimtypes "github.com/realiotech/realio-network/x/claim/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v1.8.0"
)

// V8StoreUpgrades mounts x/claim: this upgrade is the first to ship a binary
// with the module wired into app.go, so nodes running the pre-v1.8.0 binary
// never had its store created at InitChain. Without this, the post-upgrade
// binary's regular DefaultStoreLoader finds "claim" missing/inconsistent
// with the rest of the root multi-store and refuses to load -- the same
// class of failure BlacklistStoreUpgrades and V6StoreUpgrades exist to
// avoid for their own newly-mounted stores (see app/upgrades.go).
var V8StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		claimtypes.ModuleName,
	},
}
