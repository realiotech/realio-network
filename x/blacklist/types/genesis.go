package types

import "fmt"

// GenesisState is a plain (non-protobuf) JSON-serializable genesis state:
// it's just a list of addresses, simple enough that hand-writing it was
// sufficient — InitGenesis/ExportGenesis decode/encode this struct directly
// with encoding/json.
type GenesisState struct {
	// Addresses is the list of bech32 account addresses to seed the
	// blacklist with at genesis (or at chain-upgrade time, via an upgrade
	// handler calling the keeper directly instead).
	Addresses []string `json:"addresses"`
}

// DefaultGenesis returns the default (empty) genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{Addresses: []string{}}
}

// Validate performs basic genesis state validation, checking for duplicates.
func (gs GenesisState) Validate() error {
	seen := make(map[string]struct{}, len(gs.Addresses))
	for _, addr := range gs.Addresses {
		if addr == "" {
			return fmt.Errorf("blacklist genesis contains an empty address")
		}
		if _, dup := seen[addr]; dup {
			return fmt.Errorf("blacklist genesis contains duplicate address %q", addr)
		}
		seen[addr] = struct{}{}
	}
	return nil
}
