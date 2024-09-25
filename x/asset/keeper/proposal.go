package keeper

import (
	"fmt"

	"github.com/realiotech/realio-network/x/asset/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddTokenManager handles the proposals to add a new manager
func (k Keeper) AddTokenManager(
	ctx sdk.Context,
	p *types.AddTokenManager,
) error {
	managerAddress, err := sdk.AccAddressFromBech32(p.ManagerAddress)
	if err != nil {
		return err
	}

	ok := k.IsTokenManager(ctx, managerAddress)
	if ok {
		return fmt.Errorf("manager %s already exist", p.ManagerAddress)
	}

	k.SetTokenManager(ctx, managerAddress)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddTokenManager,
			sdk.NewAttribute(types.AttributeKeyAddress, p.ManagerAddress),
		),
	)
	return nil
}

// RemoveTokenManager handles the proposals to add a new manager
func (k Keeper) RemoveTokenManager(
	ctx sdk.Context,
	p *types.RemoveTokenManager,
) error {
	managerAddress, err := sdk.AccAddressFromBech32(p.ManagerAddress)
	if err != nil {
		return err
	}

	ok := k.IsTokenManager(ctx, managerAddress)
	if !ok {
		return fmt.Errorf("manager %s is not exist", p.ManagerAddress)
	}

	k.DeleteTokenManager(ctx, managerAddress)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddTokenManager,
			sdk.NewAttribute(types.AttributeKeyAddress, p.ManagerAddress),
		),
	)
	return nil
}
