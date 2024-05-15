package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ck consensuskeeper.Keeper,
	clientKeeper ibctmmigrations.ClientKeeper,
	pk paramskeeper.Keeper,
	sk *stakingkeeper.Keeper,
	cdc codec.BinaryCodec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade for multi staking...")

		migrateParamSubspace(ctx, ck, pk)
		// fixMinCommisionRate(ctx, sk)

		if _, err := ibctmmigrations.PruneExpiredConsensusStates(ctx, cdc, clientKeeper); err != nil {
			return nil, err
		}
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
