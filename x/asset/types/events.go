package types

// staking module event types
const (
	EventTypeTokenCreated = "create_token"
	EventTypeTokenUpdated = "update_token"

	EventTypeAddTokenManager    = "add_manager"
	EventTypeRemoveTokenManager = "remove_manager"

	AttributeKeyTokenId = "token_id"
	AttributeKeySymbol  = "symbol"
	AttributeKeyIndex   = "index"
	AttributeKeyAddress = "address"

	AttributeValueCategory = ModuleName
)
