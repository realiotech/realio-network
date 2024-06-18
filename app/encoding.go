package app

import (
	"cosmossdk.io/simapp/params"
	evmenc "github.com/evmos/evmos/v18/encoding"
)

// MakeEncodingConfig creates the EncodingConfig for realio network
func MakeEncodingConfig() params.EncodingConfig {
	return evmenc.MakeConfig(ModuleBasics)
}
