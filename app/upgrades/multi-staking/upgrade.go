package multistaking

import (
	"encoding/json"
	"fmt"
	"time"

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
	multistaking "github.com/realio-tech/multi-staking-module/x/multi-staking"
	minttypes "github.com/realiotech/realio-network/x/mint/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"

	"cosmossdk.io/math"
	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/realiotech/realio-network/app/upgrades/multi-staking/legacy"
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
		//

		// migrate multistaking
		appState, err = migrateMultiStaking(appState)

		fmt.Println()
		fmt.Println("=============UpgradeHandler=============")
		fmt.Printf("%s", appState[multistakingtypes.ModuleName])
		fmt.Println("=============UpgradeHandler=============")

		if err != nil {
			panic(err)
		}
		vm[multistakingtypes.ModuleName] = multistaking.AppModule{}.ConsensusVersion()
		mm.Modules[multistakingtypes.ModuleName].InitGenesis(ctx, cdc, appState[multistakingtypes.ModuleName])

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

type Params struct {
	// unbonding_time is the time duration of unbonding.
	UnbondingTime string `protobuf:"bytes,1,opt,name=unbonding_time,json=unbondingTime,proto3,stdduration" json:"unbonding_time"`
	// max_validators is the maximum number of validators.
	MaxValidators uint32 `protobuf:"varint,2,opt,name=max_validators,json=maxValidators,proto3" json:"max_validators,omitempty"`
	// max_entries is the max entries for either unbonding delegation or redelegation (per pair/trio).
	MaxEntries uint32 `protobuf:"varint,3,opt,name=max_entries,json=maxEntries,proto3" json:"max_entries,omitempty"`
	// historical_entries is the number of historical entries to persist.
	HistoricalEntries uint32 `protobuf:"varint,4,opt,name=historical_entries,json=historicalEntries,proto3" json:"historical_entries,omitempty"`
	// bond_denom defines the bondable coin denomination.
	BondDenom string `protobuf:"bytes,5,opt,name=bond_denom,json=bondDenom,proto3" json:"bond_denom,omitempty"`
	// min_commission_rate is the chain-wide minimum commission rate that a validator can charge their delegators
	MinCommissionRate sdk.Dec `protobuf:"bytes,6,opt,name=min_commission_rate,json=minCommissionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"min_commission_rate" yaml:"min_commission_rate"`
}

type Validator struct {
	// operator_address defines the address of the validator's operator; bech encoded in JSON.
	OperatorAddress string `protobuf:"bytes,1,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`
	// consensus_pubkey is the consensus public key of the validator, as a Protobuf Any.
	ConsensusPubkey *types1.Any `protobuf:"bytes,2,opt,name=consensus_pubkey,json=consensusPubkey,proto3" json:"consensus_pubkey,omitempty"`
	// jailed defined whether the validator has been jailed from bonded status or not.
	Jailed bool `protobuf:"varint,3,opt,name=jailed,proto3" json:"jailed,omitempty"`
	// status is the validator status (bonded/unbonding/unbonded).
	Status string `protobuf:"varint,4,opt,name=status,proto3,enum=cosmos.staking.v1beta1.BondStatus" json:"status,omitempty"`
	// tokens define the delegated tokens (incl. self-delegation).
	Tokens sdk.Int `protobuf:"bytes,5,opt,name=tokens,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"tokens"`
	// delegator_shares defines total shares issued to a validator's delegators.
	DelegatorShares sdk.Dec `protobuf:"bytes,6,opt,name=delegator_shares,json=delegatorShares,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"delegator_shares"`
	// description defines the description terms for the validator.
	Description stakingtypes.Description `protobuf:"bytes,7,opt,name=description,proto3" json:"description"`
	// unbonding_height defines, if unbonding, the height at which this validator has begun unbonding.
	UnbondingHeight int64 `protobuf:"varint,8,opt,name=unbonding_height,json=unbondingHeight,proto3" json:"unbonding_height,omitempty"`
	// unbonding_time defines, if unbonding, the min time for the validator to complete unbonding.
	UnbondingTime time.Time `protobuf:"bytes,9,opt,name=unbonding_time,json=unbondingTime,proto3,stdtime" json:"unbonding_time"`
	// commission defines the commission parameters.
	Commission stakingtypes.Commission `protobuf:"bytes,10,opt,name=commission,proto3" json:"commission"`
	// min_self_delegation is the validator's self declared minimum self delegation.
	//
	// Since: cosmos-sdk 0.46
	MinSelfDelegation sdk.Int `protobuf:"bytes,11,opt,name=min_self_delegation,json=minSelfDelegation,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"min_self_delegation"`
}

type GenesisState struct {
	// params defines all the paramaters of related to deposit.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// last_total_power tracks the total amounts of bonded tokens recorded during
	// the previous end block.
	LastTotalPower sdk.Int `protobuf:"bytes,2,opt,name=last_total_power,json=lastTotalPower,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"last_total_power"`
	// last_validator_powers is a special index that provides a historical list
	// of the last-block's bonded validators.
	LastValidatorPowers []stakingtypes.LastValidatorPower `protobuf:"bytes,3,rep,name=last_validator_powers,json=lastValidatorPowers,proto3" json:"last_validator_powers"`
	// delegations defines the validator set at genesis.
	Validators []Validator `protobuf:"bytes,4,rep,name=validators,proto3" json:"validators"`
	// delegations defines the delegations active at genesis.
	Delegations []stakingtypes.Delegation `protobuf:"bytes,5,rep,name=delegations,proto3" json:"delegations"`
	// unbonding_delegations defines the unbonding delegations active at genesis.
	UnbondingDelegations []stakingtypes.UnbondingDelegation `protobuf:"bytes,6,rep,name=unbonding_delegations,json=unbondingDelegations,proto3" json:"unbonding_delegations"`
	// redelegations defines the redelegations active at genesis.
	Redelegations []stakingtypes.Redelegation `protobuf:"bytes,7,rep,name=redelegations,proto3" json:"redelegations"`
	Exported      bool                        `protobuf:"varint,8,opt,name=exported,proto3" json:"exported,omitempty"`
}

type MultiStakingGenesisState struct {
	MultiStakingLocks          []multistakingtypes.MultiStakingLock          `protobuf:"bytes,1,rep,name=multi_staking_locks,json=multiStakingLocks,proto3" json:"multi_staking_locks"`
	MultiStakingUnlocks        []multistakingtypes.MultiStakingUnlock        `protobuf:"bytes,2,rep,name=multi_staking_unlocks,json=multiStakingUnlocks,proto3" json:"multi_staking_unlocks"`
	MultiStakingCoinInfo       []multistakingtypes.MultiStakingCoinInfo      `protobuf:"bytes,3,rep,name=multi_staking_coin_info,json=multiStakingCoinInfo,proto3" json:"multi_staking_coin_info"`
	ValidatorMultiStakingCoins []multistakingtypes.ValidatorMultiStakingCoin `protobuf:"bytes,4,rep,name=validator_multi_staking_coins,json=validatorMultiStakingCoins,proto3" json:"validator_multi_staking_coins"`
	IntermediaryDelegators     []string                                      `protobuf:"bytes,5,rep,name=IntermediaryDelegators,proto3" json:"IntermediaryDelegators,omitempty"`
	StakingGenesisState        GenesisState                                  `protobuf:"bytes,6,opt,name=staking_genesis_state,json=stakingGenesisState,proto3" json:"staking_genesis_state"`
}

func migrateMultiStaking(appState map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	var oldState legacy.GenesisState
	err := json.Unmarshal(appState["staking"], &oldState)
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
		BondDenom:         "stake",
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
					LockID: &multistakingtypes.LockID{
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
		newDel := stakingtypes.Delegation(del)
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
