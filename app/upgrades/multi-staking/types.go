package multistaking

import (
	"time"

	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
)

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
	StakingGenesisState        GenesisState                                  `protobuf:"bytes,6,opt,name=staking_genesis_state,json=stakingGenesisState,proto3" json:"staking_genesis_state"`
}
