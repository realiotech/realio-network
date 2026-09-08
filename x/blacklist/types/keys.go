package types

const (
	// ModuleName defines the module name.
	ModuleName = "blacklist"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
)

// AddressKeyPrefix is the collections.KeySet prefix for the blacklisted-
// addresses set, keyed by each account's raw AccAddress bytes (not its
// bech32 string, to keep keys fixed-length and canonical — a
// common.Address.Bytes() for an EVM sender is the same 20 raw bytes on this
// chain).
var AddressKeyPrefix = []byte{0x01}
