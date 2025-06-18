// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package codec

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"

	"github.com/realiotech/realio-network/crypto/ethsecp256k1"
	"github.com/realiotech/realio-network/crypto/ossecp256k1"
)

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&ethsecp256k1.PubKey{},
		ethsecp256k1.PubKeyName, nil)
	cdc.RegisterConcrete(&ethsecp256k1.PrivKey{},
		ethsecp256k1.PrivKeyName, nil)

	cdc.RegisterConcrete(&ossecp256k1.PubKey{}, ossecp256k1.PubKeyName, nil)
	cdc.RegisterConcrete(&ossecp256k1.PrivKey{}, ossecp256k1.PrivKeyName, nil)

	// NOTE: update SDK's amino codec to include the ethsecp256k1 keys.
	// DO NOT REMOVE unless deprecated on the SDK.
	legacy.Cdc = cdc
}
