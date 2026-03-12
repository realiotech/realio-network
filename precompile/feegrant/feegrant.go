// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package feegrant

import (
	"embed"

	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	cmn "github.com/cosmos/evm/precompiles/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/core/vm"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile defines the feegrant precompile
type Precompile struct {
	cmn.Precompile
	abi.ABI
	codec.Codec
	feegrantKeeper feegrantkeeper.Keeper
	addrCodec      address.Codec
}

func LoadABI() (abi.ABI, error) {
	return cmn.LoadABI(f, "abi.json")
}

// NewPrecompile creates a new feegrant Precompile instance implementing the
// PrecompiledContract interface.
func NewPrecompile(
	cdc codec.Codec,
	feegrantKeeper feegrantkeeper.Keeper,
	addrCodec address.Codec,
) (*Precompile, error) {
	newABI, err := LoadABI()
	if err != nil {
		return nil, err
	}

	p := &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
		},
		ABI:            newABI,
		Codec:          cdc,
		feegrantKeeper: feegrantKeeper,
		addrCodec:      addrCodec,
	}

	// SetAddress defines the address of the feegrant precompile contract.
	p.SetAddress(common.HexToAddress(FeeGrantPrecompileAddress))

	return p, nil
}

// RequiredGas calculates the precompiled contract's base gas rate.
func (p Precompile) RequiredGas(input []byte) uint64 {
	methodID := input[:4]

	method, err := p.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method))
}

// Run executes the precompiled contract feegrant methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.Execute(ctx, contract, readOnly)
	})
}

// Execute executes the precompiled contract feegrant methods defined in the ABI.
func (p Precompile) Execute(ctx sdk.Context, contract *vm.Contract, readOnly bool) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.ABI, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	var bz []byte
	switch method.Name {
	// Transactions
	case GrantMethod:
		bz, err = p.GrantEVM(ctx, contract.Caller(), method, args)
	case RevokeMethod:
		bz, err = p.RevokeEVM(ctx, contract.Caller(), method, args)
	}

	if err != nil {
		return nil, err
	}
	return bz, err
}

func (Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case GrantMethod, RevokeMethod:
		return true
	default:
		return false
	}
}
