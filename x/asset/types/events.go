package types

// staking module event types
const (
	EventTypeTokenCreated   = "create_token"
	EventTypeTokenUpdated   = "update_token"
	EventTypeTokenAuthorized   = "authorize_token"
	EventTypeTokenUnAuthorized   = "unauthorize_token"


	AttributeKeySymbol         = "symbol"
	AttributeKeyIndex          = "index"
	AttributeKeyAddress          = "index"

	AttributeValueCategory        = ModuleName
)
