package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "multistaking"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore keys
var (
	BondWeightKey = []byte{0x00}

	ValidatorMultiStakingCoinKey = []byte{0x01}

	MultiStakingLockPrefix = []byte{0x02}

	MultiStakingUnlockPrefix = []byte{0x11} // key for an unbonding-delegation
)

func KeyPrefix(key string) []byte {
	return []byte(key)
}

// GetBondWeightKeyKey returns a key for an index containing the bond coin weight
func GetBondWeightKey(tokenDenom string) []byte {
	return append(BondWeightKey, []byte(tokenDenom)...)
}

// GetValidatorMultiStakingCoinKey returns a key for an index containing the bond denom of a validator
func GetValidatorMultiStakingCoinKey(valAddr sdk.ValAddress) []byte {
	return append(ValidatorMultiStakingCoinKey, []byte(valAddr)...)
}

func MultiStakingLockID(multiStakerAddr string, valAddr string) LockID {
	return LockID{MultiStakerAddr: multiStakerAddr, ValAddr: valAddr}
}

func MultiStakingUnlockID(multiStakerAddr string, valAddr string) UnlockID {
	return UnlockID{MultiStakerAddr: multiStakerAddr, ValAddr: valAddr}
}

func DelAddrAndValAddrFromLockID(lockID []byte) (multiStakerAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	lenMultiStakerAddr := lockID[1]

	multiStakerAddr = lockID[2 : lenMultiStakerAddr+2]

	valAddr = lockID[2+lenMultiStakerAddr:]

	return multiStakerAddr, valAddr
}

func DelAddrAndValAddrFromUnlockID(unlockID []byte) (multiStakerAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	lenMultiStakerAddr := unlockID[1]

	multiStakerAddr = unlockID[2 : lenMultiStakerAddr+2]

	valAddr = unlockID[2+lenMultiStakerAddr:]

	return multiStakerAddr, valAddr
}

// // GetUBDKey creates the key for an unbonding delegation by delegator and validator addr
// // VALUE: multi-staking/MultiStakingUnlock
// func GetUBDKey(multiStakerAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
// 	return append(GetUBDsKey(delAddr.Bytes()), address.MustLengthPrefix(valAddr)...)
// }

func (l LockID) ToBytes() []byte {
	multiStakerAddr, valAcc, err := AccAddrAndValAddrFromStrings(l.MultiStakerAddr, l.ValAddr)
	if err != nil {
		panic(err)
	}

	lenMultiStakerAddr := len(multiStakerAddr)

	DVPair := make([]byte, 1+lenMultiStakerAddr+len(valAcc))

	DVPair[0] = uint8(lenMultiStakerAddr)

	copy(DVPair[1:], multiStakerAddr[:])

	copy(DVPair[1+lenMultiStakerAddr:], valAcc[:])

	return append(MultiStakingLockPrefix, DVPair...)
}

func (l UnlockID) ToBytes() []byte {
	multiStakerAddr, valAcc, err := AccAddrAndValAddrFromStrings(l.MultiStakerAddr, l.ValAddr)
	if err != nil {
		panic(err)
	}

	lenMultiStakerAddr := len(multiStakerAddr)

	DVPair := make([]byte, 1+lenMultiStakerAddr+len(valAcc))

	DVPair[0] = uint8(lenMultiStakerAddr)

	copy(DVPair[1:], multiStakerAddr[:])

	copy(DVPair[1+lenMultiStakerAddr:], valAcc[:])

	return append(MultiStakingUnlockPrefix, DVPair...)
}
