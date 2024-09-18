package keeper

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/asset/types"
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

	for _, coin := range amt {
		// Check if the value already exists
		// fetch bank metadata to get symbol from denom
		tokenID := coin.Denom
		tm, isFound := k.GetTokenManagement(
			ctx,
			tokenID,
		)
		if !isFound {
			continue
		}
		enabledPrivileges := tm.EnabledPrivileges
		for priv, restrictionChecker := range k.RestrictionChecker {
			if slices.Contains(enabledPrivileges, priv) {
				isAllow, err := restrictionChecker.IsAllow(ctx, tokenID, fromAddr.String())
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
		}
	}
	return newToAddr, nil
}
