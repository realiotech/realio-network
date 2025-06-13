// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package ossecp256k1

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/codec"
	// cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	// "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
)

// RegisterInterfaces register the evmOS key concrete types.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &PubKey{})
	registry.RegisterImplementations((*cryptotypes.PrivKey)(nil), &PrivKey{})
}

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&PubKey{},
		PubKeyName, nil)
	cdc.RegisterConcrete(PrivKey{},
		PrivKeyName, nil)

	// keyring.RegisterLegacyAminoCodec(cdc)
	// cryptocodec.RegisterCrypto(cdc)

	// NOTE: update SDK's amino codec to include the ethsecp256k1 keys.
	// DO NOT REMOVE unless deprecated on the SDK.
	legacy.Cdc = cdc
}