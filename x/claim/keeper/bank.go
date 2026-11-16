package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/realiotech/realio-network/x/claim/types"
)

// MigrateBankBalance moves everything spendable on the old address to the new
// one and reports what moved. No amount is taken: a recovery claims the
// whole balance, because none of it belongs to anyone else. Coins a
// vesting schedule still locks are not spendable and stay put.
//
// This recovers only what the attacker left behind. What it moved out
// before the freeze comes back through ClaimFromAttacker.
func (k Keeper) MigrateBankBalance(ctx context.Context, oldAddr, newAddr sdk.AccAddress) (swept sdk.Coins, err error) {
	swept = k.bankKeeper.SpendableCoins(ctx, oldAddr)
	if swept.IsZero() {
		return swept, nil
	}
	if err := k.payVia(ctx, oldAddr, newAddr, swept); err != nil {
		return nil, errorsmod.Wrapf(err, "sweeping %s from %s", swept, oldAddr)
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeMigratedBankBalance,
		sdk.NewAttribute(types.AttributeKeyOldAddress, oldAddr.String()),
		sdk.NewAttribute(types.AttributeKeyNewAddress, newAddr.String()),
		sdk.NewAttribute(types.AttributeKeySwept, swept.String()),
	))
	return swept, nil
}

// ClaimFromAttacker pays the hacked amount recorded for oldAddr (see
// SetHackedAmount) out of the attacker's account to newAddr, deletes the
// record, and reports what it paid.
//
// Amount and attacker both come from module state, so no caller can size a
// payout or debit anyone else. The attacker as oldAddr or newAddr is
// refused: either would move nothing and means the caller has the wrong
// address.
//
// No record pays nothing and is not an error; this is also what a repeat of
// an already-paid address sees. A record the attacker cannot cover is an
// error and the record stays.
//
// Kept apart from the sweep, with its own event: a sweep moves only the
// holder's own coins, while this draws on one finite pool shared across
// every victim.
func (k Keeper) ClaimFromAttacker(ctx context.Context, oldAddr, newAddr sdk.AccAddress) (paid sdk.Coins, err error) {
	amount, err := k.GetHackedAmount(ctx, oldAddr)
	if err != nil {
		return nil, err
	}
	if amount.IsZero() {
		return sdk.Coins{}, nil
	}

	attacker, ok := k.GetAttacker(ctx)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no attacker address is configured")
	}
	if attacker.Equals(oldAddr) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
			"old_address is the attacker; it is owed no payout")
	}
	if attacker.Equals(newAddr) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
			"the link points at the attacker; the payout would go back to the attacker itself")
	}

	if err := k.payVia(ctx, attacker, newAddr, amount); err != nil {
		return nil, errorsmod.Wrapf(err, "paying %s out of %s", amount, attacker)
	}
	if err := k.ClearHackedAmount(ctx, oldAddr); err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeClaimedFromAttacker,
		sdk.NewAttribute(types.AttributeKeyOldAddress, oldAddr.String()),
		sdk.NewAttribute(types.AttributeKeyNewAddress, newAddr.String()),
		sdk.NewAttribute(types.AttributeKeyAttacker, attacker.String()),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
	))
	return amount, nil
}

// payVia moves amount from a frozen address to a user account by escrowing
// it in the module account first.
//
// The detour is what makes the transfer possible at all. x/blacklist's
// send restriction blocks a blacklisted address paying a USER account --
// which describes both callers here, since the old address and the
// attacker are both frozen and the destination is a plain account. It
// exempts payments INTO a module account, because routine protocol
// bookkeeping debits blacklisted balances that way and none of those
// callers are written to handle SendCoins failing. So the first hop is
// permitted by that exemption, and the second is permitted because its
// sender is the module account, which is not blacklisted. Nothing is
// weakened for anyone else: only this module can spend from its own module
// account, and it only ever does so here, in the same transaction, for
// what it just escrowed.
//
// x/asset's restriction clears for the same reason -- it skips any
// transfer where either side is a module account. Note what that implies:
// a permissioned token can land on the new address before the new address
// is on that Token's authorized list. Putting it there is deliberately not
// this module's job, because that list answers to the Token's manager and
// not to this module's admin -- see x/claim/README.md.
func (k Keeper) payVia(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, amount); err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, to, amount)
}
