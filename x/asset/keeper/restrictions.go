package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (k Keeper) AssetSendRestriction(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (newToAddr sdk.AccAddress, err error) {
	newToAddr = toAddr
	err = nil

	for _, coin := range amt {
		// Check if the value already exists
		// fetch bank metadata to get symbol from denom
		symbol := coin.Denom
		tokenMetadata, found := k.bankKeeper.GetDenomMetaData(ctx, coin.Denom)
		if found {
			symbol = tokenMetadata.Symbol
		}
		token, isFound := k.GetToken(
			ctx,
			symbol,
		)
		if !isFound {
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
			err = sdkerrors.Wrapf(types.ErrNotAuthorized, "%s is not authorized to transact with %s", fromAddr, coin.Denom)
			break
		}
	}
	return newToAddr, err
}
