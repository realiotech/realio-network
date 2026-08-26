package types

const (
	// ModuleName defines the module name.
	ModuleName = "blacklist"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
)

// AddressKeyPrefix is prepended to every blacklisted account's raw address
// bytes to form its key in the module's KVStore. The stored value is always
// a single byte (presence = blacklisted); the mapping is a set, not a map to
// any richer value. Keying by the raw AccAddress bytes (not its bech32
// string) keeps keys fixed-length and canonical.
var AddressKeyPrefix = []byte{0x01}

// AddressKey returns the KVStore key for a given account address's raw
// bytes (i.e. sdk.AccAddress, or common.Address.Bytes() for an EVM sender —
// both are the same 20 raw bytes on this chain).
func AddressKey(addr []byte) []byte {
	return append(AddressKeyPrefix, addr...)
}
