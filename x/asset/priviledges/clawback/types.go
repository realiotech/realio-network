package clawback

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	SendCoins(ctx sdk.Context, senderAddr, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SpendableCoin(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
