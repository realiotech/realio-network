package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	"github.com/realiotech/realio-network/x/asset/priviledges/transfer_auth"
	"github.com/realiotech/realio-network/x/asset/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AssetSendRestriction(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (newToAddr sdk.AccAddress, err error) {
	newToAddr = toAddr
	err = nil

	tp, has := k.PrivilegeManager["transfer_auth"]
	if !has {
		return newToAddr, nil
	}

	goCtx := sdk.WrapSDKContext(ctx)

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

		queryHandler := tp.QueryHandler()
		resp, err := queryHandler(goCtx, &transfer_auth.QueryIsAllowedRequest{symbol, fromAddr.String()}, symbol)
		if err != nil {
			return newToAddr, err
		}

		queryResp, ok := resp.(*transfer_auth.QueryIsAllowedRespones)
		if !ok {
			err = errors.New("invalid response expecting QueryIsAllowedRespones")
			return newToAddr, err
		}

		isAllow := queryResp.IsAllow

		if isAllow {
			continue
		} else { //nolint:revive // superfluous else, could fix, but not worth it?
			err = errorsmod.Wrapf(types.ErrNotAuthorized, "%s is not authorized to transact with %s", fromAddr, coin.Denom)
			return newToAddr, err
		}
	}
	return newToAddr, nil
}
