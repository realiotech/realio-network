package bank

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// BalancesMethod defines the ABI method name for the bank Balances
	// query.
	BalancesMethod = "balances"
	// TotalSupplyMethod defines the ABI method name for the bank TotalSupply
	// query.
	TotalSupplyMethod = "totalSupply"
	// SupplyOfMethod defines the ABI method name for the bank SupplyOf
	// query.
	SupplyOfMethod = "supplyOf"
)

// Balances returns given account's balances of all tokens registered in the x/bank module
// and the corresponding ERC20 address (address, amount). The amount returned for each token
// has the original decimals precision stored in the x/bank.
// This method charges the account the corresponding value of an ERC-20
// balanceOf call for each token returned.
func (p Precompile) Balances(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	account, err := ParseBalancesArgs(args)
	if err != nil {
		return nil, fmt.Errorf("error calling account balances in bank precompile: %s", err)
	}

	bals := p.bankKeeper.GetAllBalances(ctx, account)

	balances := make([]Balance, 0)
	for _, bal := range bals {
		balances = append(balances, Balance{
			Denom:  bal.Denom,
			Amount: bal.Amount.BigInt(),
		})
	}

	return method.Outputs.Pack(balances)
}

// TotalSupply returns the total supply of all tokens registered in the x/bank
// module. The amount returned for each token has the original
// decimals precision stored in the x/bank.
// This method charges the account the corresponding value of a ERC-20 totalSupply
// call for each token returned.
func (p Precompile) TotalSupply(
	ctx sdk.Context,
	method *abi.Method,
	_ []interface{},
) ([]byte, error) {
	i := 0
	totalSupply := make([]Balance, 0)

	p.bankKeeper.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
		defer func() { i++ }()

		totalSupply = append(totalSupply, Balance{
			Denom:  coin.Denom,
			Amount: coin.Amount.BigInt(),
		})
		return false
	})

	return method.Outputs.Pack(totalSupply)
}

// SupplyOf returns the total supply of a given registered erc20 token
// from the x/bank module. If the ERC20 token doesn't have a registered
// TokenPair, the method returns a supply of zero.
// The amount returned with this query has the original decimals precision
// stored in the x/bank.
func (p Precompile) SupplyOf(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	denom, err := ParseSupplyOfArgs(args)
	if err != nil {
		return nil, fmt.Errorf("error getting the supply in bank precompile: %s", err)
	}

	supply := p.bankKeeper.GetSupply(ctx, denom)

	return method.Outputs.Pack(supply.Amount.BigInt())
}
