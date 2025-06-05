package network

import (
	"math/big"

	"github.com/cosmos/evm/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	testconstants "github.com/realiotech/realio-network/testutil/integration/constants"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type InitialAmounts struct {
	Base math.Int
	Evm  math.Int
}

func DefaultInitialAmounts() InitialAmounts {
	baseCoinInfo := testconstants.ExampleChainCoinInfo[defaultChain]

	return InitialAmounts{
		Base: GetInitialAmount(baseCoinInfo.Decimals),
		Evm:  GetInitialAmount(baseCoinInfo.Decimals),
	}
}

func DefaultInitialBondedAmount() math.Int {
	baseCoinInfo := testconstants.ExampleChainCoinInfo[defaultChain]

	return GetInitialBondedAmount(baseCoinInfo.Decimals)
}

func GetInitialAmount(decimals evmtypes.Decimals) math.Int {
	if err := decimals.Validate(); err != nil {
		panic("unsupported decimals")
	}

	// initialBalance defines the initial balance represented in 18 decimals.
	initialBalance, _ := math.NewIntFromString("100_000_000_000_000_000_000_000")

	// 18 decimals is the most precise representation we can have, for this
	// reason we have to divide the initial balance by the decimals value to
	// have the specific representation.
	return initialBalance.Quo(decimals.ConversionFactor())
}

func GetInitialBondedAmount(decimals evmtypes.Decimals) math.Int {
	if err := decimals.Validate(); err != nil {
		panic("unsupported decimals")
	}

	// initialBondedAmount represents the amount of tokens that each validator will
	// have initially bonded expressed in the 18 decimals representation.
	sdk.DefaultPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals.ConversionFactor().Int64()), nil))
	initialBondedAmount := sdk.TokensFromConsensusPower(1, types.AttoPowerReduction)

	return initialBondedAmount.Quo(decimals.ConversionFactor())
}

func GetInitialBaseFeeAmount(decimals evmtypes.Decimals) math.LegacyDec {
	if err := decimals.Validate(); err != nil {
		panic("unsupported decimals")
	}

	baseFee := math.LegacyNewDec(1_000_000_000)
	baseFee = baseFee.Quo(decimals.ConversionFactor().ToLegacyDec())
	return baseFee
}
