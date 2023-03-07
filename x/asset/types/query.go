package types

// NewQueryTokenRequest creates a new instance of QueryTokenRequest.
func NewQueryTokenRequest(symbol string) *QueryTokenRequest {
	return &QueryTokenRequest{Symbol: symbol}
}
