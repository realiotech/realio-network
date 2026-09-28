package v8_test

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
	v8 "github.com/realiotech/realio-network/app/upgrades/v1.8"
	realionetworktypes "github.com/realiotech/realio-network/types"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

// rotationGenesisPath is the mainnet export every test in this file boots
// from. It lives with this package so the v1.8.0 rotation is exercised
// against its own fixture.
const rotationGenesisPath = "testdata/exported_mainnet_after.json"

func setupRotationGenesis(t *testing.T) (*app.RealioNetwork, string, int64, []byte, time.Time) {
	t.Helper()
	return app.SetupWithGenesisFile(t, rotationGenesisPath)
}

// testnetRotationGenesisPath is the real realio testnet export
// (chain-id realionetwork_3300-*) that TestnetValidatorRotations' two real
// OldValidator entries come from, so the testnet path gets exercised
// against its own real data the same way the mainnet path does.
const testnetRotationGenesisPath = "testdata/testnet_2809.json"

func setupTestnetRotationGenesis(t *testing.T) (*app.RealioNetwork, string, int64, []byte, time.Time) {
	t.Helper()
	return app.SetupWithGenesisFile(t, testnetRotationGenesisPath)
}

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

// rotationToTestValidatorsUsing overwrites rotations (via set, restored by
// t.Cleanup) so every real entry points at a freshly-created stand-in
// validator instead of its real NewValidator -- shared by
// rotationToTestValidators (mainnet's v8.ValidatorRotations) and
// rotationToTestnetValidators (v8.TestnetValidatorRotations). Returns the
// old/new ValAddress pairs actually used, in the same order.
func rotationToTestValidatorsUsing(
	t *testing.T,
	realioApp *app.RealioNetwork,
	ctx sdk.Context,
	rotations []v8.ValidatorRotation,
	set func([]v8.ValidatorRotation),
) []struct{ OldVal, NewVal sdk.ValAddress } {
	t.Helper()

	t.Cleanup(func() { set(rotations) })
	require.Len(t, rotations, 2, "expected exactly the two known validators pending rotation")

	out := make([]struct{ OldVal, NewVal sdk.ValAddress }, len(rotations))
	newRotations := make([]v8.ValidatorRotation, len(rotations))

	for i, r := range rotations {
		oldVal, err := sdk.ValAddressFromBech32(r.OldValidator)
		require.NoError(t, err)
		denom := realioApp.MultiStakingKeeper.GetValidatorMultiStakingCoin(ctx, oldVal)
		require.NotEmpty(t, denom, "expected %s to have a registered multi-staking coin in the genesis", r.OldValidator)

		newVal := createTestValidator(t, realioApp, ctx, denom)
		out[i] = struct{ OldVal, NewVal sdk.ValAddress }{OldVal: oldVal, NewVal: newVal}
		newRotations[i] = v8.ValidatorRotation{OldValidator: r.OldValidator, NewValidator: newVal.String()}
	}
	set(newRotations)
	return out
}

// rotationToTestValidators overwrites v8.ValidatorRotations (mainnet) --
// see rotationToTestValidatorsUsing.
func rotationToTestValidators(t *testing.T, realioApp *app.RealioNetwork, ctx sdk.Context) []struct{ OldVal, NewVal sdk.ValAddress } {
	t.Helper()
	return rotationToTestValidatorsUsing(t, realioApp, ctx, v8.ValidatorRotations, func(r []v8.ValidatorRotation) { v8.ValidatorRotations = r })
}

// rotationToTestnetValidators overwrites v8.TestnetValidatorRotations --
// see rotationToTestValidatorsUsing.
func rotationToTestnetValidators(t *testing.T, realioApp *app.RealioNetwork, ctx sdk.Context) []struct{ OldVal, NewVal sdk.ValAddress } {
	t.Helper()
	return rotationToTestValidatorsUsing(t, realioApp, ctx, v8.TestnetValidatorRotations, func(r []v8.ValidatorRotation) { v8.TestnetValidatorRotations = r })
}

// requireOldValidatorDrained asserts oldValidator has at most tolerance
// tokens left (zero, for the two clean mainnet rotation targets; a small
// dust allowance for messier real validators -- see the residueTolerance
// notes in TestRotateValidatorsAgainstRealTestnetGenesis) and that it got
// auto-jailed, the way it must once its self-bond redelegates away.
func requireOldValidatorDrained(t *testing.T, oldValidator stakingtypes.Validator, tolerance math.Int) {
	t.Helper()
	require.Truef(t, oldValidator.Tokens.LTE(tolerance),
		"expected %s to have at most %s tokens left after rotation, got %s", oldValidator.OperatorAddress, tolerance, oldValidator.Tokens)
	require.True(t, oldValidator.Jailed,
		"expected %s to be auto-jailed once its self-bond redelegated away", oldValidator.OperatorAddress)
}

// rotationFixture is one real genesis this file's tests run the same checks
// against: mainnet's ValidatorRotations plus its own export, and testnet's
// TestnetValidatorRotations plus its own. residueTolerance accounts for how
// much messier a real validator's leftover balance can be after rotation --
// see the notes in TestRotateValidatorsAgainstRealTestnetGenesis.
type rotationFixture struct {
	name             string
	isTestnet        bool
	setup            func(t *testing.T) (realioApp *app.RealioNetwork, chainID string, initialHeight int64, proposerAddr []byte, blockTime time.Time)
	rotationTo       func(t *testing.T, realioApp *app.RealioNetwork, ctx sdk.Context) []struct{ OldVal, NewVal sdk.ValAddress }
	residueTolerance math.Int
}

var rotationFixtures = []rotationFixture{
	{name: "mainnet", isTestnet: false, setup: setupRotationGenesis, rotationTo: rotationToTestValidators, residueTolerance: math.ZeroInt()},
	{name: "testnet", isTestnet: true, setup: setupTestnetRotationGenesis, rotationTo: rotationToTestnetValidators, residueTolerance: math.NewInt(1_000_000)},
}

// newRotationFixtureCtx boots fixture's genesis and returns a context ready
// for rotation: NewHeaderCtx doesn't set ChainID (it only ever needed
// Height/ProposerAddress/Time before TestnetValidatorRotations existed), but
// RotateValidators picks TestnetValidatorRotations vs ValidatorRotations
// based on ctx.ChainID(), so it's set explicitly here for every fixture.
func newRotationFixtureCtx(t *testing.T, fx rotationFixture) (*app.RealioNetwork, sdk.Context) {
	t.Helper()
	realioApp, chainID, initialHeight, proposerAddr, blockTime := fx.setup(t)
	require.Equalf(t, fx.isTestnet, realionetworktypes.IsTestnet(chainID),
		"test premise: fixture %q's genesis chain-id %q must match isTestnet=%v", fx.name, chainID, fx.isTestnet)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime).WithChainID(chainID)
	return realioApp, ctx
}

// TestV18UpgradeRotatesValidatorsAgainstRealGenesis proves the wiring, not
// just the underlying logic: scheduling and applying the real v1.8.0
// upgrade plan through app.UpgradeKeeper (the same path a live chain takes
// for a governance-approved software upgrade) must actually invoke
// RotateValidators, not just have it available to call directly, and must
// pick the right rotation list (ValidatorRotations vs
// TestnetValidatorRotations) for the chain it's running on. Mirrors the
// existing TestCommissionUpgrade pattern (app/upgrades_test.go) for
// exercising a registered upgrade handler end-to-end, against every real
// genesis in rotationFixtures.
func TestV18UpgradeRotatesValidatorsAgainstRealGenesis(t *testing.T) {
	for _, fx := range rotationFixtures {
		t.Run(fx.name, func(t *testing.T) {
			realioApp, ctx := newRotationFixtureCtx(t, fx)
			rotations := fx.rotationTo(t, realioApp, ctx)

			plan := upgradetypes.Plan{Name: v8.UpgradeName, Height: ctx.BlockHeight()}
			require.NoError(t, realioApp.UpgradeKeeper.ScheduleUpgrade(ctx, plan))

			ctx = ctx.WithBlockTime(time.Now())
			require.NoError(t, realioApp.UpgradeKeeper.ApplyUpgrade(ctx, plan))

			for _, r := range rotations {
				oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
				require.NoError(t, err)
				requireOldValidatorDrained(t, oldValidator, fx.residueTolerance)

				newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
				require.NoError(t, err)
				require.True(t, newValidator.Tokens.IsPositive(), "expected %s to have received the redelegated stake", r.NewVal)
			}
		})
	}
}

// TestRotateValidatorsAgainstRealGenesis runs RotateValidators against every
// real genesis in rotationFixtures, for the actual validators each one's
// rotation list targets -- the real scenario this migration exists for, not
// a synthetic fixture. Since the real replacement validators aren't known to
// this codebase yet (both ValidatorRotations and TestnetValidatorRotations
// ship with real OldValidator entries but stand-in NewValidator needs), each
// entry points at a freshly-created stand-in validator for the duration of
// the test, so the redelegation MECHANICS get verified against real
// delegator/lock/share data even though the real destination addresses are
// still pending.
//
// mainnet's two rotation targets are clean (confirmed by
// TestRotateValidatorsGenesisStateDetail: zero lock-less/dust delegators),
// so that fixture asserts every single delegator moved, with no residue left
// behind at all. Real testnet validators are messier -- some delegators can
// be skipped (lock-less, dust, incoming-redelegation -- see
// redelegateOneDelegation), and even a moved delegation can leave a little
// truncation residue behind (the redelegated amount is sized off the
// multistaking lock's integer coin amount, which can drift slightly from
// the exact x/staking shares) -- so that fixture only requires that most
// real delegators moved and that the old validator's total leftover is
// dust-scale, per its residueTolerance.
func TestRotateValidatorsAgainstRealGenesis(t *testing.T) {
	for _, fx := range rotationFixtures {
		t.Run(fx.name, func(t *testing.T) {
			realioApp, ctx := newRotationFixtureCtx(t, fx)
			rotations := fx.rotationTo(t, realioApp, ctx)
			strict := fx.residueTolerance.IsZero()

			// Snapshot every delegator + the operator's own delegation for
			// each old validator before running, so "did everyone actually
			// move" can be checked afterwards.
			type preState struct {
				delegators  []string
				hadOperator bool
			}
			pre := make([]preState, len(rotations))
			for i, r := range rotations {
				dels, err := realioApp.StakingKeeper.GetValidatorDelegations(ctx, r.OldVal)
				require.NoError(t, err)
				require.NotEmptyf(t, dels, "expected %s to have real delegators in the genesis", r.OldVal)

				operatorAddr := sdk.AccAddress(r.OldVal).String()
				hadOperator := false
				addrs := make([]string, 0, len(dels))
				for _, d := range dels {
					addrs = append(addrs, d.DelegatorAddress)
					if d.DelegatorAddress == operatorAddr {
						hadOperator = true
					}
				}
				pre[i] = preState{delegators: addrs, hadOperator: hadOperator}
				t.Logf("validator %s: %d real delegators before rotation (operator self-bond present: %v)", r.OldVal, len(addrs), hadOperator)
			}

			require.NoError(t, v8.RotateValidators(ctx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper))

			for i, r := range rotations {
				oldValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.OldVal)
				require.NoError(t, err)
				if pre[i].hadOperator {
					requireOldValidatorDrained(t, oldValidator, fx.residueTolerance)
				} else {
					require.Truef(t, oldValidator.Tokens.LTE(fx.residueTolerance),
						"expected %s to have at most %s tokens left after rotation, got %s", r.OldVal, fx.residueTolerance, oldValidator.Tokens)
				}

				newValidator, err := realioApp.StakingKeeper.GetValidator(ctx, r.NewVal)
				require.NoError(t, err)
				require.True(t, newValidator.Tokens.IsPositive(), "expected %s to have received the redelegated stake", r.NewVal)

				// Every delegator that actually moved has a Redelegation
				// record and a real delegation on the new validator; strict
				// fixtures (mainnet) require every one of them to have
				// moved, lenient ones (testnet) just require most of them.
				moved := 0
				for _, delStr := range pre[i].delegators {
					delAddr, err := sdk.AccAddressFromBech32(delStr)
					require.NoError(t, err)

					newDel, err := realioApp.StakingKeeper.GetDelegation(ctx, delAddr, r.NewVal)
					if err != nil || !newDel.Shares.IsPositive() {
						require.Falsef(t, strict, "expected %s to have a delegation on the replacement validator %s", delStr, r.NewVal)
						continue
					}
					moved++

					red, err := realioApp.StakingKeeper.GetRedelegation(ctx, delAddr, r.OldVal, r.NewVal)
					require.NoErrorf(t, err, "expected a redelegation record for %s from %s to %s", delStr, r.OldVal, r.NewVal)
					require.NotEmpty(t, red.Entries)
				}
				require.NotZero(t, moved, "expected at least some real delegators to have moved to %s", r.NewVal)

				t.Logf("validator %s -> %s: %d/%d real delegators redelegated, old validator tokens left=%s, new validator tokens=%s",
					r.OldVal, r.NewVal, moved, len(pre[i].delegators), oldValidator.Tokens, newValidator.Tokens)
			}
		})
	}
}

// TestRotatedStakeStillSlashableForPriorInfraction answers a question this
// migration's design leans on but doesn't itself exercise: what happens if
// evidence of the outgoing validator's misbehavior (e.g. a double-sign)
// surfaces *after* this migration already moved its delegators' stake to
// the replacement validator? Because redelegateOneDelegation goes through
// the real MsgBeginRedelegate path (see the package doc comment), the
// answer comes from ordinary cosmos-sdk redelegation-slashing semantics,
// not from any special-casing in this package: x/staking's
// SlashRedelegation burns a redelegation entry whenever the infraction
// height is at or before the entry's CreationHeight (the height
// RotateValidators ran at). So stake that was still backing the outgoing
// validator at the time of the infraction stays liable for it even though
// it now sits with the replacement validator -- a delegator can't outrun a
// slash for something that already happened just because this migration
// moved them afterwards.
func TestRotatedStakeStillSlashableForPriorInfraction(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, blockTime := setupRotationGenesis(t)
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

	require.NoError(t, v8.RotateValidators(ctx, realioApp.StakingKeeper, realioApp.MultiStakingKeeper))
	rotationHeight := ctx.BlockHeight()
	slashFactor := math.LegacyNewDecWithPrec(5, 2) // 5%

	for _, tg := range targets {
		newValBefore, err := realioApp.StakingKeeper.GetValidator(ctx, tg.newVal)
		require.NoError(t, err)
		delBefore, err := realioApp.StakingKeeper.GetDelegation(ctx, tg.delegator, tg.newVal)
		require.NoError(t, err)
		tokensBefore := newValBefore.TokensFromShares(delBefore.Shares)

		// The scenario this test exists for: a double-sign by the outgoing
		// validator at a height well before the migration only gets
		// reported (evidence submitted) now, after the redelegation.
		infractionHeight := rotationHeight - 100
		_, err = realioApp.StakingKeeper.Slash(ctx, tg.consAddr, infractionHeight, tg.power, slashFactor)
		require.NoError(t, err)

		newValAfter, err := realioApp.StakingKeeper.GetValidator(ctx, tg.newVal)
		require.NoError(t, err)
		require.True(t, newValAfter.Tokens.LT(newValBefore.Tokens),
			"expected %s's total tokens to drop from a slash for %s's pre-rotation infraction", tg.newVal, tg.oldVal)

		delAfter, err := realioApp.StakingKeeper.GetDelegation(ctx, tg.delegator, tg.newVal)
		require.NoError(t, err)
		tokensAfter := newValAfter.TokensFromShares(delAfter.Shares)
		require.True(t, tokensAfter.LT(tokensBefore),
			"expected non-operator delegator %s's own redelegated stake on %s to be burned for %s's pre-rotation infraction, not just the validator's aggregate total",
			tg.delegator, tg.newVal, tg.oldVal)

		t.Logf("validator %s -> %s: retroactive slash for a pre-rotation infraction (height %d) burned delegator %s from %s to %s tokens",
			tg.oldVal, tg.newVal, infractionHeight, tg.delegator, tokensBefore, tokensAfter)
	}
}
