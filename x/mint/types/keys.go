package types

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

var (
	// MinterKey is the key to use for the keeper store.
	MinterKey   = collections.NewPrefix(0)
	ParamsKey   = collections.NewPrefix(1)
	EvmDeadAddr = getEvmDeadAddr()
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

func getEvmDeadAddr() []byte {
	_, EvmDeadAddr, err := bech32.DecodeAndConvert("realio1qqqqqqqqqqqqqqqqqqqqqqqqqqqqph4dujhguh")
	if err != nil {
		panic(err)
	}
	return EvmDeadAddr
}
