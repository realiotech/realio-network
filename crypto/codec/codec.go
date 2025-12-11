// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package codec

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/evm/x/vm/types/v1"
	"github.com/realiotech/realio-network/crypto/account"
	"github.com/realiotech/realio-network/crypto/ethsecp256k1"
	"github.com/realiotech/realio-network/crypto/legacytx"
	"github.com/realiotech/realio-network/crypto/ossecp256k1"
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
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &ossecp256k1.PubKey{})
	registry.RegisterImplementations((*cryptotypes.PrivKey)(nil), &ossecp256k1.PrivKey{})

	// Support /os.evm.v1.MsgEthereumTx
	registry.RegisterImplementations(
		(*tx.TxExtensionOptionI)(nil),
		&legacytx.ExtensionOptionsEthereumTx{},
		&v1.ExtensionOptionsEthereumTx{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&legacytx.MsgEthereumTx{},
		&v1.MsgEthereumTx{},
		&legacytx.MsgUpdateParams{},
	)
	registry.RegisterInterface(
		"ethermint.evm.v1.TxData",
		(*legacytx.TxData)(nil),
		&legacytx.DynamicFeeTx{},
		&legacytx.AccessListTx{},
		&legacytx.LegacyTx{},
	)
	registry.RegisterInterface(
		"os.evm.v1.TxData",
		(*legacytx.TxData)(nil),
		&legacytx.DynamicFeeTx{},
		&legacytx.AccessListTx{},
		&legacytx.LegacyTx{},
	)

	registry.RegisterInterface(
		"os.vm.v1.TxData",
		(*v1.TxData)(nil),
		&v1.DynamicFeeTx{},
		&v1.AccessListTx{},
		&v1.LegacyTx{},
	)
}
