package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/network/x/asset/types"
)

func (k Keeper) MintTokens(ctx sdk.Context, receiver sdk.AccAddress, amt sdk.Coin) error {
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amt))
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, sdk.NewCoins(amt))
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) BurnTokens(ctx sdk.Context, sender sdk.AccAddress, amt sdk.Coin) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(amt)); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amt)); err != nil {
		return err
	}

	return nil
}