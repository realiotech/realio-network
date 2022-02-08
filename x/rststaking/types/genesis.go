package types

import (
	"fmt"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		RstStakeList: []RstStake{},
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in rstStake
	rstStakeIndexMap := make(map[string]struct{})

	for _, elem := range gs.RstStakeList {
		index := string(RstStakeKey(elem.Index))
		if _, ok := rstStakeIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for rstStake")
		}
		rstStakeIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return nil
}
