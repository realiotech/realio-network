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
// This recovers only what the attacker left behind. What was moved out
// before the freeze comes back through PayRestitution, which is separate
// for the reasons documented there.
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

// ClaimFromAttacker pays amount out of the attacker's own account to the new
// address, for coins the attacker took from the old address before the
// freeze.
//
// The attacker comes from module state, not from a caller argument, so
// there is no code path by which this debits anyone else. It refuses to
// pay the attacker itself, which would be a loop that moves nothing and
// means the caller has the wrong address.
//
// Kept as its own function, and its own event, even though one message
// drives both halves of a bank recovery. The two are not the same kind of
// thing and the record should not blur them: a sweep is mechanical, safe
// to repeat, and moves only the holder's own coins; this is a decision
// sized from off-chain analysis, NOT idempotent, and draws on one finite
// pool shared across every victim. Re-sending the message with a
// from_attacker amount pays again.
func (k Keeper) ClaimFromAttacker(ctx context.Context, oldAddr, newAddr sdk.AccAddress, amount sdk.Coins) error {
	if amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "amount must not be empty")
	}

	attacker, ok := k.GetAttacker(ctx)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no attacker address is configured")
	}
	if attacker.Equals(oldAddr) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
			"old_address is the attacker; it is owed no restitution")
	}
	if attacker.Equals(newAddr) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
			"the link points at the attacker; restitution would pay it back to itself")
	}

	if err := k.payVia(ctx, attacker, newAddr, amount); err != nil {
		return errorsmod.Wrapf(err, "paying %s out of %s", amount, attacker)
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeClaimedFromAttacker,
		sdk.NewAttribute(types.AttributeKeyOldAddress, oldAddr.String()),
		sdk.NewAttribute(types.AttributeKeyNewAddress, newAddr.String()),
		sdk.NewAttribute(types.AttributeKeyAttacker, attacker.String()),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
	))
	return nil
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
