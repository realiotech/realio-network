package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "claim"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
)

var (
	// AdminKey is the collections.Item prefix for the sole address
	// authorized to submit any of the module's messages.
	AdminKey = collections.NewPrefix(0)

	// AddressLinkKeyPrefix is the collections.Map prefix mapping each old
	// (leaked) address's raw bytes to the raw bytes of the new address it
	// has been linked to.
	AddressLinkKeyPrefix = collections.NewPrefix(1)
)
