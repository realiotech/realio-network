package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is a plain (non-protobuf) JSON-serializable genesis state,
// mirroring x/blacklist's approach: simple enough that hand-writing it was
// sufficient. InitGenesis/ExportGenesis decode/encode this struct directly
// with encoding/json.
type GenesisState struct {
	// Admin is the bech32 address authorized to submit MsgLinkAddress.
	// Empty means unset — MsgLinkAddress always rejects until an admin is
	// configured, whether at genesis or by a later chain-upgrade handler
	// (the same way x/blacklist's leaked-address list and x/asset's
	// manager rotations are seeded on a live chain).
	Admin string `json:"admin"`

	// Links seeds the old->new address mapping. In practice this is
	// populated by MsgLinkAddress at runtime; it's present in genesis
	// mainly so state round-trips through export/import.
	Links []AddressLink `json:"links"`
}

// AddressLink is one old-address -> new-address mapping.
type AddressLink struct {
	OldAddress string `json:"old_address"`
	NewAddress string `json:"new_address"`
}

// DefaultGenesis returns the default (empty, no admin configured) genesis
// state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{Admin: "", Links: []AddressLink{}}
}

// Validate performs basic genesis state validation: well-formed addresses,
// no old_address/new_address collapsing to the same account, and no
// duplicate old_address entries (an old address can only ever be linked
// once).
func (gs GenesisState) Validate() error {
	if gs.Admin != "" {
		if _, err := sdk.AccAddressFromBech32(gs.Admin); err != nil {
			return fmt.Errorf("claim genesis: invalid admin address %q: %w", gs.Admin, err)
		}
	}

	seen := make(map[string]struct{}, len(gs.Links))
	for _, link := range gs.Links {
		oldAddr, err := sdk.AccAddressFromBech32(link.OldAddress)
		if err != nil {
			return fmt.Errorf("claim genesis: invalid old_address %q: %w", link.OldAddress, err)
		}
		if _, err := sdk.AccAddressFromBech32(link.NewAddress); err != nil {
			return fmt.Errorf("claim genesis: invalid new_address %q: %w", link.NewAddress, err)
		}
		if link.OldAddress == link.NewAddress {
			return fmt.Errorf("claim genesis: old_address and new_address must differ (%q)", link.OldAddress)
		}
		if _, dup := seen[oldAddr.String()]; dup {
			return fmt.Errorf("claim genesis: duplicate old_address %q", link.OldAddress)
		}
		seen[oldAddr.String()] = struct{}{}
	}
	return nil
}
