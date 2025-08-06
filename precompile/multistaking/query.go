package multistaking

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
)

const (
	// DelegationMethod defines the ABI method name for the staking Delegation
	// query.
	DelegationMethod = "delegation"
	// UnbondingDelegationMethod defines the ABI method name for the staking
	// UnbondingDelegationMethod query.
	UnbondingDelegationMethod = "unbondingDelegation"
	// ValidatorMethod defines the ABI method name for the staking
	// Validator query.
	ValidatorMethod = "validator"
)

// Delegation returns the delegation that a delegator has with a specific validator.
func (p Precompile) Delegation(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewDelegationRequest(args)
	if err != nil {
		return nil, err
	}

	queryServer := multistakingkeeper.NewQueryServerImpl(p.multiStakingKeeper)

	res, err := queryServer.MultiStakingLock(ctx, req)

	if err != nil {
		return nil, err
	}

	// If there is no delegation found, return the response with zero values.
	if !res.Found {
		return method.Outputs.Pack(cmn.Coin{Denom: res.Lock.LockedCoin.Denom, Amount: big.NewInt(0)})
	}

	out := new(DelegationOutput).FromResponse(res)

	return out.Pack(method.Outputs)
}

// UnbondingDelegation returns the delegation currently being unbonded for a delegator from
// a specific validator.
func (p Precompile) UnbondingDelegation(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewUnbondingDelegationRequest(args)
	if err != nil {
		return nil, err
	}

	queryServer := multistakingkeeper.NewQueryServerImpl(p.multiStakingKeeper)

	res, err := queryServer.MultiStakingUnlock(ctx, req)
	if err != nil {
		return nil, err
	}

	if !res.Found {
		return method.Outputs.Pack(UnbondingDelegationResponse{})
	}

	out := new(UnbondingDelegationOutput).FromResponse(res)

	return method.Outputs.Pack(out.UnbondingDelegation)
}

// Validator returns the validator information for a given validator address.
func (p Precompile) Validator(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewValidatorRequest(args)
	if err != nil {
		return nil, err
	}

	queryServer := multistakingkeeper.NewQueryServerImpl(p.multiStakingKeeper)

	res, err := queryServer.Validator(ctx, req)
	if err != nil {
		return method.Outputs.Pack(DefaultValidatorOutput().Validator)
	}

	out := new(ValidatorOutput).FromResponse(res)

	return method.Outputs.Pack(out.Validator)
}
