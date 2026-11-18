package v9

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	storetypes "cosmossdk.io/store/types"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v1.9.0"
)

var V9StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		wasmtypes.ModuleName,
		ibcwasmtypes.ModuleName,
	},
}
