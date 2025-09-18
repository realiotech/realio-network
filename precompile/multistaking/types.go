// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package multistaking

import (
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	cmn "github.com/cosmos/evm/precompiles/common"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
)

const (
	MultistakingPrecompileAddress = "0x0000000000000000000000000000000000000900"
	// Transactions
	DelegateMethod                  = "delegate"
	UndelegateMethod                = "undelegate"
	RedelegateMethod                = "redelegate"
	CancelUnbondingDelegationMethod = "cancelUnbondingDelegation"
	CreateValidatorMethod           = "createValidator"
)

func NewDelegationRequest(args []interface{}) (*multistakingtypes.QueryMultiStakingLockRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorAddress, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "validatorAddress", "string", args[1])
	}

	return &multistakingtypes.QueryMultiStakingLockRequest{
		MultiStakerAddress: sdk.AccAddress(delegatorAddr.Bytes()).String(), // bech32 formatted
		ValidatorAddress:   validatorAddress,
	}, nil
}

// DelegationOutput is a struct to represent the key information from
// a delegation response.
type DelegationOutput struct {
	Balance cmn.Coin
}

// FromResponse populates the DelegationOutput from a QueryDelegationResponse.
func (do *DelegationOutput) FromResponse(res *multistakingtypes.QueryMultiStakingLockResponse) *DelegationOutput {
	do.Balance = cmn.Coin{
		Denom:  res.Lock.LockedCoin.Denom,
		Amount: res.Lock.LockedCoin.Amount.BigInt(),
	}
	return do
}

// Pack packs a given slice of abi arguments into a byte array.
func (do *DelegationOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(do.Balance)
}

func NewUnbondingDelegationRequest(args []interface{}) (*multistakingtypes.QueryMultiStakingUnlockRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorAddress, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "validatorAddress", "string", args[1])
	}

	return &multistakingtypes.QueryMultiStakingUnlockRequest{
		MultiStakerAddress: sdk.AccAddress(delegatorAddr.Bytes()).String(), // bech32 formatted
		ValidatorAddress:   validatorAddress,
	}, nil
}

// UnbondingDelegationEntry is a struct that contains the information about an unbonding delegation entry.
type UnbondingDelegationEntry struct {
	CreationHeight int64
	Balance        *big.Int
}

// UnbondingDelegationResponse is a struct that contains the information about an unbonding delegation.
type UnbondingDelegationResponse struct {
	DelegatorAddress string
	ValidatorAddress string
	Entries          []UnbondingDelegationEntry
}

// UnbondingDelegationOutput is the output response returned by the query method.
type UnbondingDelegationOutput struct {
	UnbondingDelegation UnbondingDelegationResponse
}

// FromResponse populates the DelegationOutput from a QueryDelegationResponse.
func (do *UnbondingDelegationOutput) FromResponse(res *multistakingtypes.QueryMultiStakingUnlockResponse) *UnbondingDelegationOutput {
	do.UnbondingDelegation.Entries = make([]UnbondingDelegationEntry, len(res.Unlock.Entries))
	do.UnbondingDelegation.ValidatorAddress = res.Unlock.UnlockID.ValAddr
	do.UnbondingDelegation.DelegatorAddress = res.Unlock.UnlockID.MultiStakerAddr
	for i, entry := range res.Unlock.Entries {
		do.UnbondingDelegation.Entries[i] = UnbondingDelegationEntry{
			CreationHeight: entry.CreationHeight,
			Balance:        entry.UnlockingCoin.Amount.BigInt(),
		}
	}
	return do
}

func NewValidatorRequest(args []interface{}) (*multistakingtypes.QueryValidatorRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	validatorHexAddr, ok := args[0].(common.Address)
	if !ok || validatorHexAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidValidator, args[0])
	}

	validatorAddress := sdk.ValAddress(validatorHexAddr.Bytes()).String()

	return &multistakingtypes.QueryValidatorRequest{ValidatorAddr: validatorAddress}, nil
}

// Description use golang type alias defines a validator description.
type Description = struct {
	Moniker         string "json:\"moniker\""
	Identity        string "json:\"identity\""
	Website         string "json:\"website\""
	SecurityContact string "json:\"securityContact\""
	Details         string "json:\"details\""
}

// ValidatorInfo is a struct to represent the key information from
// a validator response.
type ValidatorInfo struct {
	OperatorAddress   string   `abi:"operatorAddress"`
	ConsensusPubkey   string   `abi:"consensusPubkey"`
	Jailed            bool     `abi:"jailed"`
	Status            uint8    `abi:"status"`
	Tokens            *big.Int `abi:"tokens"`
	DelegatorShares   *big.Int `abi:"delegatorShares"` // TODO: Decimal
	Description       stakingtypes.Description   `abi:"description"`
	UnbondingHeight   int64    `abi:"unbondingHeight"`
	UnbondingTime     int64    `abi:"unbondingTime"`
	Commission        *big.Int `abi:"commission"`
	MinSelfDelegation *big.Int `abi:"minSelfDelegation"`
	BondDenom         string   `abi:"bondDenom"`
}

type ValidatorOutput struct {
	Validator ValidatorInfo
}

// DefaultValidatorOutput returns a ValidatorOutput with default values.
func DefaultValidatorOutput() ValidatorOutput {
	return ValidatorOutput{
		ValidatorInfo{
			OperatorAddress:   "",
			ConsensusPubkey:   "",
			Jailed:            false,
			Status:            uint8(0),
			Tokens:            big.NewInt(0),
			DelegatorShares:   big.NewInt(0),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     int64(0),
			Commission:        big.NewInt(0),
			MinSelfDelegation: big.NewInt(0),
			BondDenom:         "",
		},
	}
}

// FromResponse populates the ValidatorOutput from a QueryValidatorResponse.
func (vo *ValidatorOutput) FromResponse(res *multistakingtypes.QueryValidatorResponse) ValidatorOutput {
	operatorAddress, err := sdk.ValAddressFromBech32(res.Validator.OperatorAddress)
	if err != nil {
		return DefaultValidatorOutput()
	}

	return ValidatorOutput{
		Validator: ValidatorInfo{
			OperatorAddress:   sdk.ValAddress(operatorAddress.Bytes()).String(),
			ConsensusPubkey:   FormatConsensusPubkey(res.Validator.ConsensusPubkey),
			Jailed:            res.Validator.Jailed,
			Status:            uint8(stakingtypes.BondStatus_value[res.Validator.Status.String()]), //#nosec G115 // enum will always be convertible to uint8
			Tokens:            res.Validator.Tokens.BigInt(),
			DelegatorShares:   res.Validator.DelegatorShares.BigInt(),
			Description:       res.Validator.Description,
			UnbondingHeight:   res.Validator.UnbondingHeight,
			UnbondingTime:     res.Validator.UnbondingTime.UTC().Unix(),
			Commission:        res.Validator.Commission.CommissionRates.Rate.BigInt(),
			MinSelfDelegation: res.Validator.MinSelfDelegation.BigInt(),
			BondDenom:         res.Validator.BondDenom,
		},
	}
}

// FormatConsensusPubkey format ConsensusPubkey into a base64 string
func FormatConsensusPubkey(consensusPubkey *codectypes.Any) string {
	ed25519pk, ok := consensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	if ok {
		return base64.StdEncoding.EncodeToString(ed25519pk.Bytes())
	}
	return consensusPubkey.String()
}
