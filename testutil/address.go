package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/os/crypto/ethsecp256k1"
)

func GenAddress() sdk.AccAddress {
	priv, err := ethsecp256k1.GenerateKey()
	if err != nil {
		panic(err)
	}
	return sdk.AccAddress(priv.PubKey().Address())
}
