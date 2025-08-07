package v4

import (
	storetypes "cosmossdk.io/store/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v1.4.0"
)

var V4StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		erc20types.ModuleName,
	},
}
