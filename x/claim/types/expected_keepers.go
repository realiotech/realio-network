package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the subset of x/bank the claim module needs.
//
// SpendableCoins is how MsgClaimBalances decides what a sweep takes: the
// whole balance minus whatever a vesting schedule still locks, since
// locked coins cannot move regardless of who asks. The two Send methods
// move the coins, always via the module account and never as a direct
// account-to-account send (see keeper/bank.go).
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// AccountKeeper defines the subset of x/auth the claim module needs.
// MsgClaimTokens runs its EVM calls with the old address as the caller,
// and reading a caller's nonce requires that address to have an auth
// account -- which an address that only ever received tokens does not.
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}
