package types

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # genesis/types/default
		Params:             DefaultParams(),
		RegisteredCoins:    []CoinAuthority{},
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

	for _, coinAuth := range gs.RegisteredCoins {
		err := coinAuth.Coin.Validate()
		if err != nil {
			return err
		}
	}

	return gs.RatelimitEpochInfo.Validate()
}
