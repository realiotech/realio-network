// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package multistaking

import (
	"embed"

	storetypes "cosmossdk.io/store/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/cosmos/cosmos-sdk/codec"
	cmn "github.com/cosmos/evm/precompiles/common"
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile defines the precompiled contract for multistaking.
type Precompile struct {
	cmn.Precompile
	codec.Codec
	multistakingKeeper multistakingkeeper.Keeper
	bankKeeper         bankkeeper.Keeper
	stakingKeeper      stakingkeeper.Keeper
	erc20Keeper        erc20keeper.Keeper
	authzKeeper        authzkeeper.Keeper
}

// NewPrecompile creates a new multistaking Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	cdc codec.Codec,
	multistakingKeeper multistakingkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) (*Precompile, error) {
	newABI, err := cmn.LoadABI(f, "abi.json")
	if err != nil {
		return nil, err
	}

	p := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newABI,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
		},
		Codec:              cdc,
		multistakingKeeper: multistakingKeeper,
		bankKeeper:         bankKeeper,
		erc20Keeper:        erc20Keeper,
	}
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
		bz, err = p.Delegate(ctx, evm.Origin, stateDB, method, args)
	case UndelegateMethod:
		bz, err = p.Undelegate(ctx, evm.Origin, stateDB, method, args)
	case RedelegateMethod:
		bz, err = p.Redelegate(ctx, evm.Origin, stateDB, method, args)
	case CancelUnbondingDelegationMethod:
		bz, err = p.CancelUnbondingDelegation(ctx, evm.Origin, stateDB, method, args)
	case CreateValidatorMethod:
		bz, err = p.CreateValidator(ctx, evm.Origin, stateDB, method, args)
	case EditValidatorMethod:
		bz, err = p.EditValidator(ctx, evm.Origin, stateDB, method, args)
		// Queries
		// case DelegationMethod:
		// 	bz, err = p.Delegation(ctx, method, args)
		// case UnbondingDelegationMethod:
		// 	bz, err = p.UnbondingDelegation(ctx, method, args)
		// case ValidatorMethod:
		// 	bz, err = p.Validator(ctx, method, args)
		// case ValidatorsMethod:
		// 	bz, err = p.Validators(ctx, method, args)
		// case DelegatorDelegationsMethod:
		// 	bz, err = p.DelegatorDelegations(ctx, method, args)
		// case DelegatorUnbondingDelegationsMethod:
		// 	bz, err = p.DelegatorUnbondingDelegations(ctx, method, args)
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

// IsTransaction checks if the given method name corresponds to a transaction or query.
//
// Available multistaking transactions are:
//   - Delegate
//   - Undelegate
//   - Redelegate
//   - CancelUnbondingDelegation
//   - CreateValidator
//   - EditValidator
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
