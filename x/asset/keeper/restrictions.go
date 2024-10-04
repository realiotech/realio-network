package keeper

import (
	"context"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (k Keeper) AssetSendRestriction(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (newToAddr sdk.AccAddress, err error) {
	newToAddr = toAddr
	err = nil

	// module whitelisted addresses can send coins without restrictions
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
			strings.ToLower(symbol),
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
			err = errorsmod.Wrapf(types.ErrNotAuthorized, "%s is not authorized to transact with %s", fromAddr, coin.Denom)
			break
		}
	}
	return newToAddr, err
}

// AllowAddr addr checks if a given address is in the list of allowAddrs to skip restrictions
func (k Keeper) AllowAddr(addr sdk.AccAddress) bool {
	return k.allowAddrs[addr.String()]
}
