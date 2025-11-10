package app

import (
	"fmt"
	"strings"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	realionetworktypes "github.com/realiotech/realio-network/types"
)

// EVMOptionsFn defines a function type for setting app options specifically for
// the Cosmos EVM app. The function should receive the chainID and return an error if
// any.
type EVMOptionsFn func(string) error

// NoOpEVMOptions is a no-op function that can be used when the app does not
// need any specific configuration.
func NoOpEVMOptions(_ string) error {
	return nil
}

var sealed = false

// ChainsCoinInfo is a map of the chain id and its corresponding EvmCoinInfo
// that allows initializing the app with different coin info based on the
// chain id
var ChainsCoinInfo = map[string]evmtypes.EvmCoinInfo{
	// mainnet
	"realionetwork_3301": {
		Denom:         realionetworktypes.BaseDenom,
		ExtendedDenom: realionetworktypes.BaseDenom,
		DisplayDenom:  realionetworktypes.BaseDenom,
		Decimals:      18,
	},
	// testnet
	"realionetwork_3300": {
		Denom:         realionetworktypes.BaseDenom,
		ExtendedDenom: realionetworktypes.BaseDenom,
		DisplayDenom:  realionetworktypes.BaseDenom,
		Decimals:      18,
	},
	// local net
	"realionetworklocal_7777": {
		Denom:         realionetworktypes.BaseDenom,
		ExtendedDenom: realionetworktypes.BaseDenom,
		DisplayDenom:  realionetworktypes.BaseDenom,
		Decimals:      18,
	},
}

const (
	// MainnetChainID defines the RealioNetwork EIP155 chain ID for mainnet
	MainnetChainID = 3301
	// TestnetChainID defines the RealioNetwork EIP155 chain ID for testnet
	TestnetChainID = 3300
)

// EvmAppOptions allows to setup the global configuration
// for the Cosmos EVM chain.
func EvmAppOptions(chainID string) error {
	if sealed {
		return nil
	}

	id := strings.Split(chainID, "-")[0]
	coinInfo, found := ChainsCoinInfo[id]
	if !found {
		return fmt.Errorf("unknown chain id: %s", id)
	}

	err := evmtypes.NewEVMConfigurator().
		WithExtendedEips(cosmosEVMActivators).
		// NOTE: we're using the 18 decimals default for the example chain
		WithEVMCoinInfo(coinInfo).
		Configure()
	if err != nil {
		return err
	}

	sealed = true
	return nil
}
