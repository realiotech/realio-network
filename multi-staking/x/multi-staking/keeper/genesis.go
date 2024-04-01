package keeper

import (
	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) (res []abci.ValidatorUpdate) {
	// multi-staking state
	for _, multiStakingLock := range data.MultiStakingLocks {
		k.SetMultiStakingLock(ctx, multiStakingLock)
	}
	for _, multiStakingUnlock := range data.MultiStakingUnlocks {
		k.SetMultiStakingUnlock(ctx, multiStakingUnlock)
	}
	for _, multiStakingCoinInfo := range data.MultiStakingCoinInfo {
		k.SetBondWeight(ctx, multiStakingCoinInfo.Denom, multiStakingCoinInfo.BondWeight)
	}

	for _, valMultiStakingCoin := range data.ValidatorMultiStakingCoins {
		valAddr, err := sdk.ValAddressFromBech32(valMultiStakingCoin.ValAddr)
		if err != nil {
			panic("error validator address")
		}
		k.SetValidatorMultiStakingCoin(ctx, valAddr, valMultiStakingCoin.CoinDenom)
	}

	k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)

	return k.stakingKeeper.InitGenesis(ctx, &data.StakingGenesisState)
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// get multiStakingLock
	var multiStakingLocks []types.MultiStakingLock
	k.MultiStakingLockIterator(ctx, func(stakingLock types.MultiStakingLock) bool {
		multiStakingLocks = append(multiStakingLocks, stakingLock)
		return false
	})

	var multiStakingUnlocks []types.MultiStakingUnlock
	k.MultiStakingUnlockIterator(ctx, func(unlock types.MultiStakingUnlock) bool {
		multiStakingUnlocks = append(multiStakingUnlocks, unlock)
		return false
	})

	var multiStakingCoinInfos []types.MultiStakingCoinInfo
	k.BondWeightIterator(ctx, func(denom string, bondWeight sdk.Dec) bool {
		multiStakingCoinInfos = append(multiStakingCoinInfos, types.MultiStakingCoinInfo{
			Denom:      denom,
			BondWeight: bondWeight,
		})
		return false
	})

	// get validator allowed coin
	var ValidatorMultiStakingCoinLists []types.ValidatorMultiStakingCoin
	k.ValidatorMultiStakingCoinIterator(ctx, func(valAddr string, denom string) (stop bool) {
		ValidatorMultiStakingCoin := types.ValidatorMultiStakingCoin{
			ValAddr:   valAddr,
			CoinDenom: denom,
		}
		ValidatorMultiStakingCoinLists = append(ValidatorMultiStakingCoinLists, ValidatorMultiStakingCoin)
		return false
	})

	return &types.GenesisState{
		MultiStakingLocks:          multiStakingLocks,
		MultiStakingUnlocks:        multiStakingUnlocks,
		MultiStakingCoinInfo:       multiStakingCoinInfos,
		ValidatorMultiStakingCoins: ValidatorMultiStakingCoinLists,
		StakingGenesisState:        *k.stakingKeeper.ExportGenesis(ctx),
	}
}
