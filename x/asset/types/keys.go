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
	// TokenKey is the key use for keeper to store token
	TokenKey = []byte{0x00}
	// TokenManagementKey is the key use for keeper to store the management information of token
	TokenManagementKey = []byte{0x01}
	// PrivilegedAccountsKey is the key to store all privileged account
	PrivilegedAccountsKey = []byte{0x02}
	// PrivilegeStoreKey
	PrivilegeStoreKey = []byte{0x03}
)
