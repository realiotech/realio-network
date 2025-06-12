package ossecp256k1

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *codec.LegacyAmino) {
	// cdc.RegisterConcrete(&PubKey{},
	// 	PubKeyName, nil)
	// cdc.RegisterConcrete(&PrivKey{},
	// 	PrivKeyName, nil)

	// // NOTE: update SDK's amino codec to include the ethsecp256k1 keys.
	// // DO NOT REMOVE unless deprecated on the SDK.
	// legacy.Cdc = cdc
}
