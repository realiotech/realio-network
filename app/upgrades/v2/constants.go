package v2

import (
	storetypes "cosmossdk.io/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	bridgetypes "github.com/realiotech/realio-network/x/bridge/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v2"
)

var V2StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		consensustypes.ModuleName,
		crisistypes.ModuleName,
		bridgetypes.ModuleName,
	},
}
