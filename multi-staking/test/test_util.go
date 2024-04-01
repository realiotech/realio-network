package test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GenPubKey() cryptotypes.PubKey {
	return secp256k1.GenPrivKey().PubKey()
}

func GenAddress() sdk.AccAddress {
	priv := secp256k1.GenPrivKey()

	return sdk.AccAddress(priv.PubKey().Address())
}

func GenValAddress() sdk.ValAddress {
	priv := secp256k1.GenPrivKey()

	return sdk.ValAddress(priv.PubKey().Address())
}

func GenValAddressWithPrivKey() (*secp256k1.PrivKey, sdk.ValAddress) {
	priv := secp256k1.GenPrivKey()

	return priv, sdk.ValAddress(priv.PubKey().Address())
}
