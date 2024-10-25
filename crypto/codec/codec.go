// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package codec

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/crypto/ethsecp256k1"
	"github.com/realiotech/realio-network/crypto/account"
)

// RegisterInterfaces register the evmOS key concrete types.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.AccountI)(nil),
		&account.EthAccount{},
	)
	registry.RegisterImplementations(
		(*authtypes.GenesisAccount)(nil),
		&account.EthAccount{},
	)
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &ethsecp256k1.PubKey{})
	registry.RegisterImplementations((*cryptotypes.PrivKey)(nil), &ethsecp256k1.PrivKey{})
}
