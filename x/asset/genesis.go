package asset

import (
	"context"

	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
)

// InitGenesis initializes the assets module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	err := k.Params.Set(ctx, genState.Params)
	if err != nil {
		panic(err)
	}
	for _, token := range genState.Tokens {
		err := k.Token.Set(ctx, types.TokenKey(token.Symbol), token)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}
	genesis.Params = params
	tokens := []types.Token{}
	err = k.Token.Walk(ctx, nil, func(_ string, token types.Token) (stop bool, err error) {
		tokens = append(tokens, token)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	genesis.Tokens = tokens
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
