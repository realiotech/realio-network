package v8

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
)

// ValidatorRotations lists each validator being rotated out and its
// replacement. Every delegation currently on OldValidator -- including the
// outgoing operator's own self-bond -- gets redelegated to NewValidator.
// NewValidator must already exist on chain (created normally, ahead of this
// upgrade, by its real-world operator) with the same multi-staking coin
// registered as OldValidator: BeginRedelegate requires both sides to accept
// the same coin (see multistaking's msgServer.BeginRedelegate), and each
// outgoing validator's delegators are exclusively locked in one coin each
// (confirmed against the real mainnet export: every delegator on both
// validators being rotated has a lock, one denom per validator).
//
// NewValidator is deliberately left blank below -- the real replacement
// validator addresses aren't known to this codebase yet. Filling them in
// (and wiring RotateValidators into a governance-gated x/upgrade handler)
// is the last step before this can run for real; RotateValidators returns
// an error rather than silently no-op'ing if any entry is left blank, so an
// incomplete config can't accidentally ship.
var ValidatorRotations = []struct {
	OldValidator string
	NewValidator string
}{
	{OldValidator: "realiovaloper18a32el4maw3pqr8xh3yrl9ja4lejs265a5nxtm", NewValidator: ""},
	{OldValidator: "realiovaloper13jrrtkfuuvzdak6zxmr95hek9c228ug50sdsvs", NewValidator: ""},
}

// RotateValidators redelegates every delegation on each OldValidator in
// ValidatorRotations to its NewValidator, including the outgoing operator's
// own self-bond. Real MsgBeginRedelegate is used (not a direct re-key):
// this moves stake belonging to delegators who never individually consented
// to this specific move, so it needs to go through the same mechanics --
// and the same safety checks -- any normal redelegation does, including the
// one that matters most here: BeginRedelegation calls native Unbond
// internally (cosmos-sdk x/staking/keeper/delegation.go), which auto-jails
// a validator whose OWN operator redelegates away enough self-bond to drop
// it below MinSelfDelegation. Since scope here includes the outgoing
// operator's self-bond, that safety net fires exactly when it should: once
// the old validator's self-bond hits zero, it gets jailed automatically,
// with no special-casing needed in this function.
//
// Order per validator: regular delegators first, the operator's own
// self-bond last, purely for readability (native Unbond doesn't refuse to
// unbond from an already-jailed validator, so processing order has no
// effect on correctness here).
//
// One kind of delegation is deliberately not moved: a delegator with a
// redelegation still in progress INTO the outgoing validator is skipped and
// logged (the reasoning is at the skip in redelegateOneDelegation).
//
// Returns an error rather than panicking on any unrecoverable problem: the
// caller is CreateUpgradeHandler, whose signature is exactly
// (module.VersionMap, error), and x/upgrade already treats a non-nil error
// from an upgrade handler as fatal to the block -- the chain halts either
// way, but a returned error gets there through the same path an ordinary
// keeper failure would, instead of a raw panic/stack-trace, and it can be
// asserted on directly in tests instead of through recover().
func RotateValidators(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, multiStakingKeeper multistakingkeeper.Keeper) error {
	msMsgServer := multistakingkeeper.NewMsgServerImpl(multiStakingKeeper)

	// x/staking allows at most MaxEntries redelegation entries per (delegator,
	// source validator, destination validator). A delegator who already moved
	// stake from an outgoing validator to its replacement in small steps
	// during the days before the upgrade can be at that limit, and the
	// rotation needs to add one more entry for them; BeginRedelegate would
	// fail with ErrMaxRedelegationEntries and halt the chain. The rotation
	// adds exactly one entry per delegator per validator, so allow one extra
	// entry for the duration of this function and put the original value back
	// afterwards -- the same approach multistaking's own
	// RemoveMultiStakingCoinProposal uses for its forced undelegations.
	params, err := stakingKeeper.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("validator rotation: failed to read staking params: %w", err)
	}
	originalMaxEntries := params.MaxEntries
	params.MaxEntries = originalMaxEntries + 1
	if err := stakingKeeper.SetParams(ctx, params); err != nil {
		return fmt.Errorf("validator rotation: failed to raise max redelegation entries: %w", err)
	}

	for _, r := range ValidatorRotations {
		if r.NewValidator == "" {
			return fmt.Errorf("validator rotation: no NewValidator configured for %s", r.OldValidator)
		}
		if err := redelegateOneValidator(ctx, stakingKeeper, multiStakingKeeper, msMsgServer, r.OldValidator, r.NewValidator); err != nil {
			return err
		}
	}

	params.MaxEntries = originalMaxEntries
	if err := stakingKeeper.SetParams(ctx, params); err != nil {
		return fmt.Errorf("validator rotation: failed to restore max redelegation entries: %w", err)
	}
	return nil
}

func redelegateOneValidator(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, multiStakingKeeper multistakingkeeper.Keeper, msMsgServer stakingtypes.MsgServer, oldValStr, newValStr string) error {
	oldVal, err := sdk.ValAddressFromBech32(oldValStr)
	if err != nil {
		return fmt.Errorf("validator rotation: invalid old validator %q: %w", oldValStr, err)
	}

	delegations, err := stakingKeeper.GetValidatorDelegations(ctx, oldVal)
	if err != nil {
		return fmt.Errorf("validator rotation: failed to list delegations for %s: %w", oldValStr, err)
	}

	// The outgoing operator's own self-bond is redelegated last -- see the
	// doc comment above for why order doesn't affect correctness, only
	// readability of the resulting event/log order.
	operatorAcc := sdk.AccAddress(oldVal).String()
	var operatorDel *stakingtypes.Delegation
	for i := range delegations {
		if delegations[i].DelegatorAddress == operatorAcc {
			operatorDel = &delegations[i]
			continue
		}
		if err := redelegateOneDelegation(ctx, stakingKeeper, multiStakingKeeper, msMsgServer, delegations[i], oldValStr, newValStr); err != nil {
			return err
		}
	}
	if operatorDel != nil {
		if err := redelegateOneDelegation(ctx, stakingKeeper, multiStakingKeeper, msMsgServer, *operatorDel, oldValStr, newValStr); err != nil {
			return err
		}
	}
	return nil
}

// redelegateOneDelegation moves one delegation off the outgoing validator,
// unless it cannot be moved and is skipped: a delegator with no multi-staking
// lock backing their delegation, a delegator with a redelegation in progress
// into the outgoing validator, or a delegation worth less than one token (see
// the comments at each skip below). A skip is not an error -- it returns nil
// -- only a genuinely unrecoverable problem (bad address data, a keeper read
// that fails, BeginRedelegate itself failing) returns one.
func redelegateOneDelegation(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, multiStakingKeeper multistakingkeeper.Keeper, msMsgServer stakingtypes.MsgServer, del stakingtypes.Delegation, oldValStr, newValStr string) error {
	lockID := multistakingtypes.MultiStakingLockID(del.DelegatorAddress, oldValStr)
	lock, found := multiStakingKeeper.GetMultiStakingLock(ctx, lockID)
	if !found {
		// Skip this delegation, and log it, rather than treat it as corrupt
		// data or halt.
		//
		// A delegation with no matching lock is normal state on this chain,
		// not damage: SetMultiStakingLock deletes a lock once its amount hits
		// zero, but AdjustUnbondAmount's token/share conversion can leave a
		// few units of truncation residue behind in x/staking that a "full"
		// undelegate never fully clears. Measured against the real mainnet
		// export this migration is meant to run against, 374 of 3,764
		// delegations (across 36 validators) are already in this state
		// today, 161 of them above the dust threshold below -- so this is
		// not rare, and not something the dust skip already catches.
		//
		// It also cannot be handled by falling through to the dust check:
		// with no lock, there is no lock.LockedCoin to read a denom or
		// amount from, so there is no coin to pass to BeginRedelegate in the
		// first place, regardless of how small or large the leftover
		// x/staking shares are.
		ctx.Logger().Error("validator rotation: skipping delegation, no multi-staking lock backs it",
			"delegator", del.DelegatorAddress,
			"old_validator", oldValStr,
			"shares", del.Shares.String())
		return nil
	}

	delAddr, err := sdk.AccAddressFromBech32(del.DelegatorAddress)
	if err != nil {
		return fmt.Errorf("validator rotation: invalid delegator address %q: %w", del.DelegatorAddress, err)
	}
	oldVal, err := sdk.ValAddressFromBech32(oldValStr)
	if err != nil {
		return fmt.Errorf("validator rotation: invalid old validator %q: %w", oldValStr, err)
	}

	// This has to be checked up front, not by catching BeginRedelegate's error
	// afterwards: the multistaking msg server moves the lock between the two
	// validators before calling into x/staking, and an upgrade handler has no
	// per-message rollback, so a failed attempt would leave a lock moved with
	// no delegation behind it.
	receiving, err := stakingKeeper.HasReceivingRedelegation(ctx, delAddr, oldVal)
	if err != nil {
		return fmt.Errorf("validator rotation: failed to check redelegations into %s for %s: %w", oldValStr, del.DelegatorAddress, err)
	}
	if receiving {
		// Skip this delegation, and log it, rather than move it or halt.
		//
		// x/staking refuses to redelegate stake out of a validator the
		// delegator is currently receiving a redelegation into
		// (ErrTransitiveRedelegation). The rule is what lets a slash follow
		// stake one hop, so evidence against the validator the stake came from
		// can still reach it. Anyone can put themselves in this state during
		// the unbonding period before the upgrade, deliberately or not, so it
		// has to be handled rather than assumed away. The alternatives are
		// worse:
		//   - erroring out would halt the chain at the upgrade height,
		//     letting one small redelegation block the whole upgrade;
		//   - deleting or rewriting the delegator's redelegation record has no
		//     supported path in x/staking, leaves the redelegation queue and
		//     unbonding-id index pointing at nothing, and drops the trail a
		//     later slash would follow.
		//
		// Skipping costs nothing permanent: the stake stays where it is, and
		// once the earlier redelegation matures (one unbonding period) the
		// delegator can move or unbond it themselves. The log is at error level
		// with the delegator, amount and the time it unblocks, so it can be
		// followed up.
		//
		// It is also safe because of who can end up here. Starting a
		// redelegation takes a signed transaction, and x/blacklist rejects
		// every transaction signed by a blacklisted address
		// (app/ante/blacklist.go). So a blacklisted delegator -- which covers
		// the outgoing operators' own self-bonds and a large share of the stake
		// being rotated -- can never have a redelegation in progress, and is
		// always moved. Only accounts able to sign, that is, not blacklisted,
		// can be skipped, and those are exactly the accounts able to
		// redelegate or unbond by themselves later. (The one gap: an address
		// that governance blacklists after it redelegated, inside that same
		// unbonding window, is skipped and cannot move itself until it is
		// removed from the blacklist.)
		ctx.Logger().Error("validator rotation: skipping delegation, delegator has a redelegation in progress into the outgoing validator",
			"delegator", del.DelegatorAddress,
			"old_validator", oldValStr,
			"amount", lock.LockedCoin.String(),
			"movable_after", latestIncomingRedelegationCompletion(ctx, stakingKeeper, delAddr, oldValStr))
		return nil
	}

	// Skip a delegation that is dust: it still has shares, but they are worth
	// less than one token, so their value truncates to zero.
	//
	// A delegation record with zero shares does not exist (x/staking deletes
	// it), so this is not about empty delegations. It happens when the
	// validator's tokens-per-share rate has dropped below 1 through a slash and
	// a position of only a few of the smallest units is left worth under one
	// unit: 1 share at a rate of 0.5 is 0.5 tokens, which truncates to 0.
	// BeginRedelegate then fails (ErrTinyRedelegationAmount, or "invalid shares
	// amount" once multistaking has adjusted the amount down to zero) and
	// would halt the chain.
	//
	// Skipping this is not a concern. The amount is worth less than 1 unit of
	// the smallest denomination -- 10^-18 of a coin -- so there is no value to
	// move, nothing meaningful is left behind, and x/staking cannot redelegate
	// an amount that truncates to zero for anyone. It only matters that one
	// such delegation, which anybody can create for next to nothing, cannot be
	// used to stop the upgrade. It is logged at info level, unlike the skip
	// above, because nobody needs to follow it up.
	//
	// Like the check above, this must run before BeginRedelegate: by then the
	// multistaking msg server has already moved the lock.
	oldValidator, err := stakingKeeper.GetValidator(ctx, oldVal)
	if err != nil {
		return fmt.Errorf("validator rotation: failed to read outgoing validator %s: %w", oldValStr, err)
	}
	if oldValidator.TokensFromShares(del.Shares).TruncateInt().IsZero() {
		ctx.Logger().Info("validator rotation: skipping dust delegation, its shares are worth less than one token",
			"delegator", del.DelegatorAddress,
			"old_validator", oldValStr,
			"shares", del.Shares.String())
		return nil
	}

	_, err = msMsgServer.BeginRedelegate(ctx, &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    del.DelegatorAddress,
		ValidatorSrcAddress: oldValStr,
		ValidatorDstAddress: newValStr,
		Amount:              sdk.NewCoin(lock.LockedCoin.Denom, lock.LockedCoin.Amount),
	})
	if err != nil {
		return fmt.Errorf("validator rotation: failed to redelegate %s from %s to %s: %w", del.DelegatorAddress, oldValStr, newValStr, err)
	}
	return nil
}

// latestIncomingRedelegationCompletion returns when the last redelegation
// from any validator into oldValStr completes for delAddr, i.e. when the
// stake it blocks becomes movable. Purely informational, for the log; it
// returns the zero time if the redelegations can't be read.
func latestIncomingRedelegationCompletion(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, delAddr sdk.AccAddress, oldValStr string) time.Time {
	reds, err := stakingKeeper.GetRedelegations(ctx, delAddr, 1000)
	if err != nil {
		return time.Time{}
	}
	var latest time.Time
	for _, red := range reds {
		if red.ValidatorDstAddress != oldValStr {
			continue
		}
		for _, e := range red.Entries {
			if e.CompletionTime.After(latest) {
				latest = e.CompletionTime
			}
		}
	}
	return latest
}
