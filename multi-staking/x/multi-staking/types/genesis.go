package types

import (
	"fmt"

	"cosmossdk.io/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func DefaultGenesis() *GenesisState {
	stakingGenesis := stakingtypes.DefaultGenesisState()

	return &GenesisState{
		StakingGenesisState: *stakingGenesis,
	}
}

// Validate performs basic genesis state validation, returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// validate staking genesis
	if err := staking.ValidateGenesis(&gs.StakingGenesisState); err != nil {
		return err
	}

	fmt.Println("dcm")

	// validate locks
	if len(gs.MultiStakingLocks) != len(gs.StakingGenesisState.Delegations) {
		return errors.Wrapf(
			ErrInvalidTotalMultiStakingLocks,
			"locks and delegations not equal: locks: %v delegations: %v",
			len(gs.MultiStakingLocks),
			len(gs.StakingGenesisState.Delegations),
		)
	}

	for _, lock := range gs.MultiStakingLocks {
		err := lock.Validate()
		if err != nil {
			return err
		}
	}

	fmt.Println("dcm")

	// validate unlocks
	if len(gs.MultiStakingUnlocks) != len(gs.StakingGenesisState.UnbondingDelegations) {
		return errors.Wrapf(
			ErrInvalidTotalMultiStakingUnlocks,
			"unlocks and unbondingdelegations not equal: unlocks: %v unbondingdelegations: %v",
			len(gs.MultiStakingUnlocks),
			len(gs.StakingGenesisState.UnbondingDelegations),
		)
	}

	for _, unlock := range gs.MultiStakingUnlocks {
		err := unlock.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (g GenesisState) UnpackInterfaces(c codectypes.AnyUnpacker) error {
	for i := range g.StakingGenesisState.Validators {
		if err := g.StakingGenesisState.Validators[i].UnpackInterfaces(c); err != nil {
			return err
		}
	}
	return nil
}
