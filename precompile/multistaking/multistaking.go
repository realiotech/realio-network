// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package multistaking

import (
	"embed"

	cmn "github.com/cosmos/evm/precompiles/common"
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/ethereum/go-ethereum/core/vm"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile defines the multistaking precompile
type Precompile struct {
	cmn.Precompile
	abi.ABI
	codec.Codec
	stakingKeeper      stakingkeeper.Keeper
	multiStakingKeeper multistakingkeeper.Keeper
	erc20Keeper        erc20keeper.Keeper
	addrCodec          address.Codec
	valAddrCodec       address.Codec
}

func LoadABI() (abi.ABI, error) {
	return cmn.LoadABI(f, "abi.json")
}

// NewPrecompile creates a new multistaking Precompile instance implementing the
// PrecompiledContract interface.
func NewPrecompile(
	cdc codec.Codec,
	stakingKeeper stakingkeeper.Keeper,
	multiStakingKeeper multistakingkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	addrCodec address.Codec,
	valAddrCodec address.Codec,
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
		ABI:                newABI,
		Codec:              cdc,
		stakingKeeper:      stakingKeeper,
		multiStakingKeeper: multiStakingKeeper,
		erc20Keeper:        erc20Keeper,
		addrCodec:          addrCodec,
		valAddrCodec:       valAddrCodec,
	}

	// SetAddress defines the address of the multistaking compile contract.
	p.SetAddress(common.HexToAddress(MultistakingPrecompileAddress))

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

// Run executes the precompiled contract multistaking methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.Execute(ctx, contract, readOnly)
	})
}

// Execute executes the precompiled contract bank query methods defined in the ABI.
func (p Precompile) Execute(ctx sdk.Context, contract *vm.Contract, readOnly bool) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.ABI, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	var bz []byte
	switch method.Name {
	// Transactions
	case DelegateMethod:
		bz, err = p.DelegateEVM(ctx, contract.Caller(), method, args)
	case UndelegateMethod:
		bz, err = p.UndelegateEVM(ctx, contract.Caller(), method, args)
	case RedelegateMethod:
		bz, err = p.BeginRedelegateEVM(ctx, contract.Caller(), method, args)
	case CancelUnbondingDelegationMethod:
		bz, err = p.CancelUnbondingEVMDelegation(ctx, contract.Caller(), method, args)
	case CreateValidatorMethod:
		bz, err = p.CreateEVMValidator(ctx, contract.Caller(), method, args)
	case DelegationMethod:
		bz, err = p.Delegation(ctx, contract, method, args)
	case UnbondingDelegationMethod:
		bz, err = p.UnbondingDelegation(ctx, contract, method, args)
	case ValidatorMethod:
		bz, err = p.Validator(ctx, contract, method, args)

	}

	if err != nil {
		return nil, err
	}
	return bz, err
}

func (Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case DelegateMethod,
		UndelegateMethod,
		RedelegateMethod,
		CancelUnbondingDelegationMethod,
		CreateValidatorMethod:
		return true
	default:
		return false
	}
}
