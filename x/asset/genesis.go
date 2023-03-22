package asset

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
)

// InitGenesis initializes the assets module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	//k.SetTransferRestrictionFn()
	k.SetParams(ctx, genState.Params)
	for _, token := range genState.Tokens {
		k.SetToken(ctx, token)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.Tokens = k.GetAllToken(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
