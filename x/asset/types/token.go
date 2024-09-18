package types

func NewToken(tokenId string, name string, symbol string, decimal uint32, description string) Token {
	return Token{
		TokenId:     tokenId,
		Name:        name,
		Symbol:      symbol,
		Decimal:     decimal,
		Description: description,
	}
}

func NewTokenManagement(manager string, addNewPrivilege bool, excludedPrivileges []string, enabledPrivileges []string) TokenManagement {
	return TokenManagement{
		Manager:            manager,
		AddNewPrivilege:    addNewPrivilege,
		ExcludedPrivileges: excludedPrivileges,
		EnabledPrivileges:  enabledPrivileges,
	}
}
