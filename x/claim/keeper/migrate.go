package keeper

import (
	"context"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
)

// maxDelegatorValidators bounds how many of the old address's delegations
// MigrateDelegations walks in one call. No realistic account delegates to
// anywhere near this many validators; the bound exists only because
// GetDelegatorDelegations requires a retrieve limit.
const maxDelegatorValidators = 65535

// MigrateDelegations re-keys every ACTIVE (bonded) delegation the old
// address holds over to the new address, without unbonding/rebonding:
// shares move directly in the staking and multi-staking stores, so voting
// power and the validator's total stake are completely unaffected.
//
// Deliberately out of scope: any unbonding delegation or in-flight
// redelegation the old address already has is left untouched. It matures
// and pays out to the old address exactly as it would have otherwise --
// safe, because x/blacklist only blocks OUTGOING transfers from a
// blacklisted address, never a module account paying funds in. See
// x/claim/README.md.
//
// Any reward this migration DOES pay out immediately (the merge case
// below) always goes to the NEW address, never the old one: the whole
// reason this module exists is that the old address's key is compromised,
// so nothing should ever be credited there again. SetDelegatorWithdrawAddr
// is set unconditionally up front (once, not per validator -- it's a
// per-account setting) so that guarantee holds no matter which validators
// end up needing a flush.
func (k Keeper) MigrateDelegations(ctx context.Context, oldAddr, newAddr sdk.AccAddress) error {
	delegations, err := k.stakingKeeper.GetDelegatorDelegations(ctx, oldAddr, maxDelegatorValidators)
	if err != nil {
		return err
	}
	if len(delegations) == 0 {
		return nil
	}

	if err := k.distrKeeper.SetDelegatorWithdrawAddr(ctx, oldAddr, newAddr); err != nil {
		return err
	}

	for _, del := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(del.GetValidatorAddr())
		if err != nil {
			return err
		}
		if err := k.migrateOneDelegation(ctx, oldAddr, newAddr, valAddr, del); err != nil {
			return errorsmod.Wrapf(err, "migrating delegation to validator %s", valAddr)
		}
	}
	return nil
}

// migrateOneDelegation moves a single (oldAddr, valAddr) delegation to
// (newAddr, valAddr), then its backing multi-staking lock. Which of the
// two strategies below applies depends on whether newAddr already
// delegates to valAddr -- see migrateDelegationMerge and
// migrateDelegationFresh for why they have to differ.
func (k Keeper) migrateOneDelegation(
	ctx context.Context,
	oldAddr, newAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	oldDel stakingtypes.Delegation,
) error {
	newDel, err := k.stakingKeeper.GetDelegation(ctx, newAddr, valAddr)
	switch {
	case err == nil:
		if err := k.migrateDelegationMerge(ctx, oldAddr, newAddr, valAddr, oldDel, newDel); err != nil {
			return err
		}
	case errors.Is(err, stakingtypes.ErrNoDelegation):
		if err := k.migrateDelegationFresh(ctx, oldAddr, newAddr, valAddr, oldDel); err != nil {
			return err
		}
	default:
		return err
	}

	return k.migrateMultiStakingLock(ctx, oldAddr, newAddr, valAddr)
}

// migrateDelegationMerge handles the case where newAddr already delegates
// to valAddr: oldAddr's and newAddr's shares have to combine into the one
// existing delegation record. The hook sequencing mirrors exactly what
// native x/staking's Unbond (for removing oldAddr's side) and Delegate's
// existing-delegation branch (for growing newAddr's side) do internally --
// see cosmos-sdk x/staking/keeper/delegation.go and
// x/distribution/keeper/hooks.go. It is load-bearing, not cosmetic:
// BeforeDelegationSharesModified is what flushes pending rewards (paid to
// newAddr, per MigrateDelegations' withdraw-address redirect) and drops
// the old DelegatorStartingInfo; AfterDelegationModified is what
// re-initializes it for the combined position. Getting the order wrong
// doesn't corrupt state silently -- it surfaces later as a wrong reward
// calculation or a panic in distribution's historical-rewards reference
// counting.
//
// A flush is unavoidable here: oldAddr's and newAddr's delegations are two
// independent reward-accrual ledgers (each with its own
// DelegatorStartingInfo), and there's no lossless way to combine two such
// ledgers into the single scalar a merged delegation needs. Paying out
// both up front and starting the merged position fresh is what native
// x/staking itself does any time delegated shares change, so this isn't a
// weaker guarantee than an ordinary delegation gets.
func (k Keeper) migrateDelegationMerge(
	ctx context.Context,
	oldAddr, newAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	oldDel, newDel stakingtypes.Delegation,
) error {
	if err := k.distrKeeper.Hooks().BeforeDelegationSharesModified(ctx, oldAddr, valAddr); err != nil {
		return err
	}
	if err := k.stakingKeeper.RemoveDelegation(ctx, oldDel); err != nil {
		return err
	}

	if err := k.distrKeeper.Hooks().BeforeDelegationSharesModified(ctx, newAddr, valAddr); err != nil {
		return err
	}
	newDel.Shares = newDel.Shares.Add(oldDel.Shares)
	if err := k.stakingKeeper.SetDelegation(ctx, newDel); err != nil {
		return err
	}
	return k.distrKeeper.Hooks().AfterDelegationModified(ctx, newAddr, valAddr)
}

// migrateDelegationFresh handles the common case where newAddr has no
// existing delegation to valAddr yet. Rather than flushing oldAddr's
// pending reward into an immediate payout, it carries the exact
// reward-accrual ledger (DelegatorStartingInfo: which historical period it
// last settled against, and its token-value stake as of that settlement)
// over to newAddr verbatim. newAddr then withdraws whenever it likes, via
// the normal MsgWithdrawDelegatorReward, for exactly what accrued since
// oldAddr's last claim -- nothing paid out early as a side effect of the
// admin's LinkAddress call, nothing lost either.
//
// This deliberately bypasses distribution's BeforeDelegationCreated /
// BeforeDelegationSharesModified / AfterDelegationModified hooks: every
// one of them would either flush (pay out now, the opposite of what this
// path is for) or reset the new delegation to a fresh zero-reward
// baseline (silently forfeiting oldAddr's accrued reward instead of
// preserving it). The historical-rewards reference count backing
// DelegatorStartingInfo.PreviousPeriod is left untouched on purpose, not
// forgotten: moving the record from oldAddr's key to newAddr's key doesn't
// change how many DelegatorStartingInfo entries point at that period --
// still exactly one, just under a different owner -- so there is nothing
// to increment or decrement. (The reference-count functions are
// unexported in cosmos-sdk precisely because callers are never meant to
// touch them directly; not needing to here is what makes this safe.)
func (k Keeper) migrateDelegationFresh(
	ctx context.Context,
	oldAddr, newAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	oldDel stakingtypes.Delegation,
) error {
	has, err := k.distrKeeper.HasDelegatorStartingInfo(ctx, valAddr, oldAddr)
	if err != nil {
		return err
	}
	if !has {
		// Every active delegation gets a DelegatorStartingInfo the moment
		// it's created (native Delegate always runs the hooks). Reaching
		// here would mean that invariant broke elsewhere -- fail loudly
		// rather than silently hand newAddr a zero-value ledger.
		return fmt.Errorf("no DelegatorStartingInfo for existing delegation (validator %s, delegator %s)", valAddr, oldAddr)
	}
	startingInfo, err := k.distrKeeper.GetDelegatorStartingInfo(ctx, valAddr, oldAddr)
	if err != nil {
		return err
	}

	if err := k.stakingKeeper.RemoveDelegation(ctx, oldDel); err != nil {
		return err
	}

	newDel := stakingtypes.NewDelegation(newAddr.String(), valAddr.String(), oldDel.Shares)
	if err := k.stakingKeeper.SetDelegation(ctx, newDel); err != nil {
		return err
	}

	if err := k.distrKeeper.SetDelegatorStartingInfo(ctx, valAddr, newAddr, startingInfo); err != nil {
		return err
	}
	return k.distrKeeper.DeleteDelegatorStartingInfo(ctx, valAddr, oldAddr)
}

// migrateMultiStakingLock moves the MultiStakingLock backing this
// delegation (the escrowed original-denom coin a bond-coin delegation
// represents) from the old address's key to the new one's. This must
// happen in lockstep with the staking-side re-key above: multi-staking's
// unbonding payout later looks up this same LockID/UnlockID purely by the
// address recorded in the native staking delegation at maturity time, with
// no independent cross-reference -- leaving one side re-keyed and not the
// other either orphans the lock (silently, for a still-bonded delegation)
// or, for an unbonding one, panics EndBlocker with "unlock entry not
// found" when it matures. Skipped entirely if the old address never
// locked a multi-staking coin for this validator (a delegation funded
// directly in the native bond denom has no lock to move).
func (k Keeper) migrateMultiStakingLock(ctx context.Context, oldAddr, newAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	oldLockID := multistakingtypes.MultiStakingLockID(oldAddr.String(), valAddr.String())
	lock, found := k.multiStakingKeeper.GetMultiStakingLock(ctx, oldLockID)
	if !found {
		return nil
	}
	k.multiStakingKeeper.RemoveMultiStakingLock(ctx, oldLockID)

	newLockID := multistakingtypes.MultiStakingLockID(newAddr.String(), valAddr.String())
	newLock := k.multiStakingKeeper.GetOrCreateMultiStakingLock(ctx, newLockID)
	if err := newLock.AddCoinToMultiStakingLock(lock.LockedCoin); err != nil {
		return err
	}
	k.multiStakingKeeper.SetMultiStakingLock(ctx, newLock)
	return nil
}
