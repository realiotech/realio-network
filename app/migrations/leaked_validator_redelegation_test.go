package migrations_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"

	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/app/migrations"
	v8 "github.com/realiotech/realio-network/app/upgrades/v1.8"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

// createTestValidator creates a brand-new validator self-bonded in denom
// through the real multistaking CreateValidator message (so AfterValidatorCreated
// and every other hook a live chain would run actually fires), to stand in
// for a real replacement validator whose address isn't known to this
// codebase yet.
func createTestValidator(t *testing.T, realioApp *app.RealioNetwork, ctx sdk.Context, denom string) sdk.ValAddress {
	t.Helper()

	priv := ed25519.GenPrivKey()
	valAddr := sdk.ValAddress(priv.PubKey().Address())

	selfBond := math.NewInt(1_000_000_000_000)
	coins := sdk.NewCoins(sdk.NewCoin(denom, selfBond))
	require.NoError(t, realioApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, realioApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sdk.AccAddress(valAddr), coins))

	createMsg, err := stakingtypes.NewMsgCreateValidator(
		valAddr.String(),
		priv.PubKey(),
		sdk.NewCoin(denom, selfBond),
		stakingtypes.Description{Moniker: "replacement-" + denom},
		stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2), math.LegacyZeroDec()),
		math.OneInt(),
	)
	require.NoError(t, err)

	msMsgServer := multistakingkeeper.NewMsgServerImpl(realioApp.MultiStakingKeeper)
	_, err = msMsgServer.CreateValidator(ctx, createMsg)
	require.NoError(t, err)

	return valAddr
}

// rotationToTestValidators overwrites migrations.LeakedValidatorRedelegations
// (restored via t.Cleanup) so every real leaked-validator entry points at a
// freshly-created stand-in validator instead of its blank/real NewValidator
// -- the real replacement addresses aren't known to this codebase yet.
// Returns the old/new ValAddress pairs actually used, in the same order.
func rotationToTestValidators(t *testing.T, realioApp *app.RealioNetwork, ctx sdk.Context) []struct{ OldVal, NewVal sdk.ValAddress } {
	t.Helper()

	origRotations := migrations.LeakedValidatorRedelegations
	t.Cleanup(func() { migrations.LeakedValidatorRedelegations = origRotations })
	require.Len(t, origRotations, 2, "expected exactly the two known real leaked validators")

	out := make([]struct{ OldVal, NewVal sdk.ValAddress }, len(origRotations))
	newRotations := make([]struct {
		OldValidator string
		NewValidator string
	}, len(origRotations))

	for i, r := range origRotations {
		oldVal, err := sdk.ValAddressFromBech32(r.OldValidator)
		require.NoError(t, err)
		denom := realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, oldVal)
		require.NotEmpty(t, denom, "expected %s to have a registered multi-staking coin in the real genesis", r.OldValidator)

		newVal := createTestValidator(t, realioApp, ctx, denom)
		out[i] = struct{ OldVal, NewVal sdk.ValAddress }{OldVal: oldVal, NewVal: newVal}
		newRotations[i] = struct {
			OldValidator string
			NewValidator string
		}{OldValidator: r.OldValidator, NewValidator: newVal.String()}
	}
	migrations.LeakedValidatorRedelegations = newRotations
	return out
}

// TestV18UpgradeRedelegatesLeakedValidatorsAgainstRealGenesis proves the
// wiring, not just the underlying logic: scheduling and applying the real
// v1.8.0 upgrade plan through app.UpgradeKeeper (the same path a live chain
// takes for a governance-approved software upgrade) must actually invoke
// RedelegateLeakedValidators, not just have it available to call directly.
// Mirrors the existing TestCommissionUpgrade pattern (app/upgrades_test.go)
// for exercising a registered upgrade handler end-to-end.
func TestV18UpgradeRedelegatesLeakedValidatorsAgainstRealGenesis(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := app.SetupWithRealGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	rotations := rotationToTestValidators(t, realioApp, ctx)

	plan := upgradetypes.Plan{Name: v8.UpgradeName, Height: ctx.BlockHeight()}
	require.NoError(t, realioApp.UpgradeKeeper.ScheduleUpgrade(ctx, plan))

	ctx = ctx.WithBlockTime(time.Now())
	require.NoError(t, realioApp.UpgradeKeeper.ApplyUpgrade(ctx, plan))

	for _, r := range rotations {
		oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
		require.NoError(t, err)
		require.True(t, oldValidator.Tokens.IsZero(), "expected %s to have zero tokens after the v1.8.0 upgrade ran", r.OldVal)
		require.True(t, oldValidator.Jailed, "expected %s to be auto-jailed after its self-bond redelegated away", r.OldVal)

		newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
		require.NoError(t, err)
		require.True(t, newValidator.Tokens.IsPositive(), "expected %s to have received the redelegated stake", r.NewVal)
	}
}

// TestRedelegateLeakedValidatorsAgainstRealGenesis runs
// RedelegateLeakedValidators against the real pre-incident genesis for the
// two real leaked validators — the actual scenario this migration exists
// for, not a synthetic fixture. Since the real replacement validators
// aren't known to this codebase yet (LeakedValidatorRedelegations ships
// with NewValidator left blank), this test points both entries at
// freshly-created stand-in validators for the duration of the test, so the
// redelegation MECHANICS get verified against real delegator/lock/share
// data even though the real destination addresses are still pending.
func TestRedelegateLeakedValidatorsAgainstRealGenesis(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := app.SetupWithRealGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	rotations := rotationToTestValidators(t, realioApp, ctx)

	// Snapshot every delegator + the operator's own delegation for each old
	// validator before running, so "did everyone actually move" can be
	// checked precisely afterwards.
	type preState struct {
		delegators   []string
		operatorAddr string
		hadOperator  bool
	}
	pre := make([]preState, len(rotations))
	for i, r := range rotations {
		dels, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, r.OldVal)
		require.NoError(t, err)
		require.NotEmpty(t, dels, "expected %s to have real delegators in the genesis", r.OldVal)

		operatorAddr := sdk.AccAddress(r.OldVal).String()
		hadOperator := false
		addrs := make([]string, 0, len(dels))
		for _, d := range dels {
			addrs = append(addrs, d.DelegatorAddress)
			if d.DelegatorAddress == operatorAddr {
				hadOperator = true
			}
		}
		pre[i] = preState{delegators: addrs, operatorAddr: operatorAddr, hadOperator: hadOperator}
		t.Logf("validator %s: %d real delegators before redelegation (operator self-bond present: %v)", r.OldVal, len(addrs), hadOperator)
	}

	migrations.RedelegateLeakedValidators(realioApp.MigrationKeepers(), ctx)

	for i, r := range rotations {
		// Old validator: every delegation gone (redelegated away in full),
		// tokens at zero.
		remaining, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, r.OldVal)
		require.NoError(t, err)
		require.Empty(t, remaining, "expected every delegation to have moved off %s", r.OldVal)

		oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
		require.NoError(t, err)
		require.True(t, oldValidator.Tokens.IsZero(), "expected %s to have zero tokens left after redelegation", r.OldVal)

		if pre[i].hadOperator {
			require.True(t, oldValidator.Jailed,
				"expected %s to be auto-jailed once its own self-bond (included in this redelegation) dropped below MinSelfDelegation", r.OldVal)
		}

		// New validator: received it all.
		newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
		require.NoError(t, err)
		require.True(t, newValidator.Tokens.IsPositive(), "expected %s to have received the redelegated stake", r.NewVal)

		// Every original delegator now has a Redelegation record and a real
		// delegation on the new validator.
		for _, delStr := range pre[i].delegators {
			delAddr, err := sdk.AccAddressFromBech32(delStr)
			require.NoError(t, err)

			red, err := realioApp.StakingKeeper.GetRedelegation(ctx, delAddr, r.OldVal, r.NewVal)
			require.NoErrorf(t, err, "expected a redelegation record for %s from %s to %s", delStr, r.OldVal, r.NewVal)
			require.NotEmpty(t, red.Entries)

			newDel, err := realioApp.StakingKeeper.GetDelegation(ctx, delAddr, r.NewVal)
			require.NoErrorf(t, err, "expected %s to have a delegation on the replacement validator %s", delStr, r.NewVal)
			require.True(t, newDel.Shares.IsPositive())
		}

		t.Logf("validator %s -> %s: %d real delegators redelegated, old validator tokens=0 jailed=%v, new validator tokens=%s",
			r.OldVal, r.NewVal, len(pre[i].delegators), oldValidator.Jailed, newValidator.Tokens.String())
	}
}

// TestRedelegatedStakeStillSlashableForPreMigrationInfraction answers a
// question this migration's design leans on but doesn't itself exercise:
// what happens if evidence of the leaked validator's misbehavior (e.g. a
// double-sign) surfaces *after* this migration already moved its delegators'
// stake to the replacement validator? Because redelegateOneDelegation goes
// through the real MsgBeginRedelegate path (see the package doc comment),
// the answer comes from ordinary cosmos-sdk redelegation-slashing semantics,
// not from any special-casing in this package: x/staking's SlashRedelegation
// burns a redelegation entry whenever the infraction height is at or before
// the entry's CreationHeight (the height RedelegateLeakedValidators ran at).
// So stake that was still backing the leaked validator at the time of the
// infraction stays liable for it even though it now sits with the
// replacement validator -- a delegator can't outrun a slash for something
// that already happened just because this migration moved them afterwards.
func TestRedelegatedStakeStillSlashableForPreMigrationInfraction(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := app.SetupWithRealGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

	rotations := rotationToTestValidators(t, realioApp, ctx)

	// Snapshot each old validator's consensus address + power *before* the
	// migration drains it -- this is what a real double-sign evidence
	// submission would reference for an infraction that predates the move.
	// Also pick one real, non-operator delegator per old validator so the
	// assertions below can check a specific innocent delegator's stake, not
	// just the replacement validator's aggregate total.
	type target struct {
		oldVal, newVal sdk.ValAddress
		consAddr       sdk.ConsAddress
		power          int64
		delegator      sdk.AccAddress
	}
	targets := make([]target, 0, len(rotations))
	for _, r := range rotations {
		v, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
		require.NoError(t, err)
		consAddrBz, err := v.GetConsAddr()
		require.NoError(t, err)

		dels, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, r.OldVal)
		require.NoError(t, err)
		operatorAddr := sdk.AccAddress(r.OldVal).String()
		var delegator sdk.AccAddress
		found := false
		for _, d := range dels {
			if d.DelegatorAddress == operatorAddr {
				continue
			}
			delegator, err = sdk.AccAddressFromBech32(d.DelegatorAddress)
			require.NoError(t, err)
			found = true
			break
		}
		require.True(t, found, "expected %s to have at least one non-operator delegator in the real genesis", r.OldVal)

		targets = append(targets, target{
			oldVal:    r.OldVal,
			newVal:    r.NewVal,
			consAddr:  sdk.ConsAddress(consAddrBz),
			power:     v.ConsensusPower(sdk.DefaultPowerReduction),
			delegator: delegator,
		})
	}

	migrations.RedelegateLeakedValidators(realioApp.MigrationKeepers(), ctx)
	redelegationHeight := ctx.BlockHeight()
	slashFactor := math.LegacyNewDecWithPrec(5, 2) // 5%

	for _, tg := range targets {
		newValBefore, err := realioApp.StakingKeeper.GetValidator(ctx, tg.newVal)
		require.NoError(t, err)
		delBefore, err := realioApp.StakingKeeper.GetDelegation(ctx, tg.delegator, tg.newVal)
		require.NoError(t, err)
		tokensBefore := newValBefore.TokensFromShares(delBefore.Shares)

		// The scenario this test exists for: a double-sign by the leaked
		// validator at a height well before the migration only gets
		// reported (evidence submitted) now, after the redelegation.
		infractionHeight := redelegationHeight - 100
		_, err = realioApp.StakingKeeper.Slash(ctx, tg.consAddr, infractionHeight, tg.power, slashFactor)
		require.NoError(t, err)

		newValAfter, err := realioApp.StakingKeeper.GetValidator(ctx, tg.newVal)
		require.NoError(t, err)
		require.True(t, newValAfter.Tokens.LT(newValBefore.Tokens),
			"expected %s's total tokens to drop from a slash for %s's pre-migration infraction", tg.newVal, tg.oldVal)

		delAfter, err := realioApp.StakingKeeper.GetDelegation(ctx, tg.delegator, tg.newVal)
		require.NoError(t, err)
		tokensAfter := newValAfter.TokensFromShares(delAfter.Shares)
		require.True(t, tokensAfter.LT(tokensBefore),
			"expected non-operator delegator %s's own redelegated stake on %s to be burned for %s's pre-migration infraction, not just the validator's aggregate total",
			tg.delegator, tg.newVal, tg.oldVal)

		t.Logf("validator %s -> %s: retroactive slash for a pre-redelegation infraction (height %d) burned delegator %s from %s to %s tokens",
			tg.oldVal, tg.newVal, infractionHeight, tg.delegator, tokensBefore, tokensAfter)
	}
}
