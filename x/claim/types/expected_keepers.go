package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the subset of x/bank the claim module needs: the two
// Send methods MsgClaimBalances moves coins with -- always via the module
// account, never a direct account-to-account send (see keeper/bank.go).
type BankKeeper interface {
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
