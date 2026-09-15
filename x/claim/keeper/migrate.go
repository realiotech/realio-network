package keeper

import (
	"context"
	"errors"

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
// Any reward accrued but not yet withdrawn is paid out to the NEW address,
// never the old one: the whole reason this module exists is that the old
// address's key is compromised, so nothing should ever be credited there
// again.
func (k Keeper) MigrateDelegations(ctx context.Context, oldAddr, newAddr sdk.AccAddress) error {
	delegations, err := k.stakingKeeper.GetDelegatorDelegations(ctx, oldAddr, maxDelegatorValidators)
	if err != nil {
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
// (newAddr, valAddr). The hook sequencing below mirrors exactly what
// native x/staking's Delegate/Unbond do internally (see
// cosmos-sdk x/staking/keeper/delegation.go and
// x/distribution/keeper/hooks.go) -- it is load-bearing, not cosmetic:
// BeforeDelegationSharesModified is what flushes rewards and deletes the
// old DelegatorStartingInfo (a no-op on RemoveDelegation's own
// BeforeDelegationRemoved hook), and BeforeDelegationCreated /
// AfterDelegationModified are what create the new DelegatorStartingInfo.
// Getting this order wrong doesn't corrupt state silently -- it surfaces
// later as a wrong reward calculation or a panic in the distribution
// keeper's historical-rewards reference counting.
func (k Keeper) migrateOneDelegation(
	ctx context.Context,
	oldAddr, newAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	oldDel stakingtypes.Delegation,
) error {
	// Redirect any reward this delegation still owes to the NEW address
	// before touching anything else, so the flush below never pays the
	// compromised old address again.
	if err := k.distrKeeper.SetDelegatorWithdrawAddr(ctx, oldAddr, newAddr); err != nil {
		return err
	}

	// Flush pending rewards and drop the old (valAddr, oldAddr)
	// DelegatorStartingInfo -- the same hook native x/staking's Unbond
	// calls before a delegation is fully removed.
	if err := k.distrKeeper.Hooks().BeforeDelegationSharesModified(ctx, oldAddr, valAddr); err != nil {
		return err
	}
	if err := k.stakingKeeper.RemoveDelegation(ctx, oldDel); err != nil {
		return err
	}

	// Re-create the shares under the new address. SetDelegation alone
	// moves no coins and doesn't touch the validator's total tokens/
	// shares -- it only changes who the bookkeeping entry belongs to.
	newDel, err := k.stakingKeeper.GetDelegation(ctx, newAddr, valAddr)
	switch {
	case err == nil:
		if err := k.distrKeeper.Hooks().BeforeDelegationSharesModified(ctx, newAddr, valAddr); err != nil {
			return err
		}
		newDel.Shares = newDel.Shares.Add(oldDel.Shares)
	case errors.Is(err, stakingtypes.ErrNoDelegation):
		if err := k.distrKeeper.Hooks().BeforeDelegationCreated(ctx, newAddr, valAddr); err != nil {
			return err
		}
		newDel = stakingtypes.NewDelegation(newAddr.String(), valAddr.String(), oldDel.Shares)
	default:
		return err
	}
	if err := k.stakingKeeper.SetDelegation(ctx, newDel); err != nil {
		return err
	}
	if err := k.distrKeeper.Hooks().AfterDelegationModified(ctx, newAddr, valAddr); err != nil {
		return err
	}

	return k.migrateMultiStakingLock(ctx, oldAddr, newAddr, valAddr)
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
