package app

import (
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

var ChainsCoinInfo = evmtypes.EvmCoinInfo{
	Denom:         realionetworktypes.BaseDenom,
	ExtendedDenom: realionetworktypes.BaseDenom,
	DisplayDenom:  realionetworktypes.BaseDenom,
	Decimals:      18,
}

// EvmAppOptions allows to setup the global configuration
// for the Cosmos EVM chain.
func EvmAppOptions(chainID string) error {
	if sealed {
		return nil
	}

	ethCfg := evmtypes.DefaultChainConfig(chainID)

	err := evmtypes.NewEVMConfigurator().
		WithExtendedEips(cosmosEVMActivators).
		WithChainConfig(ethCfg).
		// NOTE: we're using the 18 decimals default for the example chain
		WithEVMCoinInfo(ChainsCoinInfo).
		Configure()
	if err != nil {
		return err
	}

	sealed = true
	return nil
}
