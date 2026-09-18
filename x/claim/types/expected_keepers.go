package types

import (
	"context"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// BankKeeper defines the subset of x/bank the claim module needs: resolving
// a denom to its bank metadata Symbol, the same lookup x/asset's own
// AssetSendRestriction uses to find which permissioned Token (if any) a
// denom corresponds to.
type BankKeeper interface {
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
}
