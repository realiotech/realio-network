package multistaking

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	minttypes "github.com/realiotech/realio-network/x/mint/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"

	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	"github.com/spf13/cast"
)

var (
	bondedPoolAddress   = authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
	unbondedPoolAddress = authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName)
	multiStakingAddress = authtypes.NewModuleAddress(multistakingtypes.ModuleName)
	mintModuleAddress   = authtypes.NewModuleAddress(minttypes.ModuleName)
	newBondedCoinDenom  = "stake"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	appOpts servertypes.AppOptions,
	cdc codec.Codec,
	bk bankkeeper.Keeper,
	msk multistakingkeeper.Keeper,
	dk distrkeeper.Keeper,
	keys map[string]*storetypes.KVStoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade for multi staking...")

		nodeHome := cast.ToString(appOpts.Get(flags.FlagHome))
		upgradeGenFile := nodeHome + "/config/state.json"
		fmt.Println(upgradeGenFile)
		appState, _, err := genutiltypes.GenesisStateFromGenFile(upgradeGenFile)
		if err != nil {
			fmt.Println(err)
			panic("Unable to read genesis")
		}
		// migrate bank
		migrateBank(ctx, bk)

		// migrate distribute
		// delaccount to interdiate account
		//

		// migrate multistaking
		var multistakingGenesis = multistakingtypes.GenesisState{}
		err = cdc.UnmarshalJSON(appState["multi-staking"], &multistakingGenesis)
		if err != nil {
			fmt.Println("multistakingGenesis", err)
		}
		msk.InitGenesis(ctx, multistakingGenesis)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func migrateBank(ctx sdk.Context, bk bankkeeper.Keeper) {
	// Send coins from bonded pool add same amout to multistaking account
	bondedPoolBalances := bk.GetAllBalances(ctx, bondedPoolAddress)
	bk.SendCoins(ctx, bondedPoolAddress, multiStakingAddress, bondedPoolBalances)
	// mint stake to bonded pool
	bondedCoinsAmount := math.ZeroInt()
	for _, coinAmount := range bondedPoolBalances {
		bondedCoinsAmount = bondedCoinsAmount.Add(coinAmount.Amount)
	}
	amount := sdk.NewCoins(sdk.NewCoin(newBondedCoinDenom, bondedCoinsAmount))
	bk.MintCoins(ctx, minttypes.ModuleName, amount)
	bk.SendCoins(ctx, mintModuleAddress, bondedPoolAddress, amount)

	//----------------------//

	// Send coins from unbonded pool add same amout to multistaking account
	unbondedPoolBalances := bk.GetAllBalances(ctx, unbondedPoolAddress)
	bk.SendCoins(ctx, unbondedPoolAddress, multiStakingAddress, unbondedPoolBalances)
	// mint stake to unbonded pool
	unbondedCoinsAmount := math.ZeroInt()
	for _, coinAmount := range unbondedPoolBalances {
		unbondedCoinsAmount = unbondedCoinsAmount.Add(coinAmount.Amount)
	}
	amount = sdk.NewCoins(sdk.NewCoin(newBondedCoinDenom, unbondedCoinsAmount))
	bk.MintCoins(ctx, minttypes.ModuleName, amount)
	bk.SendCoins(ctx, mintModuleAddress, unbondedPoolAddress, amount)
}
