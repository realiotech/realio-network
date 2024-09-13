package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/realiotech/realio-network/x/asset/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RestrictionChecker interface {
	IsAllow(ctx sdk.Context, tokenId string, sender string) (bool, error)
}

func (k Keeper) AssetSendRestriction(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (newToAddr sdk.AccAddress, err error) {
	newToAddr = toAddr
	err = nil

	// if no checker exist allow all sender
	if len(k.RestrictionChecker) == 0 {
		return newToAddr, nil
	}

	checker := k.RestrictionChecker[0]

	for _, coin := range amt {
		// Check if the value already exists
		// fetch bank metadata to get symbol from denom
		symbol := coin.Denom
		tokenMetadata, found := k.bankKeeper.GetDenomMetaData(ctx, coin.Denom)
		if found {
			symbol = tokenMetadata.Symbol
		}
		_, isFound := k.GetToken(
			ctx,
			symbol,
		)
		if !isFound {
			continue
		}

		isAllow, err := checker.IsAllow(ctx, symbol, fromAddr.String())
		if err != nil {
			return newToAddr, err
		}

		if isAllow {
			continue
		} else { //nolint:revive // superfluous else, could fix, but not worth it?
			err = errorsmod.Wrapf(types.ErrNotAuthorized, "%s is not authorized to transact with %s", fromAddr, coin.Denom)
			return newToAddr, err
		}
	}
	return newToAddr, nil
}
