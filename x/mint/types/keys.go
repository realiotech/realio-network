package types

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// MinterKey is the key to use for the keeper store.
	MinterKey   = collections.NewPrefix(0)
	ParamsKey   = collections.NewPrefix(1)
	EvmDeadAddr = sdk.MustAccAddressFromBech32("realio1qqqqqqqqqqqqqqqqqqqqqqqqqqqqph4dujhguh")
)

const (
	// module name
	ModuleName = "mint"

	// StoreKey is the default store key for mint
	StoreKey = ModuleName

	// Query endpoints supported by the minting querier
	QueryParameters       = "parameters"
	QueryInflation        = "inflation"
	QueryAnnualProvisions = "annual_provisions"
)
