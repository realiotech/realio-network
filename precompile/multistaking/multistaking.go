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
	codec.Codec
	stakingKeeper      stakingkeeper.Keeper
	multiStakingKeeper multistakingkeeper.Keeper
	erc20Keeper        erc20keeper.Keeper
	addrCodec          address.Codec
}

// NewPrecompile creates a new multistaking Precompile instance implementing the
// PrecompiledContract interface.
func NewPrecompile(
	cdc codec.Codec,
	stakingKeeper stakingkeeper.Keeper,
	multiStakingKeeper multistakingkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	addrCodec address.Codec,
) (*Precompile, error) {
	newABI, err := cmn.LoadABI(f, "abi.json")
	if err != nil {
		return nil, err
	}

	// NOTE: we set an empty gas configuration to avoid extra gas costs
	// during the run execution
	p := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newABI,
			KvGasConfig:          storetypes.GasConfig{},
			TransientKVGasConfig: storetypes.GasConfig{},
		},
		Codec:              cdc,
		stakingKeeper:      stakingKeeper,
		multiStakingKeeper: multiStakingKeeper,
		erc20Keeper:        erc20Keeper,
		addrCodec:          addrCodec,
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
	ctx, stateDB, snapshot, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err, stateDB, snapshot)()

	switch method.Name {
	// Transactions
	case DelegateMethod:
		bz, err = p.DelegateEVM(ctx, evm.Origin, method, args)
	case UndelegateMethod:
		bz, err = p.UndelegateEVM(ctx, evm.Origin, method, args)
	case RedelegateMethod:
		bz, err = p.BeginRedelegateEVM(ctx, evm.Origin, method, args)
	case CancelUnbondingDelegationMethod:
		bz, err = p.CancelUnbondingEVMDelegation(ctx, evm.Origin, method, args)
	case CreateValidatorMethod:
		bz, err = p.CreateEVMValidator(ctx, method, args)

	// Queries: We only support multistaking evm tx for now
	// Use multistaking module query instead

	case DelegationMethod:
		bz, err = p.Delegation(ctx, contract, method, args)
	case UnbondingDelegationMethod:
		bz, err = p.UnbondingDelegation(ctx, contract, method, args)
	case ValidatorMethod: // multistaking
		bz, err = p.Validator(ctx, method, contract, args)
	case ValidatorsMethod: // multistaking
		bz, err = p.Validators(ctx, method, contract, args)
	case RedelegationMethod:
		bz, err = p.Redelegation(ctx, method, contract, args)
	case RedelegationsMethod:
		bz, err = p.Redelegations(ctx, method, contract, args)
	}

	if err != nil {
		return nil, err
	}

	cost := ctx.GasMeter().GasConsumed() - initialGas

	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	if err := p.AddJournalEntries(stateDB, snapshot); err != nil {
		return nil, err
	}

	return bz, nil
}

func (Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case DelegateMethod,
		UndelegateMethod,
		RedelegateMethod,
		CancelUnbondingDelegationMethod,
		CreateValidatorMethod,
		EditValidatorMethod:
		return true
	default:
		return false
	}
}
