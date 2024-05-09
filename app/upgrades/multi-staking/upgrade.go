package multistaking

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	multistaking "github.com/realio-tech/multi-staking-module/x/multi-staking"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/realiotech/realio-network/app/upgrades/multi-staking/legacy"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
	"github.com/spf13/cast"
)

var (
	bondedPoolAddress   = authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
	unbondedPoolAddress = authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName)
	multiStakingAddress = authtypes.NewModuleAddress(multistakingtypes.ModuleName)
	mintModuleAddress   = authtypes.NewModuleAddress(minttypes.ModuleName)
	newBondedCoinDenom  = "stake"
	msCoinDenomLists    = []string{"ario", "arst"}
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	appOpts servertypes.AppOptions,
	cdc codec.Codec,
	bk bankkeeper.Keeper,
	ak authkeeper.AccountKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade for multi staking...")

		nodeHome := cast.ToString(appOpts.Get(flags.FlagHome))
		upgradeGenFile := nodeHome + "/config/state.json"
		appState, _, err := genutiltypes.GenesisStateFromGenFile(upgradeGenFile)
		if err != nil {
			panic(err)
		}

		// migrate multistaking
		appState, err = migrateMultiStaking(appState)
		if err != nil {
			panic(err)
		}

		// validate genesis
		var genesisState multistakingtypes.GenesisState
		cdc.MustUnmarshalJSON(appState[multistakingtypes.ModuleName], &genesisState)
		err = genesisState.Validate()
		if err != nil {
			panic(err)
		}

		ak.GetModuleAccount(ctx, multistakingtypes.ModuleName)

		// migrate bank
		migrateBank(ctx, bk)

		vm[multistakingtypes.ModuleName] = multistaking.AppModule{}.ConsensusVersion()
		msModule := mm.Modules[multistakingtypes.ModuleName].(module.HasGenesis)
		msModule.InitGenesis(ctx, cdc, appState[multistakingtypes.ModuleName])

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func migrateBank(ctx sdk.Context, bk bankkeeper.Keeper) {
	// convert bonded pool balances and send MsCoins to multi-staking module account
	convertStakingModulePoolBalances(ctx, bk, bondedPoolAddress)

	// convert unbonded pool balances and send MsCoins to multi-staking module account
	convertStakingModulePoolBalances(ctx, bk, unbondedPoolAddress)
}

func convertStakingModulePoolBalances(ctx sdk.Context, bk bankkeeper.Keeper, accAddr sdk.AccAddress) {
	// Send coins from accAddr add same amout to multistaking account
	accountBalances := bk.GetAllBalances(ctx, accAddr)
	err := bk.SendCoins(ctx, accAddr, multiStakingAddress, accountBalances)
	if err != nil {
		panic(err)
	}
	// mint stake to bonded pool
	convertedCoinsAmount := math.ZeroInt()
	for _, coinAmount := range accountBalances {
		convertedCoinsAmount = convertedCoinsAmount.Add(coinAmount.Amount)
	}
	amount := sdk.NewCoins(sdk.NewCoin(newBondedCoinDenom, convertedCoinsAmount))
	err = bk.MintCoins(ctx, minttypes.ModuleName, amount)
	if err != nil {
		panic(err)
	}
	err = bk.SendCoins(ctx, mintModuleAddress, accAddr, amount)
	if err != nil {
		panic(err)
	}
}

func migrateMultiStaking(appState map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	var oldState legacy.GenesisState
	err := json.Unmarshal(appState[stakingtypes.ModuleName], &oldState)
	if err != nil {
		return nil, err
	}

	newState := MultiStakingGenesisState{}
	// Migrate state.StakingGenesisState
	stakingGenesisState := GenesisState{}

	stakingGenesisState.Params = Params{
		UnbondingTime:     oldState.Params.UnbondingTime,
		MaxValidators:     oldState.Params.MaxValidators,
		MaxEntries:        oldState.Params.MaxEntries,
		HistoricalEntries: oldState.Params.HistoricalEntries,
		BondDenom:         newBondedCoinDenom,
		MinCommissionRate: oldState.Params.MinCommissionRate,
	}
	stakingGenesisState.LastTotalPower = oldState.LastTotalPower
	stakingGenesisState.Validators = convertValidators(oldState.Validators)
	stakingGenesisState.Delegations = convertDelegations(oldState.Delegations)
	stakingGenesisState.UnbondingDelegations = convertUnbondings(oldState.UnbondingDelegations)
	stakingGenesisState.Redelegations = convertRedelegations(oldState.Redelegations)
	stakingGenesisState.Exported = oldState.Exported

	newState.StakingGenesisState = stakingGenesisState

	// Migrate state.ValidatorAllowedToken field
	newState.ValidatorMultiStakingCoins = make([]multistakingtypes.ValidatorMultiStakingCoin, 0)

	for _, val := range oldState.Validators {
		allowedToken := multistakingtypes.ValidatorMultiStakingCoin{
			ValAddr:   val.OperatorAddress,
			CoinDenom: val.BondDenom,
		}
		newState.ValidatorMultiStakingCoins = append(newState.ValidatorMultiStakingCoins, allowedToken)
	}

	// Migrate state.MultiStakingLock field
	newState.MultiStakingLocks = make([]multistakingtypes.MultiStakingLock, 0)

	for _, val := range oldState.Validators {
		for _, del := range oldState.Delegations {
			if del.ValidatorAddress == val.OperatorAddress {
				val, amount := tokenAmountFromShares(val, del.Shares)
				lock := multistakingtypes.MultiStakingLock{
					LockID: multistakingtypes.LockID{
						MultiStakerAddr: del.DelegatorAddress,
						ValAddr:         del.ValidatorAddress,
					},
					LockedCoin: multistakingtypes.MultiStakingCoin{
						Denom:      val.BondDenom,
						Amount:     amount,
						BondWeight: sdk.OneDec(),
					},
				}
				newState.MultiStakingLocks = append(newState.MultiStakingLocks, lock)
			}
		}
	}

	newState.MultiStakingCoinInfo = make([]multistakingtypes.MultiStakingCoinInfo, 0)
	for _, msCoinDenom := range msCoinDenomLists {
		msCoinInfo := multistakingtypes.MultiStakingCoinInfo{
			Denom:      msCoinDenom,
			BondWeight: sdk.OneDec(),
		}
		newState.MultiStakingCoinInfo = append(newState.MultiStakingCoinInfo, msCoinInfo)
	}

	// Migrate state.MultiStakingUnlock field
	newState.MultiStakingUnlocks = make([]multistakingtypes.MultiStakingUnlock, 0)
	for _, ubd := range oldState.UnbondingDelegations {
		unlockID := multistakingtypes.UnlockID{
			MultiStakerAddr: ubd.DelegatorAddress,
			ValAddr:         ubd.ValidatorAddress,
		}

		unlockEntries := make([]multistakingtypes.UnlockEntry, 0)
		for _, entry := range ubd.Entries {
			unlockEntry := multistakingtypes.UnlockEntry{
				CreationHeight: entry.CreationHeight,
				UnlockingCoin: multistakingtypes.MultiStakingCoin{
					Denom:      entry.InitialBalance.Denom,
					Amount:     entry.InitialBalance.Amount,
					BondWeight: sdk.OneDec(),
				},
			}
			unlockEntries = append(unlockEntries, unlockEntry)
		}

		msUnLock := multistakingtypes.MultiStakingUnlock{
			UnlockID: unlockID,
			Entries:  unlockEntries,
		}

		newState.MultiStakingUnlocks = append(newState.MultiStakingUnlocks, msUnLock)
	}

	newData, err := json.Marshal(&newState)
	if err != nil {
		return nil, err
	}

	appState[multistakingtypes.ModuleName] = newData

	return appState, nil
}

func tokenAmountFromShares(v legacy.Validator, delShares sdk.Dec) (legacy.Validator, math.Int) {
	remainingShares := v.DelegatorShares.Sub(delShares)

	var amount math.Int
	if remainingShares.IsZero() {
		// last delegation share gets any trimmings
		amount = v.Tokens
		v.Tokens = math.ZeroInt()
	} else {
		// leave excess tokens in the validator
		// however fully use all the delegator shares
		amount = tokensFromShares(v, delShares).TruncateInt()
		v.Tokens = v.Tokens.Sub(amount)

		if v.Tokens.IsNegative() {
			panic("attempting to remove more tokens than available in validator")
		}
	}

	v.DelegatorShares = remainingShares

	return v, amount
}

func tokensFromShares(v legacy.Validator, shares sdk.Dec) sdk.Dec {
	return (shares.MulInt(v.Tokens)).Quo(v.DelegatorShares)
}

func convertValidators(validators []legacy.Validator) []Validator {
	newValidators := make([]Validator, 0)
	for _, val := range validators {
		date, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", "2023-06-20T11:54:21.351285642Z")
		fmt.Println("time", val.Commission.UpdateTime.String(), err, date)
		newVal := Validator{
			OperatorAddress: val.OperatorAddress,
			ConsensusPubkey: val.ConsensusPubkey,
			Jailed:          val.Jailed,
			Status:          val.Status,
			Tokens:          val.Tokens,
			DelegatorShares: val.DelegatorShares,
			Description:     stakingtypes.Description(val.Description),
			UnbondingHeight: val.UnbondingHeight,
			UnbondingTime:   val.UnbondingTime,
			Commission: stakingtypes.Commission{
				CommissionRates: stakingtypes.CommissionRates(val.Commission.CommissionRates),
				UpdateTime:      val.Commission.UpdateTime,
			},
			MinSelfDelegation: val.MinSelfDelegation,
		}
		newValidators = append(newValidators, newVal)
	}
	return newValidators
}

func convertDelegations(delegations []legacy.Delegation) []stakingtypes.Delegation {
	newDelegations := make([]stakingtypes.Delegation, 0)
	for _, del := range delegations {
		newDel := stakingtypes.Delegation{
			DelegatorAddress: del.DelegatorAddress,
			ValidatorAddress: del.ValidatorAddress,
			Shares:           del.Shares,
		}

		newDelegations = append(newDelegations, newDel)
	}
	return newDelegations
}

func convertUnbondings(ubds []legacy.UnbondingDelegation) []stakingtypes.UnbondingDelegation {
	newUbds := make([]stakingtypes.UnbondingDelegation, 0)
	for _, ubd := range ubds {
		newEntries := make([]stakingtypes.UnbondingDelegationEntry, 0)
		for _, entry := range ubd.Entries {
			newEntry := stakingtypes.UnbondingDelegationEntry{
				CreationHeight: entry.CreationHeight,
				CompletionTime: entry.CompletionTime,
				InitialBalance: entry.InitialBalance.Amount,
				Balance:        entry.Balance.Amount,
			}
			newEntries = append(newEntries, newEntry)
		}
		newUbd := stakingtypes.UnbondingDelegation{
			DelegatorAddress: ubd.DelegatorAddress,
			ValidatorAddress: ubd.ValidatorAddress,
			Entries:          newEntries,
		}
		newUbds = append(newUbds, newUbd)
	}
	return newUbds
}

func convertRedelegations(reDels []legacy.Redelegation) []stakingtypes.Redelegation {
	newRedels := make([]stakingtypes.Redelegation, 0)
	for _, reDel := range reDels {
		newEntries := make([]stakingtypes.RedelegationEntry, 0)
		for _, entry := range reDel.Entries {
			newEntry := stakingtypes.RedelegationEntry{
				CreationHeight: entry.CreationHeight,
				CompletionTime: entry.CompletionTime,
				InitialBalance: entry.InitialBalance.Amount,
				SharesDst:      entry.SharesDst,
			}
			newEntries = append(newEntries, newEntry)
		}
		newRedel := stakingtypes.Redelegation{
			DelegatorAddress:    reDel.DelegatorAddress,
			ValidatorSrcAddress: reDel.ValidatorSrcAddress,
			ValidatorDstAddress: reDel.ValidatorDstAddress,
			Entries:             newEntries,
		}
		newRedels = append(newRedels, newRedel)
	}
	return newRedels
}
