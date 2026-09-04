package migrations_test

import (
	"encoding/base64"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/app/migrations"
	"github.com/realiotech/realio-network/testutil"
)

// TestRotateValidator exercises RotateValidators end-to-end against a real
// app: a validator with two delegators (self + one other, mirroring the
// "72 other delegators" shape found in the real leaked-validator genesis),
// distribution history, and multistaking locks, migrated to a brand new
// operator address + consensus key. Every piece of state tied to the old
// identity must move — nothing about delegator addresses, shares, or coin
// amounts should change.
func TestRotateValidator(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1}).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	sk := realio.StakingKeeper
	dk := realio.DistrKeeper
	msk := realio.MultiStakingKeeper

	validators, err := sk.GetValidators(ctx, 10)
	require.NoError(t, err)
	require.Len(t, validators, 1)
	oldValidator := validators[0]
	oldValAddr, err := sdk.ValAddressFromBech32(oldValidator.OperatorAddress)
	require.NoError(t, err)

	// second delegator, distinct from the validator's self-delegation. Set
	// directly rather than via Delegate() — this test targets the
	// migration logic itself, not the bonded-pool bookkeeping Delegate
	// would otherwise require funding.
	otherDelegator := testutil.GenAddress()
	require.NoError(t, sk.SetDelegation(ctx, stakingtypes.NewDelegation(
		otherDelegator.String(), oldValidator.OperatorAddress, math.LegacyNewDec(1_000_000),
	)))

	otherLockID := multistakingtypes.MultiStakingLockID(otherDelegator.String(), oldValAddr.String())
	msk.SetMultiStakingLock(ctx, multistakingtypes.NewMultiStakingLock(otherLockID, app.MultiStakingCoinA))

	// distribution state tied to the old validator identity
	require.NoError(t, dk.SetValidatorCurrentRewards(ctx, oldValAddr, distrtypes.ValidatorCurrentRewards{
		Rewards: sdk.DecCoins{sdk.NewDecCoin("ario", math.NewInt(100))}, Period: 1,
	}))
	require.NoError(t, dk.SetValidatorAccumulatedCommission(ctx, oldValAddr, distrtypes.ValidatorAccumulatedCommission{
		Commission: sdk.DecCoins{sdk.NewDecCoin("ario", math.NewInt(50))},
	}))
	require.NoError(t, dk.SetValidatorOutstandingRewards(ctx, oldValAddr, distrtypes.ValidatorOutstandingRewards{
		Rewards: sdk.DecCoins{sdk.NewDecCoin("ario", math.NewInt(150))},
	}))
	require.NoError(t, dk.SetValidatorHistoricalRewards(ctx, oldValAddr, 0, distrtypes.ValidatorHistoricalRewards{
		CumulativeRewardRatio: sdk.DecCoins{sdk.NewDecCoin("ario", math.NewInt(1))}, ReferenceCount: 1,
	}))
	require.NoError(t, dk.SetDelegatorStartingInfo(ctx, oldValAddr, otherDelegator, distrtypes.DelegatorStartingInfo{
		PreviousPeriod: 0, Stake: math.LegacyNewDec(1_000_000), Height: 1,
	}))
	require.NoError(t, dk.SetValidatorSlashEvent(ctx, oldValAddr, 1, 0, distrtypes.ValidatorSlashEvent{
		ValidatorPeriod: 0, Fraction: math.LegacyNewDecWithPrec(5, 2),
	}))

	oldLocks := lockCount(msk, ctx, oldValAddr.String())
	require.Equal(t, 2, oldLocks, "expected self + other delegator lock before rotation")

	// prepare a brand new operator + consensus identity
	newOperatorKey := testutil.GenAddress()
	newValAddr := sdk.ValAddress(newOperatorKey)
	newConsPriv := ed25519.GenPrivKey()
	newConsPubKeyAny, err := codectypes.NewAnyWithValue(newConsPriv.PubKey())
	require.NoError(t, err)

	oldConsPubKey := oldValidator.ConsensusPubkey

	origRotations, origHeight := migrations.ValidatorRotations, migrations.ValidatorRotationHeight
	t.Cleanup(func() { migrations.ValidatorRotations, migrations.ValidatorRotationHeight = origRotations, origHeight })
	migrations.ValidatorRotationHeight = ctx.BlockHeight()
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      newValAddr.String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	// drives the exact same path production code takes: ScheduleValidatorRotation
	// (height-gated) capturing the pending zero-update for EndBlocker to drain.
	realio.ScheduleValidatorRotation(ctx)

	// old identity's economic state (delegations, distribution,
	// multistaking) has moved off, but the plain Validator record itself is
	// deliberately kept resolvable (see TestBeginBlockerRotationOrdering
	// for why) — it's just gone from consensus-power accounting.
	oldGhost, err := sk.GetValidator(ctx, oldValAddr)
	require.NoError(t, err, "old validator record is kept in place indefinitely, not deleted")
	require.Truef(t, oldGhost.Tokens.IsZero(),
		"old validator's Tokens must be zeroed — otherwise a genesis export sums it alongside the new "+
			"validator's Tokens against a bonded pool that only actually backs one of them, and InitGenesis panics")
	require.True(t, oldGhost.DelegatorShares.IsPositive(),
		"DelegatorShares must be left untouched (Tokens=0 + positive shares makes InvalidExRate() true, "+
			"which is what blocks new delegations into the old identity)")
	require.True(t, oldGhost.InvalidExRate(), "sanity: the zeroed-Tokens ghost must read as an invalid exchange rate")
	require.Equal(t, stakingtypes.Bonded, oldGhost.Status, "status is deliberately left untouched")
	require.Empty(t, lockCount(msk, ctx, oldValAddr.String()))
	require.Empty(t, msk.GetValidatorMultiStakingCoin(ctx, oldValAddr))
	oldCurRewards, err := dk.GetValidatorCurrentRewards(ctx, oldValAddr)
	require.NoError(t, err)
	require.True(t, oldCurRewards.Rewards.IsZero(), "old validator's current rewards must be cleared, not just readable-as-zero-value")
	oldCommission, err := dk.GetValidatorAccumulatedCommission(ctx, oldValAddr)
	require.NoError(t, err)
	require.True(t, oldCommission.Commission.IsZero())
	var oldSlashEventsRemain bool
	dk.IterateValidatorSlashEvents(ctx, func(val sdk.ValAddress, _ uint64, _ distrtypes.ValidatorSlashEvent) bool {
		if val.Equals(oldValAddr) {
			oldSlashEventsRemain = true
			return true
		}
		return false
	})
	require.False(t, oldSlashEventsRemain, "old validator's slash-event history must be cleared, not left orphaned under the old address")

	// new identity carries everything over unchanged except operator/pubkey
	newValidator, err := sk.GetValidator(ctx, newValAddr)
	require.NoError(t, err)
	require.Equal(t, newValAddr.String(), newValidator.OperatorAddress)
	require.True(t, newConsPubKeyAny.Equal(newValidator.ConsensusPubkey))
	require.False(t, oldConsPubKey.Equal(newValidator.ConsensusPubkey))
	require.True(t, oldValidator.Tokens.Equal(newValidator.Tokens))
	require.True(t, oldValidator.DelegatorShares.Equal(newValidator.DelegatorShares))
	require.Equal(t, oldValidator.Commission, newValidator.Commission)

	// both delegations moved, shares untouched
	newDelegations, err := sk.GetValidatorDelegations(ctx, newValAddr)
	require.NoError(t, err)
	require.Len(t, newDelegations, 2)
	foundOther, foundSelf := false, false
	newSelfDelegator := sdk.AccAddress(newValAddr).String()
	oldSelfDelegator := sdk.AccAddress(oldValAddr).String()
	for _, del := range newDelegations {
		if del.DelegatorAddress == otherDelegator.String() {
			foundOther = true
			require.True(t, del.Shares.Equal(math.LegacyNewDec(1_000_000)))
		}
		if del.DelegatorAddress == newSelfDelegator {
			foundSelf = true
			require.True(t, del.Shares.Equal(oldValidator.DelegatorShares))
		}
		require.NotEqual(t, oldSelfDelegator, del.DelegatorAddress,
			"self-delegation must move to the NEW operator's own account, not stay under the old (leaked) one")
	}
	require.True(t, foundOther, "other delegator's delegation must have moved to the new validator")
	require.True(t, foundSelf, "self-delegation must be re-keyed to the new operator's own account")

	// distribution state moved
	curRewards, err := dk.GetValidatorCurrentRewards(ctx, newValAddr)
	require.NoError(t, err)
	require.Equal(t, uint64(1), curRewards.Period)
	commission, err := dk.GetValidatorAccumulatedCommission(ctx, newValAddr)
	require.NoError(t, err)
	require.True(t, commission.Commission.AmountOf("ario").Equal(math.LegacyNewDec(50)))
	outstanding, err := dk.GetValidatorOutstandingRewards(ctx, newValAddr)
	require.NoError(t, err)
	require.True(t, outstanding.Rewards.AmountOf("ario").Equal(math.LegacyNewDec(150)))
	histRewards, err := dk.GetValidatorHistoricalRewards(ctx, newValAddr, 0)
	require.NoError(t, err)
	require.Equal(t, uint32(1), histRewards.ReferenceCount)
	startInfo, err := dk.GetDelegatorStartingInfo(ctx, newValAddr, otherDelegator)
	require.NoError(t, err)
	require.True(t, startInfo.Stake.Equal(math.LegacyNewDec(1_000_000)))
	newSlashEvent, found, err := dk.GetValidatorSlashEvent(ctx, newValAddr, 1, 0)
	require.NoError(t, err)
	require.True(t, found, "slash-event history must have moved to the new validator address")
	require.True(t, newSlashEvent.Fraction.Equal(math.LegacyNewDecWithPrec(5, 2)))

	// multistaking locks moved, coin amounts untouched
	require.Equal(t, oldLocks, lockCount(msk, ctx, newValAddr.String()))
	require.NotEmpty(t, msk.GetValidatorMultiStakingCoin(ctx, newValAddr))

	// EndBlocker appends the captured "zero out old key" update alongside
	// the naturally-emitted "new key" update
	res, err := realio.EndBlocker(ctx)
	require.NoError(t, err)

	var sawZeroForOld, sawPowerForNew bool
	for _, u := range res.ValidatorUpdates {
		if u.Power == 0 {
			sawZeroForOld = true
		}
		if u.Power > 0 {
			sawPowerForNew = true
		}
	}
	require.True(t, sawZeroForOld, "EndBlocker must emit a zero-power update for the old consensus key")
	require.True(t, sawPowerForNew, "staking's own EndBlocker must emit a power update for the new consensus key")
}

// TestBeginBlockerRotationOrdering reproduces two real consensus failures
// hit on a live 4-node devnet, both "validator does not exist"
// (types.ErrNoValidatorFound) crashing every node.
//
// x/slashing's BeginBlocker processes the PREVIOUS block's vote info
// (CometBFT's LastCommitInfo is always one block behind). Per the ABCI
// spec, validator_updates returned at height H don't fully take effect
// until H+3: NextValidatorsHash updates at H+1, the new set starts
// proposing/voting at H+2, and LastCommitInfo first reflects the new set
// at H+3. That means the OLD validator keeps signing blocks H and H+1, so
// its votes still show up in LastCommitInfo as late as H+2's BeginBlocker.
//
//   - 1st crash: rotation ran before app.mm.BeginBlock, so the old
//     validator's ValidatorByConsAddr index was already gone when
//     x/slashing tried to process the fork block's OWN preceding vote.
//     Fixed by running the rotation after app.mm.BeginBlock.
//   - 2nd crash: even with that fix, migrateValidatorRecord still deleted
//     the old validator's plain record immediately at the fork height —
//     which broke x/slashing at H+1 and H+2, since the old validator's
//     votes for blocks H and H+1 are processed there and it no longer
//     existed. Fixed by leaving the old Validator/ValidatorByConsAddr
//     record in place indefinitely (only removing it from the power
//     index, so it stops participating going forward) rather than trying
//     to time a deferred cleanup.
//
// This drives three consecutive real app.BeginBlocker calls (H, H+1, H+2),
// each with vote info still referencing the old validator — exactly the
// full window a live chain goes through — and confirms none of them error.
func TestBeginBlockerRotationOrdering(t *testing.T) {
	realio := app.Setup(false, nil, 1)

	baseCtx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1})
	validators, err := realio.StakingKeeper.GetValidators(baseCtx, 10)
	require.NoError(t, err)
	oldValidator := validators[0]
	oldValAddr, err := sdk.ValAddressFromBech32(oldValidator.OperatorAddress)
	require.NoError(t, err)
	oldConsAddrBytes, err := oldValidator.GetConsAddr()
	require.NoError(t, err)
	oldPower := oldValidator.ConsensusPower(realio.StakingKeeper.PowerReduction(baseCtx))

	// Setup() doesn't seed x/slashing signing info for its genesis
	// validator the way a real chain (bonded via a real InitGenesis flow)
	// always does — seed it here so this test's vote-info processing
	// matches what x/slashing actually sees on a live chain.
	require.NoError(t, realio.SlashingKeeper.SetValidatorSigningInfo(baseCtx, sdk.ConsAddress(oldConsAddrBytes), slashingtypes.ValidatorSigningInfo{
		Address:     sdk.ConsAddress(oldConsAddrBytes).String(),
		StartHeight: 0,
	}))

	newValAddr := sdk.ValAddress(testutil.GenAddress())
	newConsPriv := ed25519.GenPrivKey()

	origRotations, origHeight := migrations.ValidatorRotations, migrations.ValidatorRotationHeight
	t.Cleanup(func() {
		migrations.ValidatorRotations, migrations.ValidatorRotationHeight = origRotations, origHeight
	})
	const rotationHeight = int64(2)
	migrations.ValidatorRotationHeight = rotationHeight
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      newValAddr.String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	oldVoteInfo := []abci.VoteInfo{
		{
			Validator:   abci.Validator{Address: oldConsAddrBytes, Power: oldPower},
			BlockIdFlag: tmproto.BlockIDFlagCommit,
		},
	}

	// H, H+1, H+2: the old validator's votes for the prior 3 blocks are
	// still being processed here, per the ABCI delay above.
	for height := rotationHeight; height <= rotationHeight+2; height++ {
		ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: height}).
			WithBlockGasMeter(storetypes.NewInfiniteGasMeter()).
			WithVoteInfos(oldVoteInfo)

		require.NotPanicsf(t, func() {
			_, beginErr := realio.BeginBlocker(ctx)
			require.NoErrorf(t, beginErr, "BeginBlocker failed at height %d", height)
		}, "BeginBlocker panicked at height %d", height)
	}

	finalCtx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: rotationHeight + 2})
	_, err = realio.StakingKeeper.GetValidator(finalCtx, newValAddr)
	require.NoError(t, err, "rotation must actually have run")
	_, err = realio.StakingKeeper.GetValidator(finalCtx, oldValAddr)
	require.NoError(t, err, "old validator record is kept in place indefinitely, not deleted")
}

// TestRotateValidatorPanicsOnRedelegation proves checkNoRedelegations
// actually fires: redelegation migration isn't implemented, so a
// redelegation touching the validator being rotated must fail the fork
// loudly rather than leave an orphaned record behind once the old
// validator identity is deleted.
func TestRotateValidatorPanicsOnRedelegation(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1}).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	sk := realio.StakingKeeper
	validators, err := sk.GetValidators(ctx, 10)
	require.NoError(t, err)
	oldValidator := validators[0]
	oldValAddr, err := sdk.ValAddressFromBech32(oldValidator.OperatorAddress)
	require.NoError(t, err)

	otherValAddr := sdk.ValAddress(testutil.GenAddress())
	require.NoError(t, sk.SetRedelegation(ctx, stakingtypes.Redelegation{
		DelegatorAddress:    testutil.GenAddress().String(),
		ValidatorSrcAddress: oldValAddr.String(),
		ValidatorDstAddress: otherValAddr.String(),
		Entries: []stakingtypes.RedelegationEntry{
			{CreationHeight: 1, CompletionTime: ctx.BlockTime(), InitialBalance: math.NewInt(1), SharesDst: math.LegacyOneDec()},
		},
	}))

	newConsPriv := ed25519.GenPrivKey()
	origRotations := migrations.ValidatorRotations
	t.Cleanup(func() { migrations.ValidatorRotations = origRotations })
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      sdk.ValAddress(testutil.GenAddress()).String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	require.Panics(t, func() { migrations.RotateValidators(realio.MigrationKeepers(), ctx) })

	// and the validator record must be untouched — checkNoRedelegations
	// runs before any mutation
	_, err = sk.GetValidator(ctx, oldValAddr)
	require.NoError(t, err)
}

// TestRotateValidatorPanicsOnSelfCollision guards against a data-entry
// mistake in ValidatorRotations: NewOperator accidentally equal to
// OldOperator. Without the guard, migrateValidatorRecord's final write (the
// zeroed-Tokens ghost, keyed by the OLD address) lands on the SAME key as
// the just-migrated real record, silently zeroing out the validator's real
// stake instead of migrating it — a self-inflicted, undetected corruption
// far worse than a clean panic.
func TestRotateValidatorPanicsOnSelfCollision(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1}).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	sk := realio.StakingKeeper
	validators, err := sk.GetValidators(ctx, 10)
	require.NoError(t, err)
	oldValidator := validators[0]
	oldValAddr, err := sdk.ValAddressFromBech32(oldValidator.OperatorAddress)
	require.NoError(t, err)
	tokensBefore := oldValidator.Tokens

	newConsPriv := ed25519.GenPrivKey()
	origRotations := migrations.ValidatorRotations
	t.Cleanup(func() { migrations.ValidatorRotations = origRotations })
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      oldValidator.OperatorAddress, // mistake: same as OldOperator
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	require.Panics(t, func() { migrations.RotateValidators(realio.MigrationKeepers(), ctx) })

	// nothing should have been mutated — the guard runs before any write
	stillThere, err := sk.GetValidator(ctx, oldValAddr)
	require.NoError(t, err)
	require.Truef(t, stillThere.Tokens.Equal(tokensBefore),
		"validator's real tokens must be untouched, not zeroed by a self-collision (before=%s after=%s)",
		tokensBefore, stillThere.Tokens)
}

// TestRotateValidatorPanicsOnExistingNewOperator guards against a
// data-entry mistake in ValidatorRotations: NewOperator accidentally
// pointing at an address that's already a registered validator. Without the
// guard, migrateValidatorRecord's SetValidator calls silently overwrite
// that unrelated validator's entire record with the migrated one.
func TestRotateValidatorPanicsOnExistingNewOperator(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1}).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	sk := realio.StakingKeeper
	validators, err := sk.GetValidators(ctx, 10)
	require.NoError(t, err)
	oldValidator := validators[0]

	// a second, unrelated real validator that NewOperator will mistakenly
	// point at
	collidingValAddr := sdk.ValAddress(testutil.GenAddress())
	collidingConsPriv := ed25519.GenPrivKey()
	collidingValidator, err := stakingtypes.NewValidator(collidingValAddr.String(), collidingConsPriv.PubKey(), stakingtypes.Description{Moniker: "colliding"})
	require.NoError(t, err)
	collidingValidator.Tokens = math.NewInt(12345)
	collidingValidator.DelegatorShares = math.LegacyNewDec(12345)
	require.NoError(t, sk.SetValidator(ctx, collidingValidator))
	collidingTokensBefore := collidingValidator.Tokens

	newConsPriv := ed25519.GenPrivKey()
	origRotations := migrations.ValidatorRotations
	t.Cleanup(func() { migrations.ValidatorRotations = origRotations })
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      collidingValidator.OperatorAddress, // mistake: already a real validator
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	require.Panics(t, func() { migrations.RotateValidators(realio.MigrationKeepers(), ctx) })

	// the colliding validator must be completely untouched
	stillThere, err := sk.GetValidator(ctx, collidingValAddr)
	require.NoError(t, err)
	require.Equal(t, collidingValidator.ConsensusPubkey.Value, stillThere.ConsensusPubkey.Value,
		"colliding validator's consensus pubkey must be untouched")
	require.Truef(t, stillThere.Tokens.Equal(collidingTokensBefore),
		"colliding validator's tokens must be untouched (before=%s after=%s)", collidingTokensBefore, stillThere.Tokens)
}

// TestRotateValidatorMigratesInFlightUnbonding proves an unbonding delegation
// already in progress at rotation time actually pays out at maturity,
// instead of being silently orphaned. migrateUnbondingDelegations moves the
// UnbondingDelegation record itself, but x/staking's EndBlocker maturity
// processing (BlockValidatorUpdates) doesn't look at that record's fields —
// it dequeues a DVPair{delegator, validator} from the completion-time queue
// and looks up the record by THAT pair. If the queue still names the old
// validator after the record has moved, CompleteUnbonding fails to find it,
// the error is silently swallowed (`if err != nil { continue }` in the SDK),
// and the queue entry is dequeued anyway — permanently orphaning the funds,
// since nothing ever re-enqueues it. This drives the real queue-dequeue and
// CompleteUnbonding path (not just checking the record exists) to prove
// that doesn't happen.
func TestRotateValidatorMigratesInFlightUnbonding(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1}).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	sk := realio.StakingKeeper
	validators, err := sk.GetValidators(ctx, 10)
	require.NoError(t, err)
	oldValidator := validators[0]
	oldValAddr, err := sdk.ValAddressFromBech32(oldValidator.OperatorAddress)
	require.NoError(t, err)

	delegator := testutil.GenAddress()
	realio.AccountKeeper.SetAccount(ctx, realio.AccountKeeper.NewAccountWithAddress(ctx, delegator))
	completionTime := ctx.BlockTime().Add(time.Hour)
	newUbd, err := sk.SetUnbondingDelegationEntry(ctx, delegator, oldValAddr, ctx.BlockHeight(), completionTime, math.NewInt(4242))
	require.NoError(t, err)
	require.NoError(t, sk.InsertUBDQueue(ctx, newUbd, completionTime))

	// sanity: before rotation, the queue names the OLD validator
	sliceBefore, err := sk.GetUBDQueueTimeSlice(ctx, completionTime)
	require.NoError(t, err)
	require.Len(t, sliceBefore, 1)
	require.Equal(t, oldValidator.OperatorAddress, sliceBefore[0].ValidatorAddress)

	newValAddr := sdk.ValAddress(testutil.GenAddress())
	newConsPriv := ed25519.GenPrivKey()
	origRotations := migrations.ValidatorRotations
	t.Cleanup(func() { migrations.ValidatorRotations = origRotations })
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      newValAddr.String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	migrations.RotateValidators(realio.MigrationKeepers(), ctx)

	// the queue must now name the NEW validator — this is the actual bug:
	// without the fix, this slice is still empty (nothing to replace) or
	// still names the old validator, and maturity would silently swallow it
	sliceAfter, err := sk.GetUBDQueueTimeSlice(ctx, completionTime)
	require.NoError(t, err)
	require.Len(t, sliceAfter, 1)
	require.Equal(t, newValAddr.String(), sliceAfter[0].ValidatorAddress)
	require.Equal(t, delegator.String(), sliceAfter[0].DelegatorAddress)

	// prove it end-to-end: dequeue at maturity exactly like EndBlocker does,
	// and complete it for real — funds must actually reach the delegator.
	matureTime := completionTime.Add(time.Second)
	require.NoError(t, realio.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 4242))))
	require.NoError(t, realio.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, stakingtypes.NotBondedPoolName, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 4242))))

	matureCtx := ctx.WithBlockTime(matureTime)
	matured, err := sk.DequeueAllMatureUBDQueue(matureCtx, matureTime)
	require.NoError(t, err)
	require.Len(t, matured, 1, "the matured queue entry must be the migrated (new validator) one")
	require.Equal(t, newValAddr.String(), matured[0].ValidatorAddress)

	balanceBefore := realio.BankKeeper.GetBalance(matureCtx, delegator, sdk.DefaultBondDenom)
	_, err = sk.CompleteUnbonding(matureCtx, delegator, newValAddr)
	require.NoError(t, err, "CompleteUnbonding must succeed against the migrated record — this is what silently failed before the fix")
	balanceAfter := realio.BankKeeper.GetBalance(matureCtx, delegator, sdk.DefaultBondDenom)
	require.Equal(t, int64(4242), balanceAfter.Amount.Sub(balanceBefore.Amount).Int64(), "funds must actually reach the delegator")
}

// TestRotateValidatorMigratesGenesisImportedUnbondingType proves a subtler
// variant of the bug above: an unbonding delegation that came from a real
// genesis import (x/staking's own InitGenesis only ever calls
// SetUnbondingDelegation + InsertUBDQueue for those — never
// SetUnbondingDelegationByUnbondingID) has NEITHER the unbonding-ID pointer
// index nor its companion "type" index set at all. Migrating such an entry
// by hand-rolling just the pointer index (as an earlier version of this
// code did) leaves a WORSE state than before: the pointer now resolves, but
// PutUnbondingOnHold/UnbondingCanComplete (which key off the type index)
// still fail. Using the keeper's own SetUnbondingDelegationByUnbondingID
// sets both together, matching what a live Undelegate call would produce.
func TestRotateValidatorMigratesGenesisImportedUnbondingType(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1}).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	sk := realio.StakingKeeper
	validators, err := sk.GetValidators(ctx, 10)
	require.NoError(t, err)
	oldValidator := validators[0]
	oldValAddr, err := sdk.ValAddressFromBech32(oldValidator.OperatorAddress)
	require.NoError(t, err)

	delegator := testutil.GenAddress()
	realio.AccountKeeper.SetAccount(ctx, realio.AccountKeeper.NewAccountWithAddress(ctx, delegator))
	completionTime := ctx.BlockTime().Add(time.Hour)
	const unbondingID = uint64(999)

	// mirror x/staking's own InitGenesis exactly: SetUnbondingDelegation +
	// InsertUBDQueue only — no SetUnbondingDelegationByUnbondingID, same as
	// a real genesis-imported UBD gets.
	ubd := stakingtypes.UnbondingDelegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: oldValAddr.String(),
		Entries: []stakingtypes.UnbondingDelegationEntry{
			stakingtypes.NewUnbondingDelegationEntry(ctx.BlockHeight(), completionTime, math.NewInt(4242), unbondingID),
		},
	}
	require.NoError(t, sk.SetUnbondingDelegation(ctx, ubd))
	require.NoError(t, sk.InsertUBDQueue(ctx, ubd, completionTime))

	// sanity: matches real InitGenesis behavior — neither index exists yet
	_, err = sk.GetUnbondingDelegationByUnbondingID(ctx, unbondingID)
	require.Error(t, err, "sanity: genesis import never sets the pointer index")
	_, err = sk.GetUnbondingType(ctx, unbondingID)
	require.Error(t, err, "sanity: genesis import never sets the type index either")

	newValAddr := sdk.ValAddress(testutil.GenAddress())
	newConsPriv := ed25519.GenPrivKey()
	origRotations := migrations.ValidatorRotations
	t.Cleanup(func() { migrations.ValidatorRotations = origRotations })
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      newValAddr.String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	migrations.RotateValidators(realio.MigrationKeepers(), ctx)

	unbondingType, err := sk.GetUnbondingType(ctx, unbondingID)
	require.NoErrorf(t, err, "type index must be set by migration, not just the pointer — "+
		"otherwise PutUnbondingOnHold/UnbondingCanComplete break for this entry")
	require.Equal(t, stakingtypes.UnbondingType_UnbondingDelegation, unbondingType)

	migrated, err := sk.GetUnbondingDelegationByUnbondingID(ctx, unbondingID)
	require.NoError(t, err)
	require.Equal(t, newValAddr.String(), migrated.ValidatorAddress)
	require.Equal(t, delegator.String(), migrated.DelegatorAddress)
}

// TestRotateValidatorGhostRejectsNewDelegations proves the zeroed-Tokens
// ghost left under the old operator address can't accept fresh delegations
// by accident: Tokens=0 with positive DelegatorShares makes
// Validator.InvalidExRate() true, which x/staking's own Delegate keeper
// method rejects outright — nobody delegating to what looks like a normal
// bonded validator in a query response ends up with funds stuck earning
// nothing under a rotated-away identity.
func TestRotateValidatorGhostRejectsNewDelegations(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1}).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	sk := realio.StakingKeeper
	validators, err := sk.GetValidators(ctx, 10)
	require.NoError(t, err)
	oldValidator := validators[0]
	oldValAddr, err := sdk.ValAddressFromBech32(oldValidator.OperatorAddress)
	require.NoError(t, err)

	newValAddr := sdk.ValAddress(testutil.GenAddress())
	newConsPriv := ed25519.GenPrivKey()
	origRotations := migrations.ValidatorRotations
	t.Cleanup(func() { migrations.ValidatorRotations = origRotations })
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      newValAddr.String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	migrations.RotateValidators(realio.MigrationKeepers(), ctx)

	oldGhost, err := sk.GetValidator(ctx, oldValAddr)
	require.NoError(t, err)

	newDelegator := testutil.GenAddress()
	realio.AccountKeeper.SetAccount(ctx, realio.AccountKeeper.NewAccountWithAddress(ctx, newDelegator))
	require.NoError(t, realio.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))))
	require.NoError(t, realio.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, newDelegator, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))))

	_, err = sk.Delegate(ctx, newDelegator, math.NewInt(1000), stakingtypes.Unbonded, oldGhost, true)
	require.ErrorIsf(t, err, stakingtypes.ErrDelegatorShareExRateInvalid,
		"delegating fresh funds into the rotated-away identity must be rejected, not silently accepted")
}

// TestRotateValidatorRegistersNewConsensusPubkeyForEquivocation proves the
// new consensus key can actually be slashed for a double-sign after
// rotation. AfterValidatorCreated — the only place x/slashing's
// AddrPubkeyRelation ever gets written (via AddPubkey) — fires from the
// normal MsgCreateValidator flow, not from a raw SetValidator call like
// this migration makes; without migrateValidatorRecord explicitly calling
// AddPubkey too, x/evidence's handleEquivocationEvidence would fail its
// GetPubkey lookup and silently drop any future equivocation evidence
// against the new key — no slash, no jail, no tombstone. Downtime slashing
// is unaffected (it only needs SetValidatorSigningInfo, already covered
// elsewhere) — this is specifically about the equivocation path, which
// matters most for an incident whose entire premise is compromised
// consensus keys.
func TestRotateValidatorRegistersNewConsensusPubkeyForEquivocation(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1}).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	sk := realio.StakingKeeper
	validators, err := sk.GetValidators(ctx, 10)
	require.NoError(t, err)
	oldValidator := validators[0]

	newValAddr := sdk.ValAddress(testutil.GenAddress())
	newConsPriv := ed25519.GenPrivKey()
	newConsPubKey := newConsPriv.PubKey()

	origRotations := migrations.ValidatorRotations
	t.Cleanup(func() { migrations.ValidatorRotations = origRotations })
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      newValAddr.String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPubKey.(*ed25519.PubKey)),
		},
	}

	// sanity: before rotation, nobody has registered this brand-new key yet
	_, err = realio.SlashingKeeper.GetPubkey(ctx, newConsPubKey.Address().Bytes())
	require.Error(t, err, "sanity: the new consensus pubkey must not be registered before rotation runs")

	migrations.RotateValidators(realio.MigrationKeepers(), ctx)

	resolved, err := realio.SlashingKeeper.GetPubkey(ctx, newConsPubKey.Address().Bytes())
	require.NoErrorf(t, err, "the new consensus pubkey must be resolvable via GetPubkey after rotation — "+
		"x/evidence's handleEquivocationEvidence depends on exactly this lookup to slash a future double-sign")
	require.True(t, newConsPubKey.Equals(resolved), "the registered pubkey must be the actual new consensus key")
}

// TestRotateValidatorExportGenesisRoundTrip is the direct reproduction of
// the export→import failure a reviewer flagged: before the Tokens=0 fix,
// exporting genesis after a rotation and re-importing it panicked with
// "bonded pool balance is different from bonded coins", because the old
// (ghost) and new validator records both claimed the full Tokens amount
// while the bonded pool only ever actually held one share of it. This
// re-imports the export against the SAME context right after rotating —
// the bonded pool's real bank balance doesn't change across that round
// trip, so it's a faithful, minimal reproduction of the real bug without
// needing to spin up a second app.
func TestRotateValidatorExportGenesisRoundTrip(t *testing.T) {
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1}).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	sk := realio.StakingKeeper
	validators, err := sk.GetValidators(ctx, 10)
	require.NoError(t, err)
	oldValidator := validators[0]

	newValAddr := sdk.ValAddress(testutil.GenAddress())
	newConsPriv := ed25519.GenPrivKey()
	origRotations := migrations.ValidatorRotations
	t.Cleanup(func() { migrations.ValidatorRotations = origRotations })
	migrations.ValidatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
		AuthorizeSymbol  string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      newValAddr.String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	migrations.RotateValidators(realio.MigrationKeepers(), ctx)

	exported := sk.ExportGenesis(ctx)
	require.NotPanics(t, func() {
		sk.InitGenesis(ctx, exported)
	}, "re-importing a genesis exported after rotation must not panic on a bonded-pool/bonded-coins mismatch")
}

func lockCount(msk multistakingkeeper.Keeper, ctx sdk.Context, valAddr string) int {
	count := 0
	msk.MultiStakingLockIterator(ctx, func(lock multistakingtypes.MultiStakingLock) bool {
		if lock.LockID.ValAddr == valAddr {
			count++
		}
		return false
	})
	return count
}

func pubKeyB64(t *testing.T, pk *ed25519.PubKey) string {
	t.Helper()
	return base64.StdEncoding.EncodeToString(pk.Bytes())
}
