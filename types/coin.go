package types

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// AttoRio defines the default coin denomination used in RealioNetwork in:
	//
	// - Staking parameters: denomination used as stake in the dPoS chain
	// - Mint parameters: denomination minted due to fee distribution rewards
	// - Governance parameters: denomination used for spam prevention in proposal deposits
	// - Crisis parameters: constant fee denomination used for spam prevention to check broken invariant
	// - EVM parameters: denomination used for running EVM state transitions in RealioNetwork.
	AttoRio string = "ario"

	// BaseDenomUnit defines the base denomination unit for RealioNetwork.
	// 1 rio = 1x10^{BaseDenomUnit} ario
	BaseDenomUnit = 18

	// DefaultGasPrice is default gas price for evm transactions
	DefaultGasPrice = 20
)

// PowerReduction defines the default power reduction value for staking
var PowerReduction = sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(BaseDenomUnit), nil))

// NewRioCoin is a utility function that returns an "ario" coin with the given sdkmath.Int amount.
// The function will panic if the provided amount is negative.
func NewRioCoin(amount sdkmath.Int) sdk.Coin {
	return sdk.NewCoin(AttoRio, amount)
}

// NewRioDecCoin is a utility function that returns an "ario" decimal coin with the given sdkmath.Int amount.
// The function will panic if the provided amount is negative.
func NewRioDecCoin(amount sdkmath.Int) sdk.DecCoin {
	return sdk.NewDecCoin(AttoRio, amount)
}

// NewRioCoinInt64 is a utility function that returns an "ario" coin with the given int64 amount.
// The function will panic if the provided amount is negative.
func NewRioCoinInt64(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(AttoRio, amount)
}
