package types

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(ownerss []TokenOwner) GenesisState {
	return GenesisState{
		TokenOwners: ownerss,
	}
}

// DefaultGenesisState sets default evm genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		TokenOwners: []TokenOwner{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, token := range gs.TokenOwners {
		if !common.IsHexAddress(token.ContractAddress) {
			return fmt.Errorf("not hex addr: %s", token.ContractAddress)
		}
		if !common.IsHexAddress(token.OwnerAddress) {
			return fmt.Errorf("not hex addr: %s", token.OwnerAddress)
		}
	}
	return nil
}
