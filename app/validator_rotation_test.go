package app

import (
	"encoding/base64"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/realiotech/realio-network/testutil"
)

// TestRotateValidator exercises rotateValidators end-to-end against a real
// app: a validator with two delegators (self + one other, mirroring the
// "72 other delegators" shape found in the real leaked-validator genesis),
// distribution history, and multistaking locks, migrated to a brand new
// operator address + consensus key. Every piece of state tied to the old
// identity must move — nothing about delegator addresses, shares, or coin
// amounts should change.
func TestRotateValidator(t *testing.T) {
	realio := Setup(false, nil, 1)
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
	msk.SetMultiStakingLock(ctx, multistakingtypes.NewMultiStakingLock(otherLockID, MultiStakingCoinA))

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

	oldLocks := lockCount(msk, ctx, oldValAddr.String())
	require.Equal(t, 2, oldLocks, "expected self + other delegator lock before rotation")

	// prepare a brand new operator + consensus identity
	newOperatorKey := testutil.GenAddress()
	newValAddr := sdk.ValAddress(newOperatorKey)
	newConsPriv := ed25519.GenPrivKey()
	newConsPubKeyAny, err := codectypes.NewAnyWithValue(newConsPriv.PubKey())
	require.NoError(t, err)

	oldConsPubKey := oldValidator.ConsensusPubkey

	origRotations := validatorRotations
	t.Cleanup(func() { validatorRotations = origRotations; pendingValidatorZeroUpdates = nil })
	validatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      newValAddr.String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	rotateValidators(realio, ctx)

	// old identity is gone
	_, err = sk.GetValidator(ctx, oldValAddr)
	require.Error(t, err)
	require.Empty(t, lockCount(msk, ctx, oldValAddr.String()))
	require.Empty(t, msk.GetValidatorMultiStakingCoin(ctx, oldValAddr))
	oldCurRewards, err := dk.GetValidatorCurrentRewards(ctx, oldValAddr)
	require.NoError(t, err)
	require.True(t, oldCurRewards.Rewards.IsZero(), "old validator's current rewards must be cleared, not just readable-as-zero-value")
	oldCommission, err := dk.GetValidatorAccumulatedCommission(ctx, oldValAddr)
	require.NoError(t, err)
	require.True(t, oldCommission.Commission.IsZero())

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
	foundOther := false
	for _, del := range newDelegations {
		if del.DelegatorAddress == otherDelegator.String() {
			foundOther = true
			require.True(t, del.Shares.Equal(math.LegacyNewDec(1_000_000)))
		}
	}
	require.True(t, foundOther, "other delegator's delegation must have moved to the new validator")

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

	// multistaking locks moved, coin amounts untouched
	require.Equal(t, oldLocks, lockCount(msk, ctx, newValAddr.String()))
	require.NotEmpty(t, msk.GetValidatorMultiStakingCoin(ctx, newValAddr))

	// captured ABCI "zero out old key" update
	require.Len(t, pendingValidatorZeroUpdates, 1)
	require.EqualValues(t, 0, pendingValidatorZeroUpdates[0].Power)

	// EndBlocker appends it alongside the naturally-emitted "new key" update
	res, err := realio.EndBlocker(ctx)
	require.NoError(t, err)
	require.Empty(t, pendingValidatorZeroUpdates, "EndBlocker must drain the pending queue")

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

// TestRotateValidatorPanicsOnRedelegation proves checkNoRedelegations
// actually fires: redelegation migration isn't implemented, so a
// redelegation touching the validator being rotated must fail the fork
// loudly rather than leave an orphaned record behind once the old
// validator identity is deleted.
func TestRotateValidatorPanicsOnRedelegation(t *testing.T) {
	realio := Setup(false, nil, 1)
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
	origRotations := validatorRotations
	t.Cleanup(func() { validatorRotations = origRotations; pendingValidatorZeroUpdates = nil })
	validatorRotations = []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
	}{
		{
			OldOperator:      oldValidator.OperatorAddress,
			NewOperator:      sdk.ValAddress(testutil.GenAddress()).String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		},
	}

	require.Panics(t, func() { rotateValidators(realio, ctx) })

	// and the validator record must be untouched — checkNoRedelegations
	// runs before any mutation
	_, err = sk.GetValidator(ctx, oldValAddr)
	require.NoError(t, err)
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
