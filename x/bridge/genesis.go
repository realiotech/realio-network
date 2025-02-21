package bridge

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/bridge/keeper"
	"github.com/realiotech/realio-network/x/bridge/types"
)

// InitGenesis initializes the assets module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	err := k.Params.Set(ctx, genState.Params)
	if err != nil {
		panic(err)
	}
	err = k.EpochInfo.Set(ctx, genState.RatelimitEpochInfo)
	if err != nil {
		panic(err)
	}

	for _, coin := range genState.RegisteredCoins {
		err := k.RegisteredCoins.Set(ctx, coin.Coin.Denom, types.RateLimit{
			Ratelimit:     coin.Coin.Amount,
			CurrentInflow: math.ZeroInt(),
			Authority: coin.Authority,
		})
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

	epochInfo, err := k.EpochInfo.Get(ctx)
	if err != nil {
		panic(err)
	}
	genesis.RatelimitEpochInfo = epochInfo

	coins := []types.CoinAuthority{}
	err = k.RegisteredCoins.Walk(ctx, nil, func(denom string, ratelimit types.RateLimit) (stop bool, err error) {
		coins = append(coins, types.CoinAuthority{Coin: sdk.NewCoin(denom, ratelimit.Ratelimit), Authority: ratelimit.Authority})
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	genesis.RegisteredCoins = coins

	return genesis
}
