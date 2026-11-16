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
	// Admin is the bech32 address authorized to submit any of the module's
	// messages. Empty means unset — every message is rejected until an admin
	// is configured, whether at genesis or by a later chain-upgrade handler
	// (the same way x/blacklist's leaked-address list and x/asset's
	// manager rotations are seeded on a live chain).
	Admin string `json:"admin"`

	// Attacker is the bech32 address MsgMigrateBankBalance pays hacked
	// amounts from -- the account the stolen funds were moved into. Empty
	// means unset, and a payout fails until it is configured. Like Admin, no
	// message changes it.
	Attacker string `json:"attacker"`

	// Links seeds the old->new address mapping. In practice this is
	// populated by MsgLinkAddress at runtime; it's present in genesis
	// mainly so state round-trips through export/import.
	Links []AddressLink `json:"links"`

	// HackedAmounts records, per victim's old address, how much the attacker
	// took and has not yet repaid. MsgMigrateBankBalance pays exactly this
	// out of the attacker and removes the entry. Exported state holds only
	// what is still unpaid.
	HackedAmounts []HackedAmount `json:"hacked_amounts"`
}

// HackedAmount is what the attacker took from one victim address.
type HackedAmount struct {
	// Address is the bech32 old (compromised) address the funds were taken
	// from.
	Address string `json:"address"`

	// Amount is what is owed back, to be paid out of the attacker's account.
	Amount sdk.Coins `json:"amount"`
}

// AddressLink is one old-address -> new-address mapping.
type AddressLink struct {
	OldAddress string `json:"old_address"`
	NewAddress string `json:"new_address"`
}

// DefaultGenesis returns the default (empty, no admin configured) genesis
// state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{Admin: "", Attacker: "", Links: []AddressLink{}, HackedAmounts: []HackedAmount{}}
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

	if gs.Attacker != "" {
		if _, err := sdk.AccAddressFromBech32(gs.Attacker); err != nil {
			return fmt.Errorf("claim genesis: invalid attacker address %q: %w", gs.Attacker, err)
		}
		if gs.Attacker == gs.Admin {
			return fmt.Errorf("claim genesis: attacker and admin must differ (%q)", gs.Attacker)
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

	hackedSeen := make(map[string]struct{}, len(gs.HackedAmounts))
	for _, h := range gs.HackedAmounts {
		addr, err := sdk.AccAddressFromBech32(h.Address)
		if err != nil {
			return fmt.Errorf("claim genesis: invalid hacked_amounts address %q: %w", h.Address, err)
		}
		if h.Address == gs.Attacker {
			return fmt.Errorf("claim genesis: hacked_amounts address %q is the attacker, which is owed nothing", h.Address)
		}
		if h.Amount.IsZero() {
			return fmt.Errorf("claim genesis: hacked_amounts for %q must not be empty", h.Address)
		}
		if err := h.Amount.Validate(); err != nil {
			return fmt.Errorf("claim genesis: invalid hacked_amounts amount for %q: %w", h.Address, err)
		}
		if _, dup := hackedSeen[addr.String()]; dup {
			return fmt.Errorf("claim genesis: duplicate hacked_amounts address %q", h.Address)
		}
		hackedSeen[addr.String()] = struct{}{}
	}
	return nil
}
