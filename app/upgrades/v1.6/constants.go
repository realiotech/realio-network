package v6

import (
	storetypes "cosmossdk.io/store/types"
	feesponsor "github.com/cosmos/evm/x/feesponsor/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v1.6.0"
)

var V4StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		feesponsor.ModuleName,
	},
}
