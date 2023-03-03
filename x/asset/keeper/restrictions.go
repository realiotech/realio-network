package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
)

func AssetSendRestriction(k Keeper) banktypes.SendRestrictionFn {

	return func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (newToAddr sdk.AccAddress, err error) {
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
			if token.AuthorizationRequired == true {
				isAuthorizedFrom = k.IsAddressAuthorizedToSend(ctx, coin.Denom, fromAddr)
				isAuthorizedTo = k.IsAddressAuthorizedToSend(ctx, coin.Denom, toAddr)
			}

			if isAuthorizedFrom && isAuthorizedTo {
				continue
			} else {
				err = sdkerrors.Wrapf(types.ErrNotAuthorized, "%s is not authorized to transact with %s", fromAddr, coin.Denom)
				break
			}
		}
		return
	}
}

// SetTransferRestrictionFn Set genereric Authorization Send Transfer Restriction
func (k Keeper) SetTransferRestrictionFn() {
	k.bankKeeper.AppendSendRestriction(AssetSendRestriction(k))
}
