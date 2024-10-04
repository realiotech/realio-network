package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
	AppendSendRestriction(restriction banktypes.SendRestrictionFn)
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	HasSupply(ctx context.Context, denom string) bool
	BlockedAddr(addr sdk.AccAddress) bool
	// Methods imported from bank should be defined here
}

// leaving this here for ibc implemenation
// TransferKeeper defines the expected IBC transfer keeper.
// type TransferKeeper interface {
//	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (transfertypes.DenomTrace, bool)
//	SendTransfer(
//		ctx sdk.Context,
//		sourcePort, sourceChannel string,
//		token sdk.Coin,
//		sender sdk.AccAddress, receiver string,
//		timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
//	) error
//}
//
//// ChannelKeeper defines the expected IBC channel keeper.
// type ChannelKeeper interface {
//	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
//}
