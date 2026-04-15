package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/types"
)

// AssetSendRestriction enforces token authorization checks on all coin transfers.
// It is registered as a bank send restriction in app.go via BankKeeper.AppendSendRestriction.
//
// The allowAddrs list is populated at keeper construction with all module account addresses
// (via app.ModuleAccountAddrs()). These addresses — e.g. x/distribution, x/gov, x/staking,
// x/bridge — are exempt from authorization checks so that protocol-level transfers (fee
// collection, governance, bridging) are never blocked. The list is set once at startup and
// is never modified at runtime, so no governance path can add arbitrary addresses to it.
func (k Keeper) AssetSendRestriction(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (newToAddr sdk.AccAddress, err error) {
	newToAddr = toAddr

	// Module accounts in allowAddrs skip authorization checks (see doc above).
	if allow := k.AllowAddr(fromAddr) || k.AllowAddr(toAddr); allow {
		return newToAddr, nil
	}

	for _, coin := range amt {
		// Check if the value already exists
		// fetch bank metadata to get symbol from denom
		symbol := coin.Denom
		tokenMetadata, found := k.bankKeeper.GetDenomMetaData(ctx, coin.Denom)
		if found {
			symbol = tokenMetadata.Symbol
		}
		token, err := k.Token.Get(
			ctx,
			types.TokenKey(symbol),
		)
		if err != nil {
			continue
		}

		var isAuthorizedFrom, isAuthorizedTo bool
		if token.AuthorizationRequired {
			isAuthorizedFrom = k.IsAddressAuthorizedToSend(ctx, symbol, fromAddr)
			isAuthorizedTo = k.IsAddressAuthorizedToSend(ctx, symbol, toAddr)
		} else {
			continue
		}

		if isAuthorizedFrom && isAuthorizedTo {
			continue
		} else { //nolint:revive // superfluous else, could fix, but not worth it?
			err := errorsmod.Wrapf(types.ErrNotAuthorized, "%s is not authorized to transact with %s", fromAddr, coin.Denom)
			return nil, err
		}
	}
	return newToAddr, nil
}

// AllowAddr addr checks if a given address is in the list of allowAddrs to skip restrictions
func (k Keeper) AllowAddr(addr sdk.AccAddress) bool {
	return k.allowAddrs[addr.String()]
}
