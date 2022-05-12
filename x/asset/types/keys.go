package types

const (
	// ModuleName defines the module name
	ModuleName = "asset"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_asset"

	// Version defines the current version the IBC module supports
	Version = "asset-1"

	// PortID is the default port id that module binds to
	PortID = "asset"

	// TokenKeyPrefix is the prefix to retrieve all Token
	TokenKeyPrefix = "Token/value/"
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = KeyPrefix("asset-port-")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// TokenKey returns the store key to retrieve a Token from the index fields
func TokenKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}