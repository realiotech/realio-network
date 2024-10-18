package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # genesis/types/default
		Params:             DefaultParams(),
		RegisteredCoins:    []sdk.Coin{},
		RatelimitEpochInfo: DefaultEpochInfo(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	err := gs.Params.Validate()
	if err != nil {
		return err
	}

	err = gs.RegisteredCoins.Validate()
	if err != nil {
		return err
	}

	return gs.RatelimitEpochInfo.Validate()
}
