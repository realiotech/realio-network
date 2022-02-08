package rststaking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/rststaking/keeper"
	"github.com/realiotech/realio-network/x/rststaking/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the rstStake
	for _, elem := range genState.RstStakeList {
		k.SetRstStake(ctx, elem)
	}
	// this line is used by starport scaffolding # genesis/module/init
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.RstStakeList = k.GetAllRstStake(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
