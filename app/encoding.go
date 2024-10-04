package app

import (
	"cosmossdk.io/simapp/params"

	evmenc "github.com/evmos/os/encoding"
)

// MakeEncodingConfig creates the EncodingConfig for realio network
func MakeEncodingConfig() params.EncodingConfig {
	config := evmenc.MakeConfig()
	return params.EncodingConfig{
		InterfaceRegistry: config.InterfaceRegistry,
		Codec:             config.Codec,
		TxConfig:          config.TxConfig,
		Amino:             config.Amino,
	}
}
