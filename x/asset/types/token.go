package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewToken(name string, symbol string, total string, manager string, authorizationRequired bool) Token {
	return Token{
		Name:                  name,
		Symbol:                symbol,
		Total:                 total,
		Manager:               manager,
		AuthorizationRequired: authorizationRequired,
		Authorized:            []*TokenAuthorization{},
	}

}

func NewAuthorization(address sdk.Address) *TokenAuthorization {
	return &TokenAuthorization{Address: address.String(), Authorized: true}
}

func (t *Token) AuthorizeAddress(addr sdk.Address) *Token {
	found := false
	for _, a := range t.Authorized {
		if a.Address == addr.String() {
			a.Authorized = true
			found = true
			break
		}
	}
	if !found {
		newAuthorized := NewAuthorization(addr)
		t.Authorized = append(t.Authorized, newAuthorized)
	}
	return t
}

func (t *Token) UnAuthorizeAddress(addr sdk.Address) *Token {
	for _, a := range t.Authorized {
		if a.Address == addr.String() {
			a.Authorized = false
			break
		}
	}
	return t
}

func (t *Token) AddressIsAuthorized(addr sdk.AccAddress) bool {
	for _, a := range t.Authorized {
		if a.Address == addr.String() && a.Authorized {
			return true
		}
	}
	return false
}
