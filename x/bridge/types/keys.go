package types

import (
	"cosmossdk.io/collections"
)

var (
	ParamsKey             = collections.NewPrefix(0)
	EpochInfoKey          = collections.NewPrefix(1)
	RegisteredCoinsPrefix = collections.NewPrefix(2)
)

const (
	// ModuleName defines the module name
	ModuleName = "bridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName
)
