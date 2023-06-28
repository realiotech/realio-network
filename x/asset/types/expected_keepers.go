package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	AppendSendRestriction(restriction bankkeeper.SendRestrictionFn)
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	HasSupply(ctx sdk.Context, denom string) bool
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
