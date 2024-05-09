package app

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestFork(t *testing.T) {
	realio := Setup(false, nil)

	ctx := realio.BaseApp.NewContext(false, tmproto.Header{Height: ForkHeight})
	stakingKeeper := realio.StakingKeeper

	timeKey := time.Date(2024, 4, 1, 1, 1, 1, 1, time.UTC)

	duplicativeUnbondingDelegation := stakingtypes.UnbondingDelegation{
		DelegatorAddress: "test_del_1",
		ValidatorAddress: "test_val_1",
		Entries: []stakingtypes.UnbondingDelegationEntry{
			stakingtypes.NewUnbondingDelegationEntry(ForkHeight, timeKey, math.OneInt(), 1),
		},
	}

	stakingKeeper.InsertUBDQueue(ctx, duplicativeUnbondingDelegation, timeKey)
	stakingKeeper.InsertUBDQueue(ctx, duplicativeUnbondingDelegation, timeKey)

	duplicativeRedelegation := stakingtypes.Redelegation{
		DelegatorAddress:    "test_del_1",
		ValidatorSrcAddress: "test_val_1",
		ValidatorDstAddress: "test_val_2",
		Entries: []stakingtypes.RedelegationEntry{
			stakingtypes.NewRedelegationEntry(ForkHeight, timeKey, math.OneInt(), sdk.OneDec(), 1),
		},
	}
	stakingKeeper.InsertRedelegationQueue(ctx, duplicativeRedelegation, timeKey)
	stakingKeeper.InsertRedelegationQueue(ctx, duplicativeRedelegation, timeKey)
	stakingKeeper.InsertRedelegationQueue(ctx, duplicativeRedelegation, timeKey)

	duplicativeVal := stakingtypes.Validator{
		OperatorAddress: "test_op",
		UnbondingHeight: ForkHeight,
		UnbondingTime:   timeKey,
	}

	stakingKeeper.InsertUnbondingValidatorQueue(ctx, duplicativeVal)
	stakingKeeper.InsertUnbondingValidatorQueue(ctx, duplicativeVal)

	require.True(t, checkDuplicateUBDQueue(ctx, *realio))
	require.True(t, checkDuplicateRelegationQueue(ctx, *realio))
	require.True(t, checkDuplicateValQueue(ctx, *realio))

	realio.ScheduleForkUpgrade(ctx)

	require.False(t, checkDuplicateUBDQueue(ctx, *realio))
	require.False(t, checkDuplicateRelegationQueue(ctx, *realio))
	require.False(t, checkDuplicateValQueue(ctx, *realio))

	dvPairs := stakingKeeper.GetUBDQueueTimeSlice(ctx, timeKey)
	require.Equal(t, dvPairs[0].DelegatorAddress, duplicativeUnbondingDelegation.DelegatorAddress)
	require.Equal(t, dvPairs[0].ValidatorAddress, duplicativeUnbondingDelegation.ValidatorAddress)

	triplets := stakingKeeper.GetRedelegationQueueTimeSlice(ctx, timeKey)
	require.Equal(t, triplets[0].DelegatorAddress, duplicativeRedelegation.DelegatorAddress)
	require.Equal(t, triplets[0].ValidatorDstAddress, duplicativeRedelegation.ValidatorDstAddress)
	require.Equal(t, triplets[0].ValidatorSrcAddress, duplicativeRedelegation.ValidatorSrcAddress)

	vals := stakingKeeper.GetUnbondingValidators(ctx, timeKey, ForkHeight)
	require.Equal(t, vals[0], duplicativeVal.OperatorAddress)
}

func checkDuplicateUBDQueue(ctx sdk.Context, realio RealioNetwork) bool {
	ubdIter := realio.StakingKeeper.UBDQueueIterator(ctx, oneEnternityLater)
	defer ubdIter.Close()

	for ; ubdIter.Valid(); ubdIter.Next() {
		timeslice := stakingtypes.DVPairs{}
		value := ubdIter.Value()
		realio.appCodec.MustUnmarshal(value, &timeslice)
		if checkDuplicateUBD(timeslice.Pairs) {
			return true
		}
	}
	return false
}

func checkDuplicateUBD(eels []stakingtypes.DVPair) bool {
	uniqueEles := map[string]bool{}
	for _, ele := range eels {
		uniqueEles[ele.String()] = true
	}

	return len(uniqueEles) != len(eels)
}

func checkDuplicateRelegationQueue(ctx sdk.Context, realio RealioNetwork) bool {
	redeIter := realio.StakingKeeper.RedelegationQueueIterator(ctx, oneEnternityLater)
	defer redeIter.Close()

	for ; redeIter.Valid(); redeIter.Next() {
		timeslice := stakingtypes.DVVTriplets{}
		value := redeIter.Value()
		realio.appCodec.MustUnmarshal(value, &timeslice)
		if checkDuplicateRedelegation(timeslice.Triplets) {
			return true
		}
	}
	return false
}

func checkDuplicateRedelegation(eels []stakingtypes.DVVTriplet) bool {
	uniqueEles := map[string]bool{}
	for _, ele := range eels {
		uniqueEles[ele.String()] = true
	}

	return len(uniqueEles) != len(eels)
}

func checkDuplicateValQueue(ctx sdk.Context, realio RealioNetwork) bool {
	valsIter := realio.StakingKeeper.ValidatorQueueIterator(ctx, oneEnternityLater, 9999)
	defer valsIter.Close()

	for ; valsIter.Valid(); valsIter.Next() {
		timeslice := stakingtypes.ValAddresses{}
		value := valsIter.Value()
		realio.appCodec.MustUnmarshal(value, &timeslice)
		if checkDuplicateValAddr(timeslice.Addresses) {
			return true
		}
	}
	return false
}

func checkDuplicateValAddr(eels []string) bool {
	uniqueEles := map[string]bool{}
	for _, ele := range eels {
		uniqueEles[ele] = true
	}

	return len(uniqueEles) != len(eels)
}
