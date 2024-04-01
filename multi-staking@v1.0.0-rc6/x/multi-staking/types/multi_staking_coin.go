package types

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (coin MultiStakingCoin) Validate() error {
	if coin.Amount.IsNegative() {
		return fmt.Errorf("weight zero or negative")
	}
	return nil
}

func (coin MultiStakingCoin) ToCoin() sdk.Coin {
	return sdk.NewCoin(coin.Denom, coin.Amount)
}

func NewMultiStakingCoin(denom string, amount math.Int, weight sdk.Dec) MultiStakingCoin {
	return MultiStakingCoin{Denom: denom, Amount: amount, BondWeight: weight}
}

func (coin MultiStakingCoin) BondValue() math.Int {
	return coin.BondWeight.MulInt(coin.Amount).TruncateInt()
}

func (coin MultiStakingCoin) WithAmount(amount math.Int) MultiStakingCoin {
	return NewMultiStakingCoin(coin.Denom, amount, coin.BondWeight)
}

func (coin MultiStakingCoin) SafeSub(coinB MultiStakingCoin) (MultiStakingCoin, error) {
	if coin.Denom != coinB.Denom {
		return MultiStakingCoin{}, fmt.Errorf("denom mismatch")
	}

	coin.Amount = coin.Amount.Sub(coinB.Amount)
	if coin.Amount.IsNegative() {
		return MultiStakingCoin{}, fmt.Errorf("insufficient amount")
	}

	return coin, nil
}

func (coinA MultiStakingCoin) SafeAdd(coinB MultiStakingCoin) (MultiStakingCoin, error) {
	if coinA.Amount.IsZero() {
		return coinB, nil
	}

	if coinA.Denom != coinB.Denom {
		return MultiStakingCoin{}, fmt.Errorf("denom mismatch")
	}

	amountA := coinA.Amount
	weightA := coinA.BondWeight

	amountB := coinB.Amount
	weightB := coinB.BondWeight

	// amountAB = amountA + amountB
	amountAB := amountA.Add(amountB)
	// weightAB = (weightA * amountA + weightB * amountB) / (amountA + amountB)
	weightAB := ((weightA.MulInt(amountA)).Add(weightB.MulInt(amountB))).QuoInt(amountAB)

	return NewMultiStakingCoin(coinA.Denom, amountAB, weightAB), nil
}

func (coin MultiStakingCoin) Add(coinB MultiStakingCoin) MultiStakingCoin {
	coinAB, err := coin.SafeAdd(coinB)
	if err != nil {
		panic(err)
	}

	return coinAB
}
