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
