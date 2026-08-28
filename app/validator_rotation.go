package app

import (
	"encoding/base64"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
)

// ValidatorRotationHeight is the height at which the leaked validators below
// are migrated to their replacement operator address + consensus key. Same
// mechanism as the blacklist fork: no genesis edit, no chain halt — the
// binary swap at this height carries the migration.
//
// TODO(mainnet deploy): set to the real target height once the replacement
// validators' new operator addresses and consensus keys are known, and
// populate validatorRotations below to match. Until then this must stay
// unreachable/inert — ScheduleValidatorRotation runs unconditionally out of
// every BeginBlocker, so any real (non-placeholder) value here or below
// would fire against every other test and devnet in this repo, not just the
// real deployment.
var ValidatorRotationHeight = int64(-1)

// validatorRotations lists each leaked validator's old operator address and
// its replacement identity. NewConsPubKeyB64 is the raw 32-byte ed25519
// consensus pubkey (base64), taken from the new priv_validator_key.json the
// validator operator generates for the replacement node.
//
// Deliberately empty until the real replacement identities for
// realiovaloper18a32el4maw3pqr8xh3yrl9ja4lejs265a5nxtm and
// realiovaloper13jrrtkfuuvzdak6zxmr95hek9c228ug50sdsvs are known — see the
// TODO on ValidatorRotationHeight above. rotateValidators is a no-op when
// this is empty, so it's safe to leave in place across every other test in
// this repo (and it must stay that way until real values are filled in).
var validatorRotations = []struct {
	OldOperator      string
	NewOperator      string
	NewConsPubKeyB64 string
}{}

// pendingValidatorZeroUpdates holds the "old consensus pubkey, power 0" ABCI
// updates captured while migrating validators in BeginBlocker. The staking
// module's own EndBlocker naturally emits the "new pubkey, power N" half (the
// new operator address is genuinely new to the power index, so the normal
// diff picks it up) — but it has no way to know the old identity needs
// zeroing, since by the time EndBlocker runs it has already been removed
// from every index. EndBlocker (app/app.go) appends these to its result.
var pendingValidatorZeroUpdates []abci.ValidatorUpdate

// rotateValidators migrates every entry in validatorRotations from its old
// operator address + consensus key to the new one, carrying over every
// piece of state tied to that validator identity: the Validator record
// itself, all delegations (not just self-delegation — every delegator
// pointed at this validator), unbonding delegations, distribution rewards/
// commission history, and the multistaking lock/unlock/bond-denom records.
// Nothing about token amounts, delegator addresses, or shares changes —
// only the validator identity they point at.
func rotateValidators(app *RealioNetwork, ctx sdk.Context) {
	pendingValidatorZeroUpdates = nil

	for _, r := range validatorRotations {
		oldValAddr, err := sdk.ValAddressFromBech32(r.OldOperator)
		if err != nil {
			panic(fmt.Errorf("validator rotation: invalid old operator %q: %w", r.OldOperator, err))
		}
		newValAddr, err := sdk.ValAddressFromBech32(r.NewOperator)
		if err != nil {
			panic(fmt.Errorf("validator rotation: invalid new operator %q: %w", r.NewOperator, err))
		}

		rawPubKey, err := base64.StdEncoding.DecodeString(r.NewConsPubKeyB64)
		if err != nil {
			panic(fmt.Errorf("validator rotation: invalid new consensus pubkey for %q: %w", r.OldOperator, err))
		}
		newConsPubKeyAny, err := codectypes.NewAnyWithValue(&ed25519.PubKey{Key: rawPubKey})
		if err != nil {
			panic(err)
		}

		// Redelegations (as either src or dst) are not migrated below — as
		// of writing, the target validators have none (verified against
		// the genesis export). That can change before this fork actually
		// runs, so re-check live state rather than trusting the snapshot:
		// fail loudly here, before touching anything, rather than silently
		// leaving a redelegation record pointing at a validator identity
		// that's about to stop existing.
		checkNoRedelegations(app, ctx, oldValAddr)

		migrateValidatorRecord(app, ctx, oldValAddr, newValAddr, newConsPubKeyAny)
		migrateDelegations(app, ctx, oldValAddr, newValAddr)
		migrateUnbondingDelegations(app, ctx, oldValAddr, newValAddr)
		migrateDistribution(app, ctx, oldValAddr, newValAddr)
		migrateMultiStaking(app, ctx, oldValAddr, newValAddr)
	}
}

// migrateValidatorRecord moves the Validator record itself: copies it under
// the new operator address with the new consensus pubkey, fixes up the
// power/cons-address indexes, and captures the "zero out the old key" ABCI
// update (the new key's update is left to EndBlocker's normal diffing).
func migrateValidatorRecord(app *RealioNetwork, ctx sdk.Context, oldValAddr, newValAddr sdk.ValAddress, newConsPubKeyAny *codectypes.Any) {
	sk := app.StakingKeeper
	store := ctx.KVStore(app.keys[stakingtypes.ModuleName])

	validator, err := sk.GetValidator(ctx, oldValAddr)
	if err != nil {
		panic(fmt.Errorf("validator rotation: old validator %s not found: %w", oldValAddr, err))
	}

	// capture "old key -> power 0" BEFORE mutating anything
	pendingValidatorZeroUpdates = append(pendingValidatorZeroUpdates, validator.ABCIValidatorUpdateZero())

	powerReduction := sk.PowerReduction(ctx)

	store.Delete(stakingtypes.GetValidatorsByPowerIndexKey(validator, powerReduction, sk.ValidatorAddressCodec()))
	store.Delete(stakingtypes.GetLastValidatorPowerKey(oldValAddr))

	// same validator data, new identity
	validator.OperatorAddress = newValAddr.String()
	validator.ConsensusPubkey = newConsPubKeyAny

	if err := sk.SetValidator(ctx, validator); err != nil {
		panic(err)
	}
	if err := sk.SetValidatorByConsAddr(ctx, validator); err != nil {
		panic(err)
	}
	if err := sk.SetValidatorByPowerIndex(ctx, validator); err != nil {
		panic(err)
	}

	// fresh signing info for the new consensus address; the old one is left
	// in place as a harmless historical record (see conversation notes: not
	// migrating missed-block history is a deliberate "new key, clean start").
	newConsAddrBytes, err := validator.GetConsAddr()
	if err != nil {
		panic(err)
	}
	newConsAddr := sdk.ConsAddress(newConsAddrBytes)
	if err := app.SlashingKeeper.SetValidatorSigningInfo(ctx, newConsAddr, slashingtypes.ValidatorSigningInfo{
		Address:     newConsAddr.String(),
		StartHeight: ctx.BlockHeight(),
	}); err != nil {
		panic(err)
	}
}

func migrateDelegations(app *RealioNetwork, ctx sdk.Context, oldValAddr, newValAddr sdk.ValAddress) {
	sk := app.StakingKeeper

	// A validator's own self-delegation is a Delegation record whose
	// DelegatorAddress is literally its operator address' account form
	// (same 20 bytes, "realio1..." instead of "realiovaloper1...") — that's
	// how self-delegation is identified everywhere in x/staking (e.g. the
	// MinSelfDelegation check). Third-party delegators keep their own
	// address untouched below; the self-delegation is the one case that
	// must be re-pointed to the NEW operator's account too — otherwise
	// it stays permanently attributed to the leaked old account, which is
	// presumably being blacklisted right alongside this rotation, meaning
	// nobody could ever manage that stake again (not even the legitimate
	// new validator identity this whole migration exists to hand control
	// to).
	oldSelfDelegator := sdk.AccAddress(oldValAddr).String()
	newSelfDelegator := sdk.AccAddress(newValAddr).String()

	delegations, err := sk.GetValidatorDelegations(ctx, oldValAddr)
	if err != nil {
		panic(err)
	}
	for _, del := range delegations {
		if err := sk.RemoveDelegation(ctx, del); err != nil {
			panic(err)
		}
		del.ValidatorAddress = newValAddr.String()
		if del.DelegatorAddress == oldSelfDelegator {
			del.DelegatorAddress = newSelfDelegator
		}
		if err := sk.SetDelegation(ctx, del); err != nil {
			panic(err)
		}
	}
}

func migrateUnbondingDelegations(app *RealioNetwork, ctx sdk.Context, oldValAddr, newValAddr sdk.ValAddress) {
	sk := app.StakingKeeper
	store := ctx.KVStore(app.keys[stakingtypes.ModuleName])

	// same self-delegator remap as migrateDelegations, for the case where
	// part of the self-bond is already mid-unbonding.
	oldSelfDelegator := sdk.AccAddress(oldValAddr).String()
	newSelfDelegator := sdk.AccAddress(newValAddr).String()

	ubds, err := sk.GetUnbondingDelegationsFromValidator(ctx, oldValAddr)
	if err != nil {
		panic(err)
	}
	for _, ubd := range ubds {
		oldDelegatorAddr := ubd.DelegatorAddress
		oldValidatorAddr := ubd.ValidatorAddress

		newDelegatorAddr := oldDelegatorAddr
		if oldDelegatorAddr == oldSelfDelegator {
			newDelegatorAddr = newSelfDelegator
		}
		newDelegatorAccAddr, err := sdk.AccAddressFromBech32(newDelegatorAddr)
		if err != nil {
			panic(err)
		}

		if err := sk.RemoveUnbondingDelegation(ctx, ubd); err != nil {
			panic(err)
		}
		ubd.ValidatorAddress = newValAddr.String()
		ubd.DelegatorAddress = newDelegatorAddr
		if err := sk.SetUnbondingDelegation(ctx, ubd); err != nil {
			panic(err)
		}

		// EndBlocker's maturity processing (x/staking's BlockValidatorUpdates)
		// looks up who to pay by DEQUEUING a DVPair from the completion-time
		// queue, not by reading the UBD record's own address fields — so
		// moving the record above is not enough on its own: the queue
		// entries still point at the old (delegator, validator) pair. Left
		// unfixed, at maturity CompleteUnbonding(old delegator, old
		// validator) fails to find the record (already moved), the error is
		// silently swallowed, and the queue entry is dequeued regardless —
		// permanently orphaning the funds, since nothing ever re-enqueues it.
		seenCompletionTimes := map[time.Time]bool{}
		for _, entry := range ubd.Entries {
			ct := entry.CompletionTime
			if !seenCompletionTimes[ct] {
				seenCompletionTimes[ct] = true
				slice, err := sk.GetUBDQueueTimeSlice(ctx, ct)
				if err != nil {
					panic(err)
				}
				replaced := false
				for i, p := range slice {
					if p.DelegatorAddress == oldDelegatorAddr && p.ValidatorAddress == oldValidatorAddr {
						slice[i] = stakingtypes.DVPair{DelegatorAddress: newDelegatorAddr, ValidatorAddress: newValAddr.String()}
						replaced = true
					}
				}
				if !replaced {
					panic(fmt.Errorf("validator rotation: UBD queue entry for delegator=%s validator=%s at %s not found",
						oldDelegatorAddr, oldValidatorAddr, ct))
				}
				if err := sk.SetUBDQueueTimeSlice(ctx, ct, slice); err != nil {
					panic(err)
				}
			}

			// Re-point the unbonding-ID index (GetUnbondingDelegationByUnbondingID)
			// so anything holding a reference by ID — e.g. an interchain-security
			// or liquid-staking on-hold mechanism — still resolves to the record
			// under its new key, not the now-deleted old one.
			store.Set(stakingtypes.GetUnbondingIndexKey(entry.UnbondingId), stakingtypes.GetUBDKey(newDelegatorAccAddr, newValAddr))
		}
	}
}

// checkNoRedelegations panics if any redelegation currently has valAddr as
// either its source or destination validator. Redelegation migration isn't
// implemented — this exists so that gap fails loudly at fork time instead
// of silently leaving a redelegation record pointing at a validator
// identity that migrateValidatorRecord is about to delete.
func checkNoRedelegations(app *RealioNetwork, ctx sdk.Context, valAddr sdk.ValAddress) {
	sk := app.StakingKeeper
	valAddrStr := valAddr.String()

	var matches []stakingtypes.Redelegation
	err := sk.IterateRedelegations(ctx, func(_ int64, red stakingtypes.Redelegation) bool {
		if red.ValidatorSrcAddress == valAddrStr || red.ValidatorDstAddress == valAddrStr {
			matches = append(matches, red)
		}
		return false
	})
	if err != nil {
		panic(err)
	}
	if len(matches) > 0 {
		panic(fmt.Errorf(
			"validator rotation: %s has %d redelegation(s) as src/dst which migrateValidatorRecord does not handle — "+
				"first: delegator=%s src=%s dst=%s",
			valAddrStr, len(matches), matches[0].DelegatorAddress, matches[0].ValidatorSrcAddress, matches[0].ValidatorDstAddress,
		))
	}
}

func migrateDistribution(app *RealioNetwork, ctx sdk.Context, oldValAddr, newValAddr sdk.ValAddress) {
	dk := app.DistrKeeper

	if rewards, err := dk.GetValidatorCurrentRewards(ctx, oldValAddr); err == nil {
		if err := dk.SetValidatorCurrentRewards(ctx, newValAddr, rewards); err != nil {
			panic(err)
		}
		if err := dk.DeleteValidatorCurrentRewards(ctx, oldValAddr); err != nil {
			panic(err)
		}
	}

	if commission, err := dk.GetValidatorAccumulatedCommission(ctx, oldValAddr); err == nil {
		if err := dk.SetValidatorAccumulatedCommission(ctx, newValAddr, commission); err != nil {
			panic(err)
		}
		if err := dk.DeleteValidatorAccumulatedCommission(ctx, oldValAddr); err != nil {
			panic(err)
		}
	}

	if outstanding, err := dk.GetValidatorOutstandingRewards(ctx, oldValAddr); err == nil {
		if err := dk.SetValidatorOutstandingRewards(ctx, newValAddr, outstanding); err != nil {
			panic(err)
		}
		if err := dk.DeleteValidatorOutstandingRewards(ctx, oldValAddr); err != nil {
			panic(err)
		}
	}

	type histEntry struct {
		period  uint64
		rewards distrtypes.ValidatorHistoricalRewards
	}
	var hist []histEntry
	dk.IterateValidatorHistoricalRewards(ctx, func(val sdk.ValAddress, period uint64, rewards distrtypes.ValidatorHistoricalRewards) bool {
		if val.Equals(oldValAddr) {
			hist = append(hist, histEntry{period, rewards})
		}
		return false
	})
	for _, h := range hist {
		if err := dk.SetValidatorHistoricalRewards(ctx, newValAddr, h.period, h.rewards); err != nil {
			panic(err)
		}
	}
	dk.DeleteValidatorHistoricalRewards(ctx, oldValAddr)

	// same self-delegator remap as migrateDelegations — must stay in sync
	// with it, or the Delegation record (now under the new account) and its
	// DelegatorStartingInfo (reward accounting) end up keyed to different
	// delegator addresses for the same stake.
	oldSelfDelegator := sdk.AccAddress(oldValAddr)
	newSelfDelegator := sdk.AccAddress(newValAddr)

	type startingInfoEntry struct {
		del  sdk.AccAddress
		info distrtypes.DelegatorStartingInfo
	}
	var starts []startingInfoEntry
	dk.IterateDelegatorStartingInfos(ctx, func(val sdk.ValAddress, del sdk.AccAddress, info distrtypes.DelegatorStartingInfo) bool {
		if val.Equals(oldValAddr) {
			starts = append(starts, startingInfoEntry{del, info})
		}
		return false
	})
	for _, s := range starts {
		newDel := s.del
		if newDel.Equals(oldSelfDelegator) {
			newDel = newSelfDelegator
		}
		if err := dk.SetDelegatorStartingInfo(ctx, newValAddr, newDel, s.info); err != nil {
			panic(err)
		}
		if err := dk.DeleteDelegatorStartingInfo(ctx, oldValAddr, s.del); err != nil {
			panic(err)
		}
	}

	// ValidatorSlashEvent history: CalculateDelegationRewards walks slash
	// events between a delegator's starting period and the withdrawal
	// period to discount rewards across any slash the validator suffered
	// in between. It looks these up by the CURRENT operator address, so if
	// this history stays keyed to the old one, any future reward
	// calculation that spans a period predating a real historical slash of
	// this validator would silently skip applying it — no error, just an
	// overpayment.
	type slashEventEntry struct {
		height uint64
		period uint64
		event  distrtypes.ValidatorSlashEvent
	}
	var slashEvents []slashEventEntry
	dk.IterateValidatorSlashEvents(ctx, func(val sdk.ValAddress, height uint64, event distrtypes.ValidatorSlashEvent) bool {
		if val.Equals(oldValAddr) {
			slashEvents = append(slashEvents, slashEventEntry{height, event.ValidatorPeriod, event})
		}
		return false
	})
	for _, e := range slashEvents {
		if err := dk.SetValidatorSlashEvent(ctx, newValAddr, e.height, e.period, e.event); err != nil {
			panic(err)
		}
	}
	dk.DeleteValidatorSlashEvents(ctx, oldValAddr)
}

func migrateMultiStaking(app *RealioNetwork, ctx sdk.Context, oldValAddr, newValAddr sdk.ValAddress) {
	msk := app.MultiStakingKeeper
	store := ctx.KVStore(app.keys[multistakingtypes.ModuleName])

	// same self-staker remap as migrateDelegations: MultiStakingLock/Unlock
	// are keyed by {MultiStakerAddr, ValAddr} exactly like a Delegation is
	// keyed by {DelegatorAddress, ValidatorAddress} — the self-bond lock
	// must move to the new account, not stay under the old (leaked) one.
	oldSelfStaker := sdk.AccAddress(oldValAddr).String()
	newSelfStaker := sdk.AccAddress(newValAddr).String()

	if denom := msk.GetValidatorMultiStakingCoin(ctx, oldValAddr); denom != "" {
		msk.SetValidatorMultiStakingCoin(ctx, newValAddr, denom)
		store.Delete(multistakingtypes.GetValidatorMultiStakingCoinKey(oldValAddr))
	}

	type lockEntry struct {
		lock multistakingtypes.MultiStakingLock
	}
	var locks []lockEntry
	msk.MultiStakingLockIterator(ctx, func(lock multistakingtypes.MultiStakingLock) bool {
		if lock.LockID.ValAddr == oldValAddr.String() {
			locks = append(locks, lockEntry{lock})
		}
		return false
	})
	for _, l := range locks {
		msk.RemoveMultiStakingLock(ctx, l.lock.LockID)
		l.lock.LockID.ValAddr = newValAddr.String()
		if l.lock.LockID.MultiStakerAddr == oldSelfStaker {
			l.lock.LockID.MultiStakerAddr = newSelfStaker
		}
		msk.SetMultiStakingLock(ctx, l.lock)
	}

	type unlockEntry struct {
		unlock multistakingtypes.MultiStakingUnlock
	}
	var unlocks []unlockEntry
	msk.MultiStakingUnlockIterator(ctx, func(unlock multistakingtypes.MultiStakingUnlock) bool {
		if unlock.UnlockID.ValAddr == oldValAddr.String() {
			unlocks = append(unlocks, unlockEntry{unlock})
		}
		return false
	})
	for _, u := range unlocks {
		msk.DeleteMultiStakingUnlock(ctx, u.unlock.UnlockID)
		u.unlock.UnlockID.ValAddr = newValAddr.String()
		if u.unlock.UnlockID.MultiStakerAddr == oldSelfStaker {
			u.unlock.UnlockID.MultiStakerAddr = newSelfStaker
		}
		msk.SetMultiStakingUnlock(ctx, u.unlock)
	}
}
