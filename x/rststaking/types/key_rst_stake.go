package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// RstStakeKeyPrefix is the prefix to retrieve all RstStake
	RstStakeKeyPrefix = "RstStake/value/"
)

// RstStakeKey returns the store key to retrieve a RstStake from the index fields
func RstStakeKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
