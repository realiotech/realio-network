package v7

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	storetypes "cosmossdk.io/store/types"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v1.7.0"
)

var V7StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		wasmtypes.ModuleName,
		ibcwasmtypes.ModuleName,
	},
}
