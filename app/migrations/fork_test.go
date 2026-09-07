package migrations_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/app/migrations"
)

func TestFork(t *testing.T) {
	realio := app.Setup(false, nil, 1)

	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: migrations.ForkHeight})
	stakingKeeper := realio.StakingKeeper

	timeKey := time.Date(2024, 4, 1, 1, 1, 1, 1, time.UTC)

	duplicativeUnbondingDelegation := stakingtypes.UnbondingDelegation{
		DelegatorAddress: "test_del_1",
		ValidatorAddress: "test_val_1",
		Entries: []stakingtypes.UnbondingDelegationEntry{
			stakingtypes.NewUnbondingDelegationEntry(migrations.ForkHeight, timeKey, math.OneInt(), 0),
		},
	}

	err := stakingKeeper.InsertUBDQueue(ctx, duplicativeUnbondingDelegation, timeKey)
	require.NoError(t, err)
	err = stakingKeeper.InsertUBDQueue(ctx, duplicativeUnbondingDelegation, timeKey)
	require.NoError(t, err)

	duplicativeRedelegation := stakingtypes.Redelegation{
		DelegatorAddress:    "test_del_1",
		ValidatorSrcAddress: "test_val_1",
		ValidatorDstAddress: "test_val_2",
		Entries: []stakingtypes.RedelegationEntry{
			stakingtypes.NewRedelegationEntry(migrations.ForkHeight, timeKey, math.OneInt(), math.LegacyOneDec(), 0),
		},
	}
	err = stakingKeeper.InsertRedelegationQueue(ctx, duplicativeRedelegation, timeKey)
	require.NoError(t, err)

	err = stakingKeeper.InsertRedelegationQueue(ctx, duplicativeRedelegation, timeKey)
	require.NoError(t, err)

	err = stakingKeeper.InsertRedelegationQueue(ctx, duplicativeRedelegation, timeKey)
	require.NoError(t, err)

	duplicativeVal := stakingtypes.Validator{
		OperatorAddress: "test_op",
		UnbondingHeight: migrations.ForkHeight,
		UnbondingTime:   timeKey,
	}

	err = stakingKeeper.InsertUnbondingValidatorQueue(ctx, duplicativeVal)
	require.NoError(t, err)

	err = stakingKeeper.InsertUnbondingValidatorQueue(ctx, duplicativeVal)
	require.NoError(t, err)

	require.True(t, checkDuplicateUBDQueue(ctx, realio))
	require.True(t, checkDuplicateRelegationQueue(ctx, realio))
	require.True(t, checkDuplicateValQueue(ctx, realio))

	realio.ScheduleForkUpgrade(ctx)

	require.False(t, checkDuplicateUBDQueue(ctx, realio))
	require.False(t, checkDuplicateRelegationQueue(ctx, realio))
	require.False(t, checkDuplicateValQueue(ctx, realio))

	dvPairs, err := stakingKeeper.GetUBDQueueTimeSlice(ctx, timeKey)
	require.NoError(t, err)
	require.Equal(t, dvPairs[0].DelegatorAddress, duplicativeUnbondingDelegation.DelegatorAddress)
	require.Equal(t, dvPairs[0].ValidatorAddress, duplicativeUnbondingDelegation.ValidatorAddress)

	triplets, err := stakingKeeper.GetRedelegationQueueTimeSlice(ctx, timeKey)
	require.NoError(t, err)
	require.Equal(t, triplets[0].DelegatorAddress, duplicativeRedelegation.DelegatorAddress)
	require.Equal(t, triplets[0].ValidatorDstAddress, duplicativeRedelegation.ValidatorDstAddress)
	require.Equal(t, triplets[0].ValidatorSrcAddress, duplicativeRedelegation.ValidatorSrcAddress)

	vals, err := stakingKeeper.GetUnbondingValidators(ctx, timeKey, migrations.ForkHeight)
	require.NoError(t, err)
	require.Equal(t, vals[0], duplicativeVal.OperatorAddress)
}

func checkDuplicateUBDQueue(ctx sdk.Context, realio *app.RealioNetwork) bool {
	ubdIter, _ := realio.StakingKeeper.UBDQueueIterator(ctx, migrations.OneEternityLater)
	defer ubdIter.Close()

	for ; ubdIter.Valid(); ubdIter.Next() {
		timeslice := stakingtypes.DVPairs{}
		value := ubdIter.Value()
		realio.AppCodec().MustUnmarshal(value, &timeslice)
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

func checkDuplicateRelegationQueue(ctx sdk.Context, realio *app.RealioNetwork) bool {
	redeIter, _ := realio.StakingKeeper.RedelegationQueueIterator(ctx, migrations.OneEternityLater)
	defer redeIter.Close()

	for ; redeIter.Valid(); redeIter.Next() {
		timeslice := stakingtypes.DVVTriplets{}
		value := redeIter.Value()
		realio.AppCodec().MustUnmarshal(value, &timeslice)
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

func checkDuplicateValQueue(ctx sdk.Context, realio *app.RealioNetwork) bool {
	valsIter, _ := realio.StakingKeeper.ValidatorQueueIterator(ctx, migrations.OneEternityLater, 9999)
	defer valsIter.Close()

	for ; valsIter.Valid(); valsIter.Next() {
		timeslice := stakingtypes.ValAddresses{}
		value := valsIter.Value()
		realio.AppCodec().MustUnmarshal(value, &timeslice)
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
