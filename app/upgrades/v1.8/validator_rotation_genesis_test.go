package v8_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	"github.com/realiotech/realio-network/app"
	v8 "github.com/realiotech/realio-network/app/upgrades/v1.8"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

// genesisDelegationCounts reads rotationGenesisPath directly (independent of
// the keepers under test) and returns how many delegations the raw genesis
// holds per validator operator address.
func genesisDelegationCounts(t *testing.T) map[string]int {
	t.Helper()

	raw, err := os.ReadFile(rotationGenesisPath)
	require.NoError(t, err)

	var doc struct {
		AppState struct {
			Multistaking struct {
				StakingGenesisState struct {
					Delegations []struct {
						ValidatorAddress string `json:"validator_address"`
					} `json:"delegations"`
				} `json:"staking_genesis_state"`
			} `json:"multistaking"`
		} `json:"app_state"`
	}
	require.NoError(t, json.Unmarshal(raw, &doc))

	counts := map[string]int{}
	for _, d := range doc.AppState.Multistaking.StakingGenesisState.Delegations {
		counts[d.ValidatorAddress]++
	}
	return counts
}

type delegationSnapshot struct {
	delegator sdk.AccAddress
	tokens    math.LegacyDec
	lock      multistakingtypes.MultiStakingLock
}

func totalValidatorTokens(t *testing.T, realioApp *app.RealioNetwork, ctx sdk.Context) math.Int {
	t.Helper()
	vals, err := realioApp.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	total := math.ZeroInt()
	for _, v := range vals {
		total = total.Add(v.Tokens)
	}
	return total
}

func poolBalances(t *testing.T, realioApp *app.RealioNetwork, ctx sdk.Context) (bonded, notBonded math.Int) {
	t.Helper()
	bondDenom, err := realioApp.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	bondedAddr := realioApp.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName)
	notBondedAddr := realioApp.AccountKeeper.GetModuleAddress(stakingtypes.NotBondedPoolName)
	return realioApp.BankKeeper.GetBalance(ctx, bondedAddr, bondDenom).Amount,
		realioApp.BankKeeper.GetBalance(ctx, notBondedAddr, bondDenom).Amount
}

// TestRotateValidatorsGenesisStateDetail reconciles the rotation against the
// genesis export delegator by delegator, not just in aggregate: every
// delegation on each outgoing validator must arrive on its replacement with
// the same token amount, the same multi-staking lock, and a redelegation
// record whose figures match, while nothing leaks out of or into the staking
// pools.
func TestRotateValidatorsGenesisStateDetail(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := setupRotationGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	rotations := rotationToTestValidators(t, realioApp, ctx)
	genesisCounts := genesisDelegationCounts(t)

	// ---- snapshot ----
	pre := make([][]delegationSnapshot, len(rotations))
	preValTokens := make([]math.Int, len(rotations))
	for i, r := range rotations {
		oldValStr := r.OldVal.String()
		oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
		require.NoError(t, err)
		preValTokens[i] = oldValidator.Tokens
		coin := realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, r.OldVal)

		dels, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, r.OldVal)
		require.NoError(t, err)
		require.Len(t, dels, genesisCounts[oldValStr],
			"keeper holds a different number of delegations on %s than the raw genesis file", oldValStr)

		sumTokens := math.LegacyZeroDec()
		for _, d := range dels {
			delAddr, err := sdk.AccAddressFromBech32(d.DelegatorAddress)
			require.NoError(t, err)

			lock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(d.DelegatorAddress, oldValStr))
			require.Truef(t, found, "delegator %s has a delegation on %s but no multi-staking lock", d.DelegatorAddress, oldValStr)
			require.Equalf(t, coin, lock.LockedCoin.Denom, "lock denom for %s on %s differs from the validator's coin", d.DelegatorAddress, oldValStr)

			tokens := oldValidator.TokensFromShares(d.Shares)
			sumTokens = sumTokens.Add(tokens)
			pre[i] = append(pre[i], delegationSnapshot{delegator: delAddr, tokens: tokens, lock: lock})
		}
		require.Equal(t, oldValidator.Tokens.String(), sumTokens.TruncateInt().String(),
			"delegations on %s don't add up to the validator's tokens", oldValStr)
		t.Logf("validator %s (%s): %d delegations, %s tokens, all with matching locks", oldValStr, coin, len(dels), oldValidator.Tokens)
	}

	totalBefore := totalValidatorTokens(t, realioApp, ctx)
	bondedBefore, notBondedBefore := poolBalances(t, realioApp, ctx)
	paramsBefore, err := realioApp.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)

	// ---- rotate ----
	v8.RotateValidators(ctx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper)

	paramsAfter, err := realioApp.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, paramsBefore, paramsAfter, "the rotation must leave the staking params exactly as it found them")

	// ---- reconcile ----
	unbondingTime, err := realioApp.StakingKeeper.UnbondingTime(ctx)
	require.NoError(t, err)

	for i, r := range rotations {
		oldValStr, newValStr := r.OldVal.String(), r.NewVal.String()

		oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
		require.NoError(t, err)
		require.True(t, oldValidator.Tokens.IsZero(), "%s still holds tokens", oldValStr)
		require.True(t, oldValidator.DelegatorShares.IsZero(), "%s still holds delegator shares", oldValStr)
		require.True(t, oldValidator.Jailed, "%s not jailed", oldValStr)

		newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
		require.NoError(t, err)
		require.Equal(t, preValTokens[i].Add(math.NewInt(1_000_000_000_000)).String(), newValidator.Tokens.String(),
			"%s should hold exactly its own self-bond plus everything that left %s", newValStr, oldValStr)
		require.Equal(t, newValidator.Tokens.String(), newValidator.DelegatorShares.TruncateInt().String(),
			"%s exchange rate drifted from 1:1", newValStr)

		newCoin := realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, r.NewVal)

		// A blacklisted account can't sign anything, so it can never have a
		// redelegation in progress that would get it skipped: the outgoing
		// operator and every blacklisted delegator must have moved.
		require.True(t, realioApp.BlacklistKeeper.IsBlacklisted(ctx, sdk.AccAddress(r.OldVal)),
			"fixture assumption: the outgoing operator %s is blacklisted", oldValStr)
		blacklistedMoved := 0
		for _, s := range pre[i] {
			who := s.delegator.String()
			if realioApp.BlacklistKeeper.IsBlacklisted(ctx, s.delegator) {
				_, err := realioApp.StakingKeeper.GetDelegation(ctx, s.delegator, r.NewVal)
				require.NoErrorf(t, err, "blacklisted delegator %s was not moved to %s", who, newValStr)
				blacklistedMoved++
			}
		}
		require.NotZero(t, blacklistedMoved, "fixture assumption: some delegators on %s are blacklisted", oldValStr)
		t.Logf("validator %s: operator and %d blacklisted delegator(s) moved like everyone else", oldValStr, blacklistedMoved)

		for _, s := range pre[i] {
			who := s.delegator.String()

			newDel, err := realioApp.StakingKeeper.GetDelegation(ctx, s.delegator, r.NewVal)
			require.NoErrorf(t, err, "%s has no delegation on %s", who, newValStr)
			require.Equalf(t, s.tokens.String(), newValidator.TokensFromShares(newDel.Shares).String(),
				"%s: tokens changed while moving %s -> %s", who, oldValStr, newValStr)

			_, err = realioApp.StakingKeeper.GetDelegation(ctx, s.delegator, r.OldVal)
			require.Errorf(t, err, "%s still has a delegation on %s", who, oldValStr)

			newLock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(who, newValStr))
			require.Truef(t, found, "%s has no multi-staking lock on %s", who, newValStr)
			require.Equalf(t, newCoin, newLock.LockedCoin.Denom, "%s: lock denom on %s", who, newValStr)
			require.Equalf(t, s.lock.LockedCoin.Amount.String(), newLock.LockedCoin.Amount.String(), "%s: locked amount changed", who)

			if oldLock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(who, oldValStr)); found {
				require.Truef(t, oldLock.LockedCoin.Amount.IsZero(), "%s still has %s locked on %s", who, oldLock.LockedCoin.Amount, oldValStr)
			}

			red, err := realioApp.StakingKeeper.GetRedelegation(ctx, s.delegator, r.OldVal, r.NewVal)
			require.NoErrorf(t, err, "%s has no redelegation record %s -> %s", who, oldValStr, newValStr)
			require.Lenf(t, red.Entries, 1, "%s: expected exactly one redelegation entry", who)
			entry := red.Entries[0]
			require.Equal(t, ctx.BlockHeight(), entry.CreationHeight)
			require.Equal(t, s.tokens.TruncateInt().String(), entry.InitialBalance.String(), "%s: redelegation initial balance", who)
			require.Equal(t, newDel.Shares.String(), entry.SharesDst.String(), "%s: redelegation shares", who)
			require.True(t, entry.CompletionTime.Equal(ctx.BlockTime().Add(unbondingTime)),
				"%s: completion time %s, want block time + unbonding time", who, entry.CompletionTime)
		}
		t.Logf("validator %s -> %s: %d delegators reconciled (tokens, locks, redelegation records)", oldValStr, newValStr, len(pre[i]))
	}

	require.Equal(t, totalBefore.String(), totalValidatorTokens(t, realioApp, ctx).String(), "total tokens across validators changed")

	// Staking pools: nothing is minted or burned, and stake only crosses
	// between the bonded and not-bonded pools when a replacement isn't bonded
	// yet (stand-ins created in this test haven't been through an EndBlock,
	// so they're still unbonded; a live, active replacement moves nothing).
	bondedAfter, notBondedAfter := poolBalances(t, realioApp, ctx)
	require.Equal(t, bondedBefore.Add(notBondedBefore).String(), bondedAfter.Add(notBondedAfter).String(),
		"total staked (bonded + not-bonded) changed")

	crossed := math.ZeroInt()
	for i, r := range rotations {
		newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
		require.NoError(t, err)
		if !newValidator.IsBonded() {
			crossed = crossed.Add(preValTokens[i])
		}
	}
	require.Equal(t, bondedBefore.Sub(crossed).String(), bondedAfter.String(), "bonded pool change doesn't match stake moved to unbonded replacements")
	require.Equal(t, notBondedBefore.Add(crossed).String(), notBondedAfter.String(), "not-bonded pool change doesn't match stake moved to unbonded replacements")

	// Nothing is left on the outgoing validators, so a second run is a no-op.
	require.NotPanics(t, func() {
		v8.RotateValidators(ctx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper)
	})
	require.Equal(t, totalBefore.String(), totalValidatorTokens(t, realioApp, ctx).String(), "second run moved tokens")
}

// TestRotateValidatorsPanicsOnBlankNewValidator: the shipped config leaves
// NewValidator blank until the real replacement addresses are known, and the
// upgrade must halt rather than quietly do nothing.
func TestRotateValidatorsPanicsOnBlankNewValidator(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := setupRotationGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	require.Panics(t, func() {
		v8.RotateValidators(ctx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper)
	})
}

// TestRotateValidatorsPanicsOnCoinMismatch: the two outgoing validators run
// on different multi-staking coins (ario / arst); pointing one at a
// replacement that accepts the other coin must fail loudly, and must leave
// no partial state behind for the caller to trip over.
func TestRotateValidatorsPanicsOnCoinMismatch(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := setupRotationGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	orig := v8.ValidatorRotations
	t.Cleanup(func() { v8.ValidatorRotations = orig })

	oldVal, err := sdk.ValAddressFromBech32(orig[0].OldValidator)
	require.NoError(t, err)
	otherVal, err := sdk.ValAddressFromBech32(orig[1].OldValidator)
	require.NoError(t, err)

	oldCoin := realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, oldVal)
	otherCoin := realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, otherVal)
	require.NotEqual(t, oldCoin, otherCoin, "test premise: the two outgoing validators use different coins")

	wrongCoinVal := createTestValidator(t, realioApp, ctx, otherCoin)
	v8.ValidatorRotations = []struct {
		OldValidator string
		NewValidator string
	}{{OldValidator: orig[0].OldValidator, NewValidator: wrongCoinVal.String()}}

	require.Panics(t, func() {
		v8.RotateValidators(ctx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper)
	})
}

// unbondedRun is the state after RotateValidators ran and then every
// delegator that was moved unbonded its entire position from the
// replacement validator, before any of it matured.
type unbondedRun struct {
	app            *app.RealioNetwork
	ctx            sdk.Context
	proposerAddr   []byte
	blockTime      time.Time
	rotationHeight int64
	rotations      []struct{ OldVal, NewVal sdk.ValAddress }

	pre          [][]delegationSnapshot // per rotation: every delegation as it stood before the rotation
	preValTokens []math.Int
	consAddrs    []sdk.ConsAddress // outgoing validators' consensus addresses
	powers       []int64           // and their consensus power before the rotation
	coins        []string          // multi-staking coin of each rotation
}

const replacementSelfBond = 1_000_000_000_000 // what createTestValidator self-bonds

// payoutRoundingTolerance is how many of the smallest coin units a delegator
// may receive short of the amount it unbonded (measured worst case on this
// genesis: 4 units, on positions up to ~1.6e23 units). Moving stake by redelegation
// and then unbonding it converts tokens to shares and back twice, and the
// SDK truncates each conversion, so a payout can be a unit or two light; that
// is ordinary x/staking rounding, not value lost by the rotation.
const payoutRoundingTolerance = 10

// mergePayoutRoundingTolerance is the same allowance for a position that went
// through more conversions: delegate, redelegate-merge, a slash that unbonds
// shares, then a full unbond (measured worst case on this genesis: 12 units,
// on positions of 1e21 units and up).
const mergePayoutRoundingTolerance = 50

// rotateAndUnbondEverything rotates the validators, then has every moved
// delegator undelegate everything from the replacement through the real
// multistaking Undelegate path, requiring each call to succeed.
func rotateAndUnbondEverything(t *testing.T) *unbondedRun {
	t.Helper()

	realioApp, _, initialHeight, proposerAddr, blockTime := setupRotationGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	run := &unbondedRun{
		app:            realioApp,
		ctx:            ctx,
		proposerAddr:   proposerAddr,
		blockTime:      blockTime,
		rotationHeight: ctx.BlockHeight(),
		rotations:      rotationToTestValidators(t, realioApp, ctx),
	}

	for _, r := range run.rotations {
		oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
		require.NoError(t, err)
		consAddrBz, err := oldValidator.GetConsAddr()
		require.NoError(t, err)
		run.preValTokens = append(run.preValTokens, oldValidator.Tokens)
		run.consAddrs = append(run.consAddrs, sdk.ConsAddress(consAddrBz))
		run.powers = append(run.powers, oldValidator.ConsensusPower(sdk.DefaultPowerReduction))
		run.coins = append(run.coins, realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, r.OldVal))

		dels, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, r.OldVal)
		require.NoError(t, err)
		var snaps []delegationSnapshot
		for _, d := range dels {
			delAddr, err := sdk.AccAddressFromBech32(d.DelegatorAddress)
			require.NoError(t, err)
			lock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(d.DelegatorAddress, r.OldVal.String()))
			require.True(t, found)
			snaps = append(snaps, delegationSnapshot{delegator: delAddr, tokens: oldValidator.TokensFromShares(d.Shares), lock: lock})
		}
		run.pre = append(run.pre, snaps)
	}

	v8.RotateValidators(ctx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper)

	msMsgServer := multistakingkeeper.NewMsgServerImpl(realioApp.MultiStakingKeeper)
	for i, r := range run.rotations {
		newValStr := r.NewVal.String()
		for _, s := range run.pre[i] {
			who := s.delegator.String()
			lock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(who, newValStr))
			require.Truef(t, found, "%s has no lock on %s to unbond from", who, newValStr)

			_, err := msMsgServer.Undelegate(ctx, &stakingtypes.MsgUndelegate{
				DelegatorAddress: who,
				ValidatorAddress: newValStr,
				Amount:           sdk.NewCoin(lock.LockedCoin.Denom, lock.LockedCoin.Amount),
			})
			require.NoErrorf(t, err, "%s could not unbond everything from %s right after being rotated onto it", who, newValStr)
		}
	}
	return run
}

// TestRotatedDelegatorsUnbondEverything: every rotated delegator, including
// dust holders and delegators present on both outgoing validators, unbonds
// its entire position from the replacement straight after the rotation.
// All of it must go through, leave the replacements with only their own
// self-bond, keep the redelegation records alive, and then mature and pay
// back through a real EndBlocker without breaking anything.
func TestRotatedDelegatorsUnbondEverything(t *testing.T) {
	run := rotateAndUnbondEverything(t)
	realioApp, ctx := run.app, run.ctx

	for i, r := range run.rotations {
		oldValStr, newValStr := r.OldVal.String(), r.NewVal.String()

		newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
		require.NoError(t, err)
		require.Equal(t, math.NewInt(replacementSelfBond).String(), newValidator.Tokens.String(),
			"%s should be back to just its own self-bond", newValStr)
		require.Equal(t, math.NewInt(replacementSelfBond).String(), newValidator.DelegatorShares.TruncateInt().String())

		remaining, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, r.NewVal)
		require.NoError(t, err)
		require.Len(t, remaining, 1, "only the replacement's own self-bond should remain on %s", newValStr)

		oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
		require.NoError(t, err)
		require.True(t, oldValidator.Tokens.IsZero())

		unbondingTotal := math.ZeroInt()
		for _, s := range run.pre[i] {
			who := s.delegator.String()

			_, err := realioApp.StakingKeeper.GetDelegation(ctx, s.delegator, r.NewVal)
			require.Errorf(t, err, "%s still has a delegation on %s", who, newValStr)

			if lock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(who, newValStr)); found {
				require.Truef(t, lock.LockedCoin.Amount.IsZero(), "%s still has %s locked on %s", who, lock.LockedCoin.Amount, newValStr)
			}

			ubd, err := realioApp.StakingKeeper.GetUnbondingDelegation(ctx, s.delegator, r.NewVal)
			require.NoErrorf(t, err, "%s has no unbonding delegation on %s", who, newValStr)
			require.Len(t, ubd.Entries, 1)
			require.Equal(t, s.tokens.TruncateInt().String(), ubd.Entries[0].InitialBalance.String(), "%s: unbonded amount", who)
			unbondingTotal = unbondingTotal.Add(ubd.Entries[0].Balance)

			red, err := realioApp.StakingKeeper.GetRedelegation(ctx, s.delegator, r.OldVal, r.NewVal)
			require.NoErrorf(t, err, "%s: unbonding must not erase the redelegation record %s -> %s", who, oldValStr, newValStr)
			require.Len(t, red.Entries, 1)
		}
		require.Equal(t, run.preValTokens[i].String(), unbondingTotal.String(), "unbonding balances on %s don't add up to what left %s", newValStr, oldValStr)
	}

	// Let everything mature through a real EndBlocker.
	unbondingTime, err := realioApp.StakingKeeper.UnbondingTime(ctx)
	require.NoError(t, err)

	balanceBefore := map[string]sdk.Coin{}
	for i, r := range run.rotations {
		for _, s := range run.pre[i] {
			balanceBefore[s.delegator.String()+"/"+r.NewVal.String()] = realioApp.BankKeeper.GetBalance(ctx, s.delegator, run.coins[i])
		}
	}

	matureCtx := app.NewHeaderCtx(realioApp, run.rotationHeight+2, run.proposerAddr, run.blockTime.Add(unbondingTime+24*time.Hour))
	require.NotPanics(t, func() {
		_, err := realioApp.EndBlocker(matureCtx)
		require.NoError(t, err)
	})

	maxShortfall := math.ZeroInt()
	for i, r := range run.rotations {
		for _, s := range run.pre[i] {
			who := s.delegator.String()

			_, err := realioApp.StakingKeeper.GetUnbondingDelegation(matureCtx, s.delegator, r.NewVal)
			require.Errorf(t, err, "%s: unbonding delegation on %s should have completed", who, r.NewVal)

			_, err = realioApp.StakingKeeper.GetRedelegation(matureCtx, s.delegator, r.OldVal, r.NewVal)
			require.Errorf(t, err, "%s: redelegation record %s -> %s should have expired", who, r.OldVal, r.NewVal)

			// Rewards may top the balance up too, so this is a floor.
			before := balanceBefore[who+"/"+r.NewVal.String()]
			after := realioApp.BankKeeper.GetBalance(matureCtx, s.delegator, run.coins[i])
			floor := before.Amount.Add(s.lock.LockedCoin.Amount).SubRaw(payoutRoundingTolerance)
			require.Truef(t, after.Amount.GTE(floor),
				"%s should get %s %s back, give or take %d unit(s) of rounding (before=%s after=%s)",
				who, s.lock.LockedCoin.Amount, run.coins[i], payoutRoundingTolerance, before.Amount, after.Amount)
			if short := before.Amount.Add(s.lock.LockedCoin.Amount).Sub(after.Amount); short.IsPositive() && short.GT(maxShortfall) {
				maxShortfall = short
			}
		}

		if oldValidator, err := realioApp.StakingKeeper.GetValidator(matureCtx, r.OldVal); err == nil {
			require.True(t, oldValidator.Tokens.IsZero())
			require.False(t, oldValidator.IsBonded(), "%s is jailed and empty, it must not stay bonded", r.OldVal)
		}
		newValidator, err := realioApp.StakingKeeper.GetValidator(matureCtx, r.NewVal)
		require.NoError(t, err)
		require.Equal(t, math.NewInt(replacementSelfBond).String(), newValidator.Tokens.String())
	}
	t.Logf("largest payout shortfall across all delegators: %s unit(s)", maxShortfall)
}

// TestSlashAfterRotatedDelegatorsUnbondedEverything: evidence for a
// pre-rotation infraction of an outgoing validator arrives after every moved
// delegator has already unbonded from the replacement. The slash must reach
// the unbonding entries (the stake now lives there), burn exactly the slash
// fraction of each, not touch the replacement's own self-bond, and still let
// everything mature and pay out the reduced amounts.
func TestSlashAfterRotatedDelegatorsUnbondedEverything(t *testing.T) {
	run := rotateAndUnbondEverything(t)
	realioApp, ctx := run.app, run.ctx

	slashFactor := math.LegacyNewDecWithPrec(5, 2)
	infractionHeight := run.rotationHeight - 100

	expectedBurn := math.ZeroInt()
	expectedBalance := map[string]math.Int{}
	for i, r := range run.rotations {
		for _, s := range run.pre[i] {
			initial := s.tokens.TruncateInt()
			slashed := slashFactor.MulInt(initial).TruncateInt()
			expectedBurn = expectedBurn.Add(slashed)
			expectedBalance[s.delegator.String()+"/"+r.NewVal.String()] = initial.Sub(slashed)
		}
	}

	bondedBefore, notBondedBefore := poolBalances(t, realioApp, ctx)

	for i, r := range run.rotations {
		_, err := realioApp.StakingKeeper.Slash(ctx, run.consAddrs[i], infractionHeight, run.powers[i], slashFactor)
		require.NoErrorf(t, err, "slashing %s after its delegators unbonded from the replacement", r.OldVal)
	}

	for i, r := range run.rotations {
		newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
		require.NoError(t, err)
		require.Equal(t, math.NewInt(replacementSelfBond).String(), newValidator.Tokens.String(),
			"the slash must be absorbed by the unbonding entries, not the replacement's own self-bond")

		for _, s := range run.pre[i] {
			ubd, err := realioApp.StakingKeeper.GetUnbondingDelegation(ctx, s.delegator, r.NewVal)
			require.NoError(t, err)
			require.Len(t, ubd.Entries, 1)
			require.Equal(t, expectedBalance[s.delegator.String()+"/"+r.NewVal.String()].String(), ubd.Entries[0].Balance.String(),
				"%s: unbonding balance after a %s slash", s.delegator, slashFactor)
			require.Equal(t, s.tokens.TruncateInt().String(), ubd.Entries[0].InitialBalance.String(), "initial balance must stay as recorded")
		}
	}

	bondedAfter, notBondedAfter := poolBalances(t, realioApp, ctx)
	require.Equal(t, bondedBefore.String(), bondedAfter.String(), "a slash of unbonding entries must not touch the bonded pool")
	require.Equal(t, notBondedBefore.Sub(expectedBurn).String(), notBondedAfter.String(), "not-bonded pool must lose exactly the slashed amount")

	// The reduced amounts still mature and pay out.
	unbondingTime, err := realioApp.StakingKeeper.UnbondingTime(ctx)
	require.NoError(t, err)

	balanceBefore := map[string]sdk.Coin{}
	for i, r := range run.rotations {
		for _, s := range run.pre[i] {
			balanceBefore[s.delegator.String()+"/"+r.NewVal.String()] = realioApp.BankKeeper.GetBalance(ctx, s.delegator, run.coins[i])
		}
	}

	matureCtx := app.NewHeaderCtx(realioApp, run.rotationHeight+2, run.proposerAddr, run.blockTime.Add(unbondingTime+24*time.Hour))
	require.NotPanics(t, func() {
		_, err := realioApp.EndBlocker(matureCtx)
		require.NoError(t, err)
	})

	for i, r := range run.rotations {
		for _, s := range run.pre[i] {
			key := s.delegator.String() + "/" + r.NewVal.String()
			_, err := realioApp.StakingKeeper.GetUnbondingDelegation(matureCtx, s.delegator, r.NewVal)
			require.Errorf(t, err, "%s: unbonding on %s should have completed", s.delegator, r.NewVal)

			after := realioApp.BankKeeper.GetBalance(matureCtx, s.delegator, run.coins[i])
			floor := balanceBefore[key].Amount.Add(expectedBalance[key]).SubRaw(payoutRoundingTolerance)
			require.Truef(t, after.Amount.GTE(floor),
				"%s should receive the post-slash %s %s, give or take %d unit(s) of rounding (before=%s after=%s)",
				s.delegator, expectedBalance[key], run.coins[i], payoutRoundingTolerance, balanceBefore[key].Amount, after.Amount)
		}
	}
}

// TestRotateValidatorsMergesIntoExistingDelegation covers delegators who
// already hold a delegation on the replacement validator when the rotation
// moves their old stake onto it: half of the moved delegators are given such
// a position first, so the redelegation has to merge into it.
//
// It checks the merge itself (amounts, locks, one redelegation entry for just
// the moved part), then that a pre-rotation slash reaches only the moved part
// and never the pre-existing position, and finally that the merged positions
// can be unbonded in full and pay back exactly what is left.
func TestRotateValidatorsMergesIntoExistingDelegation(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := setupRotationGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	rotations := rotationToTestValidators(t, realioApp, ctx)
	msMsgServer := multistakingkeeper.NewMsgServerImpl(realioApp.MultiStakingKeeper)

	type mover struct {
		delegator      sdk.AccAddress
		moved          math.Int // tokens leaving the outgoing validator
		movedLock      math.Int
		existing       math.Int // tokens already on the replacement, 0 if none
		existingLock   math.Int
		afterRotation  math.Int
		afterSlash     math.Int
		balanceBeforeM math.Int
	}
	movers := make([][]*mover, len(rotations))
	consAddrs := make([]sdk.ConsAddress, len(rotations))
	powers := make([]int64, len(rotations))
	coins := make([]string, len(rotations))
	merged := 0

	for i, r := range rotations {
		oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
		require.NoError(t, err)
		consAddrBz, err := oldValidator.GetConsAddr()
		require.NoError(t, err)
		consAddrs[i], powers[i] = sdk.ConsAddress(consAddrBz), oldValidator.ConsensusPower(sdk.DefaultPowerReduction)
		coins[i] = realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, r.OldVal)

		dels, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, r.OldVal)
		require.NoError(t, err)
		for j, d := range dels {
			delAddr, err := sdk.AccAddressFromBech32(d.DelegatorAddress)
			require.NoError(t, err)
			lock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(d.DelegatorAddress, r.OldVal.String()))
			require.True(t, found)

			m := &mover{
				delegator:    delAddr,
				moved:        oldValidator.TokensFromShares(d.Shares).TruncateInt(),
				movedLock:    lock.LockedCoin.Amount,
				existing:     math.ZeroInt(),
				existingLock: math.ZeroInt(),
			}

			if j%2 == 0 {
				stake := sdk.NewCoin(coins[i], lock.LockedCoin.Amount.QuoRaw(4))
				require.NoError(t, realioApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(stake)))
				require.NoError(t, realioApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delAddr, sdk.NewCoins(stake)))
				_, err := msMsgServer.Delegate(ctx, &stakingtypes.MsgDelegate{
					DelegatorAddress: d.DelegatorAddress,
					ValidatorAddress: r.NewVal.String(),
					Amount:           stake,
				})
				require.NoError(t, err)

				newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
				require.NoError(t, err)
				existingDel, err := realioApp.StakingKeeper.GetDelegation(ctx, delAddr, r.NewVal)
				require.NoError(t, err, "premise: %s must already hold a delegation on %s", delAddr, r.NewVal)
				existingLock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(d.DelegatorAddress, r.NewVal.String()))
				require.True(t, found)
				m.existing = newValidator.TokensFromShares(existingDel.Shares).TruncateInt()
				m.existingLock = existingLock.LockedCoin.Amount
				require.True(t, m.existing.IsPositive())
				merged++
			}
			movers[i] = append(movers[i], m)
		}
	}
	require.NotZero(t, merged)

	preNewTokens := make([]math.Int, len(rotations))
	for i, r := range rotations {
		v, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
		require.NoError(t, err)
		preNewTokens[i] = v.Tokens
	}

	v8.RotateValidators(ctx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper)

	// ---- merge ----
	for i, r := range rotations {
		newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
		require.NoError(t, err)

		movedTotal := math.ZeroInt()
		for _, m := range movers[i] {
			who := m.delegator.String()
			movedTotal = movedTotal.Add(m.moved)

			del, err := realioApp.StakingKeeper.GetDelegation(ctx, m.delegator, r.NewVal)
			require.NoError(t, err)
			m.afterRotation = newValidator.TokensFromShares(del.Shares).TruncateInt()
			require.Equalf(t, m.existing.Add(m.moved).String(), m.afterRotation.String(),
				"%s: merged position should be its existing %s plus the %s moved", who, m.existing, m.moved)

			lock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(who, r.NewVal.String()))
			require.True(t, found)
			require.Equalf(t, m.existingLock.Add(m.movedLock).String(), lock.LockedCoin.Amount.String(),
				"%s: merged lock should be existing %s plus moved %s", who, m.existingLock, m.movedLock)

			red, err := realioApp.StakingKeeper.GetRedelegation(ctx, m.delegator, r.OldVal, r.NewVal)
			require.NoError(t, err)
			require.Len(t, red.Entries, 1)
			require.Equalf(t, m.moved.String(), red.Entries[0].InitialBalance.String(),
				"%s: the redelegation record must cover only the moved part, not the pre-existing position", who)
			require.Equal(t, m.moved.String(), red.Entries[0].SharesDst.TruncateInt().String(), "%s: shares gained by the redelegation", who)
		}
		require.Equal(t, preNewTokens[i].Add(movedTotal).String(), newValidator.Tokens.String(),
			"%s should hold its previous tokens plus everything that moved", r.NewVal)
	}

	// ---- slash for a pre-rotation infraction ----
	slashFactor := math.LegacyNewDecWithPrec(5, 2)
	for i, r := range rotations {
		_, err := realioApp.StakingKeeper.Slash(ctx, consAddrs[i], ctx.BlockHeight()-100, powers[i], slashFactor)
		require.NoError(t, err)

		newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
		require.NoError(t, err)
		for _, m := range movers[i] {
			del, err := realioApp.StakingKeeper.GetDelegation(ctx, m.delegator, r.NewVal)
			require.NoError(t, err)
			m.afterSlash = newValidator.TokensFromShares(del.Shares).TruncateInt()

			// Only the moved part is on the hook; the pre-existing position stays whole.
			want := m.existing.Add(m.moved.Sub(slashFactor.MulInt(m.moved).TruncateInt()))
			require.Truef(t, m.afterSlash.Sub(want).Abs().LTE(math.NewInt(payoutRoundingTolerance)),
				"%s: after a %s slash want ~%s (existing %s + 95%% of moved %s), got %s", m.delegator, slashFactor, want, m.existing, m.moved, m.afterSlash)
			require.Truef(t, m.afterSlash.GTE(m.existing),
				"%s: the slash reached into the pre-existing position (%s < %s)", m.delegator, m.afterSlash, m.existing)
		}
	}

	// ---- unbond the whole merged position, let it mature ----
	for i, r := range rotations {
		for _, m := range movers[i] {
			lock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(m.delegator.String(), r.NewVal.String()))
			require.True(t, found)
			_, err := msMsgServer.Undelegate(ctx, &stakingtypes.MsgUndelegate{
				DelegatorAddress: m.delegator.String(),
				ValidatorAddress: r.NewVal.String(),
				Amount:           sdk.NewCoin(coins[i], lock.LockedCoin.Amount),
			})
			require.NoErrorf(t, err, "%s could not unbond its merged position from %s after the slash", m.delegator, r.NewVal)
			m.balanceBeforeM = realioApp.BankKeeper.GetBalance(ctx, m.delegator, coins[i]).Amount
		}
	}

	unbondingTime, err := realioApp.StakingKeeper.UnbondingTime(ctx)
	require.NoError(t, err)
	matureCtx := app.NewHeaderCtx(realioApp, ctx.BlockHeight()+2, proposerAddr, blockTime.Add(unbondingTime+24*time.Hour))
	require.NotPanics(t, func() {
		_, err := realioApp.EndBlocker(matureCtx)
		require.NoError(t, err)
	})

	maxDiff := math.ZeroInt()
	for i, r := range rotations {
		for _, m := range movers[i] {
			_, err := realioApp.StakingKeeper.GetUnbondingDelegation(matureCtx, m.delegator, r.NewVal)
			require.Errorf(t, err, "%s: unbonding on %s should have completed", m.delegator, r.NewVal)

			paid := realioApp.BankKeeper.GetBalance(matureCtx, m.delegator, coins[i]).Amount.Sub(m.balanceBeforeM)
			diff := paid.Sub(m.afterSlash).Abs()
			if diff.GT(maxDiff) {
				maxDiff = diff
			}
			require.Truef(t, diff.LTE(math.NewInt(mergePayoutRoundingTolerance)),
				"%s: matured payout %s should equal what was left after the slash (%s), not the pre-slash lock",
				m.delegator, paid, m.afterSlash)
		}
	}
	t.Logf("%d of the moved delegators had a pre-existing delegation to merge into; largest payout difference: %s unit(s)", merged, maxDiff)
}

// TestRotateValidatorsSkipsDelegatorWithIncomingRedelegation: a delegator who,
// shortly before the upgrade, redelegated stake from some other validator INTO
// an outgoing validator still has that redelegation in progress when the
// rotation runs. x/staking refuses to redelegate stake out of a validator the
// delegator is currently receiving a redelegation into
// (ErrTransitiveRedelegation). The rotation must not halt on that: it moves
// everybody else, leaves that one delegation exactly as it was (no half-moved
// lock), logs it, and once the earlier redelegation has matured the delegator
// can unbond and redelegate the leftover themselves.
func TestRotateValidatorsSkipsDelegatorWithIncomingRedelegation(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := setupRotationGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	rotations := rotationToTestValidators(t, realioApp, ctx)
	oldVal, newVal := rotations[0].OldVal, rotations[0].NewVal
	coin := realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, oldVal)
	msMsgServer := multistakingkeeper.NewMsgServerImpl(realioApp.MultiStakingKeeper)

	// A third validator on the same coin that is actually bonded (a redelegation
	// out of an unbonded validator completes at once and leaves no record to
	// block anything), and one existing delegator of the outgoing validator.
	var otherVal sdk.ValAddress
	allVals, err := realioApp.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	for _, v := range allVals {
		valAddr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		require.NoError(t, err)
		if v.IsBonded() && !v.Jailed && !valAddr.Equals(oldVal) &&
			realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, valAddr) == coin {
			otherVal = valAddr
			break
		}
	}
	require.NotNil(t, otherVal, "no bonded %s validator in the genesis to redelegate from", coin)

	dels, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, oldVal)
	require.NoError(t, err)
	operatorAddr := sdk.AccAddress(oldVal).String()
	var blocked sdk.AccAddress
	for _, d := range dels {
		if d.DelegatorAddress != operatorAddr {
			blocked, err = sdk.AccAddressFromBech32(d.DelegatorAddress)
			require.NoError(t, err)
			break
		}
	}
	require.NotNil(t, blocked)

	// The delegator stakes on the third validator, then redelegates that stake
	// into the outgoing validator: the redelegation is now in progress.
	stake := sdk.NewCoin(coin, math.NewInt(1_000_000_000_000_000_000))
	require.NoError(t, realioApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(stake)))
	require.NoError(t, realioApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, blocked, sdk.NewCoins(stake)))
	_, err = msMsgServer.Delegate(ctx, &stakingtypes.MsgDelegate{DelegatorAddress: blocked.String(), ValidatorAddress: otherVal.String(), Amount: stake})
	require.NoError(t, err)
	_, err = msMsgServer.BeginRedelegate(ctx, &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress: blocked.String(), ValidatorSrcAddress: otherVal.String(), ValidatorDstAddress: oldVal.String(), Amount: stake,
	})
	require.NoError(t, err)

	receiving, err := realioApp.StakingKeeper.HasReceivingRedelegation(ctx, blocked, oldVal)
	require.NoError(t, err)
	require.True(t, receiving, "premise: the delegator must have a redelegation in progress into %s", oldVal)

	// ---- snapshot everything the skip must leave untouched ----
	blockedDelBefore, err := realioApp.StakingKeeper.GetDelegation(ctx, blocked, oldVal)
	require.NoError(t, err)
	blockedLockBefore, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(blocked.String(), oldVal.String()))
	require.True(t, found)
	oldValBefore, err := realioApp.StakingKeeper.GetValidator(ctx, oldVal)
	require.NoError(t, err)
	newValBefore, err := realioApp.StakingKeeper.GetValidator(ctx, newVal)
	require.NoError(t, err)
	blockedTokens := oldValBefore.TokensFromShares(blockedDelBefore.Shares).TruncateInt()

	// ---- rotate: must not panic, must log the skip ----
	var logs bytes.Buffer
	logCtx := ctx.WithLogger(log.NewLogger(&logs))
	require.NotPanics(t, func() {
		v8.RotateValidators(logCtx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper)
	})

	require.Contains(t, logs.String(), "skipping delegation", "the skipped delegation must be logged")
	require.Contains(t, logs.String(), blocked.String(), "the log must name the delegator")
	require.Equal(t, 1, strings.Count(logs.String(), "skipping delegation"), "exactly the one blocked delegation should be skipped")

	// The blocked delegation is exactly as it was: same shares, same lock, nothing on the replacement.
	blockedDelAfter, err := realioApp.StakingKeeper.GetDelegation(ctx, blocked, oldVal)
	require.NoError(t, err)
	require.Equal(t, blockedDelBefore.Shares.String(), blockedDelAfter.Shares.String())
	blockedLockAfter, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(blocked.String(), oldVal.String()))
	require.True(t, found)
	require.Equal(t, blockedLockBefore.LockedCoin.Amount.String(), blockedLockAfter.LockedCoin.Amount.String(), "the lock must not have been moved")
	_, err = realioApp.StakingKeeper.GetDelegation(ctx, blocked, newVal)
	require.Error(t, err, "nothing of the blocked delegator should have reached the replacement")
	_, err = realioApp.StakingKeeper.GetRedelegation(ctx, blocked, oldVal, newVal)
	require.Error(t, err)

	// Everyone else on that validator moved, including the operator, who is therefore jailed.
	remaining, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, oldVal)
	require.NoError(t, err)
	require.Len(t, remaining, 1, "only the blocked delegation should be left on %s", oldVal)
	oldValAfter, err := realioApp.StakingKeeper.GetValidator(ctx, oldVal)
	require.NoError(t, err)
	require.Equal(t, blockedTokens.String(), oldValAfter.Tokens.String())
	require.True(t, oldValAfter.Jailed)
	newValAfter, err := realioApp.StakingKeeper.GetValidator(ctx, newVal)
	require.NoError(t, err)
	require.Equal(t, newValBefore.Tokens.Add(oldValBefore.Tokens.Sub(blockedTokens)).String(), newValAfter.Tokens.String(),
		"the replacement should have received everything except the blocked delegation")

	// The other outgoing validator has no such delegator: fully moved.
	otherOld, err := realioApp.StakingKeeper.GetValidator(ctx, rotations[1].OldVal)
	require.NoError(t, err)
	require.True(t, otherOld.Tokens.IsZero())

	// ---- once the earlier redelegation matures, the leftover moves ----
	unbondingTime, err := realioApp.StakingKeeper.UnbondingTime(ctx)
	require.NoError(t, err)
	matureCtx := app.NewHeaderCtx(realioApp, ctx.BlockHeight()+2, proposerAddr, blockTime.Add(unbondingTime+24*time.Hour))
	_, err = realioApp.EndBlocker(matureCtx)
	require.NoError(t, err)

	receiving, err = realioApp.StakingKeeper.HasReceivingRedelegation(matureCtx, blocked, oldVal)
	require.NoError(t, err)
	require.False(t, receiving, "the incoming redelegation should have matured")

	// The comment on the skip promises the delegator can act for themselves
	// from here on: unbond part of the leftover, redelegate the rest.
	lock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(matureCtx, multistakingtypes.MultiStakingLockID(blocked.String(), oldVal.String()))
	require.True(t, found)
	half := lock.LockedCoin.Amount.QuoRaw(2)
	_, err = msMsgServer.Undelegate(matureCtx, &stakingtypes.MsgUndelegate{
		DelegatorAddress: blocked.String(), ValidatorAddress: oldVal.String(), Amount: sdk.NewCoin(coin, half),
	})
	require.NoError(t, err, "once unblocked, the delegator must be able to unbond from the outgoing validator")
	_, err = msMsgServer.BeginRedelegate(matureCtx, &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress: blocked.String(), ValidatorSrcAddress: oldVal.String(), ValidatorDstAddress: newVal.String(),
		Amount: sdk.NewCoin(coin, lock.LockedCoin.Amount.Sub(half)),
	})
	require.NoError(t, err, "once unblocked, the delegator must be able to redelegate to the replacement")

	_, err = realioApp.StakingKeeper.GetDelegation(matureCtx, blocked, newVal)
	require.NoError(t, err)
	leftover, err := realioApp.StakingKeeper.GetValidatorDelegations(matureCtx, oldVal)
	require.NoError(t, err)
	require.Empty(t, leftover, "nothing should be left on the outgoing validator")
}

// TestRotateValidatorsAtMaxRedelegationEntries: a delegator who, in the days
// before the upgrade, already moved stake from the outgoing validator to its
// replacement in as many small steps as x/staking allows (MaxEntries) is one
// entry short for the rotation. It must still be moved in full, and the
// staking params must come back exactly as they were.
func TestRotateValidatorsAtMaxRedelegationEntries(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := setupRotationGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	rotations := rotationToTestValidators(t, realioApp, ctx)
	oldVal, newVal := rotations[0].OldVal, rotations[0].NewVal
	coin := realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, oldVal)
	msMsgServer := multistakingkeeper.NewMsgServerImpl(realioApp.MultiStakingKeeper)

	maxEntries, err := realioApp.StakingKeeper.MaxEntries(ctx)
	require.NoError(t, err)

	dels, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, oldVal)
	require.NoError(t, err)
	var user sdk.AccAddress
	for _, d := range dels {
		if d.DelegatorAddress != sdk.AccAddress(oldVal).String() {
			user, err = sdk.AccAddressFromBech32(d.DelegatorAddress)
			require.NoError(t, err)
			break
		}
	}
	require.NotNil(t, user)

	oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, oldVal)
	require.NoError(t, err)
	userDel, err := realioApp.StakingKeeper.GetDelegation(ctx, user, oldVal)
	require.NoError(t, err)
	userTokens := oldValidator.TokensFromShares(userDel.Shares).TruncateInt()

	// The delegator migrates voluntarily in small steps until it is at the limit.
	step := math.NewInt(1_000_000_000_000)
	for i := uint32(0); i < maxEntries; i++ {
		_, err := msMsgServer.BeginRedelegate(ctx, &stakingtypes.MsgBeginRedelegate{
			DelegatorAddress: user.String(), ValidatorSrcAddress: oldVal.String(), ValidatorDstAddress: newVal.String(),
			Amount: sdk.NewCoin(coin, step),
		})
		require.NoError(t, err)
	}
	red, err := realioApp.StakingKeeper.GetRedelegation(ctx, user, oldVal, newVal)
	require.NoError(t, err)
	require.Len(t, red.Entries, int(maxEntries), "premise: the delegator is at the redelegation entry limit")
	// Probe in a throwaway context: the multistaking msg server moves the lock
	// before x/staking rejects the message, and outside a transaction nothing
	// rolls that half-done move back.
	probeCtx, _ := ctx.CacheContext()
	_, err = msMsgServer.BeginRedelegate(probeCtx, &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress: user.String(), ValidatorSrcAddress: oldVal.String(), ValidatorDstAddress: newVal.String(),
		Amount: sdk.NewCoin(coin, step),
	})
	require.ErrorIs(t, err, stakingtypes.ErrMaxRedelegationEntries, "premise: one more entry is normally refused")

	paramsBefore, err := realioApp.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)

	require.NotPanics(t, func() {
		v8.RotateValidators(ctx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper)
	})

	// Moved in full: nothing left on the outgoing validator, all of it on the replacement.
	_, err = realioApp.StakingKeeper.GetDelegation(ctx, user, oldVal)
	require.Error(t, err, "the delegator should have nothing left on the outgoing validator")
	remaining, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, oldVal)
	require.NoError(t, err)
	require.Empty(t, remaining)

	newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, newVal)
	require.NoError(t, err)
	newDel, err := realioApp.StakingKeeper.GetDelegation(ctx, user, newVal)
	require.NoError(t, err)
	require.Equal(t, userTokens.String(), newValidator.TokensFromShares(newDel.Shares).TruncateInt().String(),
		"the delegator's whole position, voluntary steps plus the rotation, should be on the replacement")

	red, err = realioApp.StakingKeeper.GetRedelegation(ctx, user, oldVal, newVal)
	require.NoError(t, err)
	require.Len(t, red.Entries, int(maxEntries)+1, "the rotation adds exactly one entry on top of the voluntary ones")
	require.Equal(t, userTokens.Sub(step.MulRaw(int64(maxEntries))).String(), red.Entries[len(red.Entries)-1].InitialBalance.String(),
		"the rotation's own entry covers only what was still on the outgoing validator")

	// The extra allowance was temporary.
	paramsAfter, err := realioApp.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, paramsBefore, paramsAfter)
	restored, err := realioApp.StakingKeeper.MaxEntries(ctx)
	require.NoError(t, err)
	require.Equal(t, maxEntries, restored, "MaxEntries must be back to its original value")
	probeCtx, _ = ctx.CacheContext()
	_, err = msMsgServer.BeginRedelegate(probeCtx, &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress: user.String(), ValidatorSrcAddress: oldVal.String(), ValidatorDstAddress: newVal.String(),
		Amount: sdk.NewCoin(coin, step),
	})
	require.Error(t, err, "with the limit restored, an entry beyond it is refused again")
}

// TestRotateValidatorsSkipsDustDelegation: a delegation of a single smallest
// unit is worth one token, until the validator is slashed and each share is
// worth less than that. Its shares still exist but their value truncates to
// zero, which x/staking refuses to redelegate. The rotation must skip it and
// move everybody else instead of halting.
//
// (After a slash every mover also leaves a sub-token residue of shares behind,
// from the truncating conversion at a rate other than 1; that is checked too.)
func TestRotateValidatorsSkipsDustDelegation(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := setupRotationGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	rotations := rotationToTestValidators(t, realioApp, ctx)
	oldVal, newVal := rotations[0].OldVal, rotations[0].NewVal
	coin := realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, oldVal)
	msMsgServer := multistakingkeeper.NewMsgServerImpl(realioApp.MultiStakingKeeper)

	dust := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	stake := sdk.NewCoin(coin, math.NewInt(1))
	require.NoError(t, realioApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(stake)))
	require.NoError(t, realioApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, dust, sdk.NewCoins(stake)))
	_, err := msMsgServer.Delegate(ctx, &stakingtypes.MsgDelegate{DelegatorAddress: dust.String(), ValidatorAddress: oldVal.String(), Amount: stake})
	require.NoError(t, err)

	// Slash the outgoing validator: every share is now worth about half a token.
	oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, oldVal)
	require.NoError(t, err)
	consAddrBz, err := oldValidator.GetConsAddr()
	require.NoError(t, err)
	_, err = realioApp.StakingKeeper.Slash(ctx, sdk.ConsAddress(consAddrBz), ctx.BlockHeight(),
		oldValidator.ConsensusPower(sdk.DefaultPowerReduction), math.LegacyNewDecWithPrec(5, 1))
	require.NoError(t, err)

	oldValidator, err = realioApp.StakingKeeper.GetValidator(ctx, oldVal)
	require.NoError(t, err)
	dustDelBefore, err := realioApp.StakingKeeper.GetDelegation(ctx, dust, oldVal)
	require.NoError(t, err, "premise: the dust delegation still exists")
	require.True(t, dustDelBefore.Shares.IsPositive())
	require.True(t, oldValidator.TokensFromShares(dustDelBefore.Shares).TruncateInt().IsZero(),
		"premise: its shares are worth less than one token")
	dustLockBefore, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(dust.String(), oldVal.String()))
	require.True(t, found)
	newValBefore, err := realioApp.StakingKeeper.GetValidator(ctx, newVal)
	require.NoError(t, err)

	var logs bytes.Buffer
	logCtx := ctx.WithLogger(log.NewLogger(&logs))
	require.NotPanics(t, func() {
		v8.RotateValidators(logCtx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper)
	})

	require.Contains(t, logs.String(), "skipping dust delegation")
	require.Contains(t, logs.String(), dust.String())
	require.Equal(t, 1, strings.Count(logs.String(), "skipping dust delegation"), "only the dust delegation should be skipped")

	// The dust delegation is untouched...
	dustDelAfter, err := realioApp.StakingKeeper.GetDelegation(ctx, dust, oldVal)
	require.NoError(t, err)
	require.Equal(t, dustDelBefore.Shares.String(), dustDelAfter.Shares.String())
	dustLockAfter, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx, multistakingtypes.MultiStakingLockID(dust.String(), oldVal.String()))
	require.True(t, found)
	require.Equal(t, dustLockBefore.LockedCoin.Amount.String(), dustLockAfter.LockedCoin.Amount.String(), "the lock must not have been moved")
	_, err = realioApp.StakingKeeper.GetDelegation(ctx, dust, newVal)
	require.Error(t, err)

	// ...and everyone else on that validator moved, the operator included. With
	// the slash the tokens-per-share rate is no longer 1, so moving a position
	// converts tokens to shares with truncation and each delegator is left with
	// a fraction of a token's worth of shares behind; nothing bigger may remain.
	oldValidatorAfter, err := realioApp.StakingKeeper.GetValidator(ctx, oldVal)
	require.NoError(t, err)
	remaining, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, oldVal)
	require.NoError(t, err)
	for _, d := range remaining {
		require.Truef(t, oldValidatorAfter.TokensFromShares(d.Shares).TruncateInt().IsZero(),
			"%s still has %s shares (worth at least one token) on %s", d.DelegatorAddress, d.Shares, oldVal)
	}
	require.True(t, oldValidatorAfter.Tokens.LT(math.NewInt(1000)),
		"what is left on the outgoing validator should be rounding residue, got %s", oldValidatorAfter.Tokens)
	require.True(t, oldValidatorAfter.Jailed)
	newValAfter, err := realioApp.StakingKeeper.GetValidator(ctx, newVal)
	require.NoError(t, err)
	require.True(t, newValAfter.Tokens.GT(newValBefore.Tokens), "the replacement should have received everyone else's stake")
}
