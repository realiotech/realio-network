package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/claim/types"
)

// MoveBalances moves amount from the old address to the new one, in two
// hops through the claim module account rather than one direct transfer.
//
// The detour is what makes the transfer possible at all. x/blacklist's
// send restriction blocks a blacklisted address paying a USER account --
// which is exactly this transfer, old (blacklisted) to new (a user
// account). It exempts payments INTO a module account, because routine
// protocol bookkeeping debits blacklisted balances that way and none of
// those callers are written to handle SendCoins failing. So the first hop
// is permitted by that exemption, and the second hop is permitted because
// its sender is the module account, which is not blacklisted. Nothing
// about the restriction is weakened for anyone else: only this module can
// spend from its own module account, and it only ever does so here.
//
// x/asset's restriction clears for the same reason -- it skips any
// transfer where either side is a module account. Note what that implies:
// a permissioned token can land on the new address before the new address
// is on that Token's authorized list. Putting it there is deliberately not
// this module's job, because that list answers to the Token's manager and
// not to this module's admin -- see x/claim/README.md.
func (k Keeper) MoveBalances(ctx context.Context, oldAddr, newAddr sdk.AccAddress, amount sdk.Coins) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, oldAddr, types.ModuleName, amount); err != nil {
		return errorsmod.Wrapf(err, "escrowing %s from %s", amount, oldAddr)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, newAddr, amount); err != nil {
		return errorsmod.Wrapf(err, "paying %s to %s", amount, newAddr)
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeClaimedBalance,
		sdk.NewAttribute(types.AttributeKeyOldAddress, oldAddr.String()),
		sdk.NewAttribute(types.AttributeKeyNewAddress, newAddr.String()),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
	))
	return nil
}
