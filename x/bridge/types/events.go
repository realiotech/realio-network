package types

// staking module event types
const (
	EventTypeBridgeIn        = "bridge_in"
	EventTypeBridgeOut       = "bridge_out"
	EventTypeRegisterNewCoin = "register_new_coin"
	EventTypeDeregisterCoin  = "deregister_coin"

	AttributeKeyCoins   = "coins"
	AttributeKeyDenom   = "denom"
	AttributeKeyAddress = "address"

	AttributeValueCategory = ModuleName
)
