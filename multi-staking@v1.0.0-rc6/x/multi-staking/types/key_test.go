package types_test

import (
	"testing"

	"github.com/realio-tech/multi-staking-module/test"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/stretchr/testify/require"
)

func TestDelAddrAndValAddrFromLockID(t *testing.T) {
	val := test.GenValAddress()
	del := test.GenAddress()

	lockID := multistakingtypes.MultiStakingLockID(del.String(), val.String())
	lockBytes := lockID.ToBytes()
	rsDel, rsVal := multistakingtypes.DelAddrAndValAddrFromLockID(lockBytes)

	require.Equal(t, del, rsDel)
	require.Equal(t, val, rsVal)
}

func TestMultiStakingLockIterator(t *testing.T) {
	val := test.GenValAddress()
	delA := test.GenAddress()
	delB := test.GenAddress()

	lockIDA := multistakingtypes.LockID{
		MultiStakerAddr: delA.String(),
		ValAddr:         val.String(),
	}

	lockIDB := multistakingtypes.LockID{
		MultiStakerAddr: delB.String(),
		ValAddr:         val.String(),
	}

	require.NotEqual(t, lockIDA, lockIDB)
	require.NotEqual(t, lockIDA.ToBytes(), lockIDB.ToBytes())
}
