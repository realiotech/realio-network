package types

func NewToken(name string, symbol string, decimal string, description string) Token {
	return Token{
		Name:        name,
		Symbol:      symbol,
		Decimal:     decimal,
		Description: description,
	}
}
