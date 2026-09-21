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
// is the last step before this can run for real; RotateValidators panics
// rather than silently no-op'ing if any entry is left blank, so an
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
func RotateValidators(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, multiStakingKeeper multistakingkeeper.Keeper) {
	msMsgServer := multistakingkeeper.NewMsgServerImpl(multiStakingKeeper)

	for _, r := range ValidatorRotations {
		if r.NewValidator == "" {
			panic(fmt.Errorf("validator rotation: no NewValidator configured for %s", r.OldValidator))
		}
		redelegateOneValidator(ctx, stakingKeeper, multiStakingKeeper, msMsgServer, r.OldValidator, r.NewValidator)
	}
}

func redelegateOneValidator(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, multiStakingKeeper multistakingkeeper.Keeper, msMsgServer stakingtypes.MsgServer, oldValStr, newValStr string) {
	oldVal, err := sdk.ValAddressFromBech32(oldValStr)
	if err != nil {
		panic(fmt.Errorf("validator rotation: invalid old validator %q: %w", oldValStr, err))
	}

	delegations, err := stakingKeeper.GetValidatorDelegations(ctx, oldVal)
	if err != nil {
		panic(fmt.Errorf("validator rotation: failed to list delegations for %s: %w", oldValStr, err))
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
		redelegateOneDelegation(ctx, stakingKeeper, multiStakingKeeper, msMsgServer, delegations[i], oldValStr, newValStr)
	}
	if operatorDel != nil {
		redelegateOneDelegation(ctx, stakingKeeper, multiStakingKeeper, msMsgServer, *operatorDel, oldValStr, newValStr)
	}
}

// redelegateOneDelegation moves one delegation off the outgoing validator,
// unless the delegator has a redelegation in progress into it (see below).
func redelegateOneDelegation(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, multiStakingKeeper multistakingkeeper.Keeper, msMsgServer stakingtypes.MsgServer, del stakingtypes.Delegation, oldValStr, newValStr string) {
	lockID := multistakingtypes.MultiStakingLockID(del.DelegatorAddress, oldValStr)
	lock, found := multiStakingKeeper.GetMultiStakingLock(ctx, lockID)
	if !found {
		panic(fmt.Errorf("validator rotation: no multi-staking lock for delegator %s on %s", del.DelegatorAddress, oldValStr))
	}

	delAddr, err := sdk.AccAddressFromBech32(del.DelegatorAddress)
	if err != nil {
		panic(fmt.Errorf("validator rotation: invalid delegator address %q: %w", del.DelegatorAddress, err))
	}
	oldVal, err := sdk.ValAddressFromBech32(oldValStr)
	if err != nil {
		panic(fmt.Errorf("validator rotation: invalid old validator %q: %w", oldValStr, err))
	}

	// This has to be checked up front, not by catching BeginRedelegate's error
	// afterwards: the multistaking msg server moves the lock between the two
	// validators before calling into x/staking, and an upgrade handler has no
	// per-message rollback, so a failed attempt would leave a lock moved with
	// no delegation behind it.
	receiving, err := stakingKeeper.HasReceivingRedelegation(ctx, delAddr, oldVal)
	if err != nil {
		panic(fmt.Errorf("validator rotation: failed to check redelegations into %s for %s: %w", oldValStr, del.DelegatorAddress, err))
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
		//   - panicking would halt the chain at the upgrade height, letting
		//     one small redelegation block the whole upgrade;
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
		return
	}

	_, err = msMsgServer.BeginRedelegate(ctx, &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    del.DelegatorAddress,
		ValidatorSrcAddress: oldValStr,
		ValidatorDstAddress: newValStr,
		Amount:              sdk.NewCoin(lock.LockedCoin.Denom, lock.LockedCoin.Amount),
	})
	if err != nil {
		panic(fmt.Errorf("validator rotation: failed to redelegate %s from %s to %s: %w", del.DelegatorAddress, oldValStr, newValStr, err))
	}
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
