package migrations

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
)

// LeakedValidatorRedelegations lists each leaked validator's replacement.
// Every delegation currently on OldValidator -- including the leaked
// operator's own self-bond -- gets redelegated to NewValidator. NewValidator
// must already exist on chain (created normally, ahead of this fork/
// upgrade, by its real-world operator) with the same multi-staking coin
// registered as OldValidator: BeginRedelegate requires both sides to accept
// the same coin (see multistaking's msgServer.BeginRedelegate), and a
// leaked validator's delegators are exclusively locked in one coin each
// (confirmed against the real mainnet export: every delegator on both
// leaked validators has a lock, one denom per validator).
//
// NewValidator is deliberately left blank below -- the real replacement
// validator addresses aren't known to this codebase yet. Filling them in
// (and wiring redelegateLeakedValidators into ScheduleForkUpgrade at a
// chosen height, or into a governance-gated x/upgrade handler) is the
// last step before this can run for real; RedelegateLeakedValidators
// panics rather than silently no-op'ing if any entry is left blank, so an
// incomplete config can't accidentally ship.
var LeakedValidatorRedelegations = []struct {
	OldValidator string
	NewValidator string
}{
	{OldValidator: "realiovaloper18a32el4maw3pqr8xh3yrl9ja4lejs265a5nxtm", NewValidator: ""},
	{OldValidator: "realiovaloper13jrrtkfuuvzdak6zxmr95hek9c228ug50sdsvs", NewValidator: ""},
}

// RedelegateLeakedValidators redelegates every delegation on each
// OldValidator in LeakedValidatorRedelegations to its NewValidator,
// including the leaked operator's own self-bond. Real MsgBeginRedelegate
// is used (not a direct re-key): unlike x/claim's address-migration path,
// this moves stake belonging to delegators who were never compromised and
// never consented to this specific move, so it needs to go through the
// same mechanics -- and the same safety checks -- any normal redelegation
// does, including the one that matters most here: BeginRedelegation calls
// native Unbond internally (cosmos-sdk x/staking/keeper/delegation.go),
// which auto-jails a validator whose OWN operator redelegates away enough
// self-bond to drop it below MinSelfDelegation. Since scope here includes
// the leaked operator's self-bond, that safety net fires exactly when it
// should: once the old validator's self-bond hits zero, it gets jailed
// automatically, with no special-casing needed in this function.
//
// Order per validator: regular delegators first, the operator's own
// self-bond last, purely for readability (native Unbond doesn't refuse to
// unbond from an already-jailed validator, so processing order has no
// effect on correctness here).
func RedelegateLeakedValidators(k Keepers, ctx sdk.Context) {
	msMsgServer := multistakingkeeper.NewMsgServerImpl(k.MultiStakingKeeper)

	for _, r := range LeakedValidatorRedelegations {
		if r.NewValidator == "" {
			panic(fmt.Errorf("leaked validator redelegation: no NewValidator configured for %s", r.OldValidator))
		}
		redelegateOneValidator(k, ctx, msMsgServer, r.OldValidator, r.NewValidator)
	}
}

func redelegateOneValidator(k Keepers, ctx sdk.Context, msMsgServer stakingtypes.MsgServer, oldValStr, newValStr string) {
	oldVal, err := sdk.ValAddressFromBech32(oldValStr)
	if err != nil {
		panic(fmt.Errorf("leaked validator redelegation: invalid old validator %q: %w", oldValStr, err))
	}

	delegations, err := k.StakingKeeper.GetValidatorDelegations(ctx, oldVal)
	if err != nil {
		panic(fmt.Errorf("leaked validator redelegation: failed to list delegations for %s: %w", oldValStr, err))
	}

	// The leaked operator's own self-bond is redelegated last -- see the
	// doc comment above for why order doesn't affect correctness, only
	// readability of the resulting event/log order.
	operatorAcc := sdk.AccAddress(oldVal).String()
	var operatorDel *stakingtypes.Delegation
	for i := range delegations {
		if delegations[i].DelegatorAddress == operatorAcc {
			operatorDel = &delegations[i]
			continue
		}
		redelegateOneDelegation(k, ctx, msMsgServer, delegations[i], oldValStr, newValStr)
	}
	if operatorDel != nil {
		redelegateOneDelegation(k, ctx, msMsgServer, *operatorDel, oldValStr, newValStr)
	}
}

func redelegateOneDelegation(k Keepers, ctx sdk.Context, msMsgServer stakingtypes.MsgServer, del stakingtypes.Delegation, oldValStr, newValStr string) {
	lockID := multistakingtypes.MultiStakingLockID(del.DelegatorAddress, oldValStr)
	lock, found := k.MultiStakingKeeper.GetMultiStakingLock(ctx, lockID)
	if !found {
		panic(fmt.Errorf("leaked validator redelegation: no multi-staking lock for delegator %s on %s", del.DelegatorAddress, oldValStr))
	}

	_, err := msMsgServer.BeginRedelegate(ctx, &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    del.DelegatorAddress,
		ValidatorSrcAddress: oldValStr,
		ValidatorDstAddress: newValStr,
		Amount:              sdk.NewCoin(lock.LockedCoin.Denom, lock.LockedCoin.Amount),
	})
	if err != nil {
		panic(fmt.Errorf("leaked validator redelegation: failed to redelegate %s from %s to %s: %w", del.DelegatorAddress, oldValStr, newValStr, err))
	}
}
