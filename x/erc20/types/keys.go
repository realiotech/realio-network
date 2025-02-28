package types

import (
	"cosmossdk.io/collections"
)

var (
	ModuleName       = "extend-erc20"
	ContractOwnerKey = collections.NewPrefix(11)
	StoreKey         = ModuleName
)
