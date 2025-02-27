package erc20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/erc20/keeper"
	"github.com/realiotech/realio-network/x/erc20/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Erc20Keeper,
	data types.GenesisState,
) error {
	for _, token := range data.TokenOwners {
		err := k.SetContractOwner(ctx, token.ContractAddress, token.OwnerAddress)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis export module status
func ExportGenesis(ctx sdk.Context, k keeper.Erc20Keeper) *types.GenesisState {
	return &types.GenesisState{
		TokenOwners: k.GetContractOwners(ctx),
	}
}
