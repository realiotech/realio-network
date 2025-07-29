// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package multistaking

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	cmn "github.com/cosmos/evm/precompiles/common"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
	EditValidatorMethod             = "editValidator"

	// Queries
	DelegationMethod          = "delegation"
	UnbondingDelegationMethod = "unbondingDelegation"
	ValidatorMethod           = "validator"
	ValidatorsMethod          = "validators"
	RedelegationMethod        = "redelegation"
	RedelegationsMethod       = "redelegations"
)

// RedelegationRequest is a struct that contains the information to pass into a redelegation query.
type RedelegationRequest struct {
	DelegatorAddress    sdk.AccAddress
	ValidatorSrcAddress sdk.ValAddress
	ValidatorDstAddress sdk.ValAddress
}

// NewRedelegationRequest create a new QueryRedelegationRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewRedelegationRequest(args []interface{}) (*RedelegationRequest, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorSrcAddress, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "validatorSrcAddress", "string", args[1])
	}

	validatorSrcAddr, err := sdk.ValAddressFromBech32(validatorSrcAddress)
	if err != nil {
		return nil, err
	}

	validatorDstAddress, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "validatorDstAddress", "string", args[2])
	}

	validatorDstAddr, err := sdk.ValAddressFromBech32(validatorDstAddress)
	if err != nil {
		return nil, err
	}

	return &RedelegationRequest{
		DelegatorAddress:    delegatorAddr.Bytes(), // bech32 formatted
		ValidatorSrcAddress: validatorSrcAddr,
		ValidatorDstAddress: validatorDstAddr,
	}, nil
}

// DelegationOutput is a struct to represent the key information from
// a delegation response.
type DelegationOutput struct {
	Shares  *big.Int
	Balance cmn.Coin
}

// FromResponse populates the DelegationOutput from a QueryDelegationResponse.
func (do *DelegationOutput) FromResponse(res *stakingtypes.QueryDelegationResponse) *DelegationOutput {
	do.Shares = res.DelegationResponse.Delegation.Shares.BigInt()
	do.Balance = cmn.Coin{
		Denom:  res.DelegationResponse.Balance.Denom,
		Amount: res.DelegationResponse.Balance.Amount.BigInt(),
	}
	return do
}

// Pack packs a given slice of abi arguments into a byte array.
func (do *DelegationOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(do.Shares, do.Balance)
}

// NewDelegationRequest creates a new QueryDelegationRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewDelegationRequest(args []interface{}) (*stakingtypes.QueryDelegationRequest, error) {
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

	return &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: sdk.AccAddress(delegatorAddr.Bytes()).String(), // bech32 formatted
		ValidatorAddr: validatorAddress,
	}, nil
}

// UnbondingDelegationEntry is a struct that contains the information about an unbonding delegation entry.
type UnbondingDelegationEntry struct {
	CreationHeight          int64
	CompletionTime          int64
	InitialBalance          *big.Int
	Balance                 *big.Int
	UnbondingId             uint64 //nolint
	UnbondingOnHoldRefCount int64
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
func (do *UnbondingDelegationOutput) FromResponse(res *stakingtypes.QueryUnbondingDelegationResponse) *UnbondingDelegationOutput {
	do.UnbondingDelegation.Entries = make([]UnbondingDelegationEntry, len(res.Unbond.Entries))
	do.UnbondingDelegation.ValidatorAddress = res.Unbond.ValidatorAddress
	do.UnbondingDelegation.DelegatorAddress = res.Unbond.DelegatorAddress
	for i, entry := range res.Unbond.Entries {
		do.UnbondingDelegation.Entries[i] = UnbondingDelegationEntry{
			UnbondingId:             entry.UnbondingId,
			UnbondingOnHoldRefCount: entry.UnbondingOnHoldRefCount,
			CreationHeight:          entry.CreationHeight,
			CompletionTime:          entry.CompletionTime.UTC().Unix(),
			InitialBalance:          entry.InitialBalance.BigInt(),
			Balance:                 entry.Balance.BigInt(),
		}
	}
	return do
}

// NewUnbondingDelegationRequest creates a new QueryUnbondingDelegationRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewUnbondingDelegationRequest(args []interface{}) (*stakingtypes.QueryUnbondingDelegationRequest, error) {
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

	return &stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: sdk.AccAddress(delegatorAddr.Bytes()).String(), // bech32 formatted
		ValidatorAddr: validatorAddress,
	}, nil
}

// NewValidatorRequest create a new QueryValidatorRequest instance and does sanity checks
// on the given arguments before populating the request.
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

// ValidatorInfo is a struct to represent the key information from
// a validator response.
type ValidatorInfo struct {
	OperatorAddress   string   `abi:"operatorAddress"`
	ConsensusPubkey   string   `abi:"consensusPubkey"`
	Jailed            bool     `abi:"jailed"`
	Status            uint8    `abi:"status"`
	Tokens            *big.Int `abi:"tokens"`
	DelegatorShares   *big.Int `abi:"delegatorShares"` // TODO: Decimal
	Description       string   `abi:"description"`
	UnbondingHeight   int64    `abi:"unbondingHeight"`
	UnbondingTime     int64    `abi:"unbondingTime"`
	Commission        *big.Int `abi:"commission"`
	MinSelfDelegation *big.Int `abi:"minSelfDelegation"`
}

// ValidatorsInput is a struct to represent the input information for
// the validators query. Needed to unpack arguments into the PageRequest struct.
type ValidatorsInput struct {
	Status      string
	PageRequest query.PageRequest
}

// ValidatorsOutput is a struct to represent the key information from
// a validators response.
type ValidatorsOutput struct {
	Validators   []ValidatorInfo
	PageResponse query.PageResponse
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
			Description:       "",
			UnbondingHeight:   int64(0),
			UnbondingTime:     int64(0),
			Commission:        big.NewInt(0),
			MinSelfDelegation: big.NewInt(0),
		},
	}
}

// FromResponse populates the ValidatorsOutput from a QueryValidatorsResponse.
func (vo *ValidatorsOutput) FromResponse(res *multistakingtypes.QueryValidatorsResponse) *ValidatorsOutput {
	vo.Validators = make([]ValidatorInfo, len(res.Validators))
	for i, v := range res.Validators {
		operatorAddress, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		if err != nil {
			vo.Validators[i] = DefaultValidatorOutput().Validator
		} else {
			vo.Validators[i] = ValidatorInfo{
				OperatorAddress:   common.BytesToAddress(operatorAddress.Bytes()).String(),
				ConsensusPubkey:   FormatConsensusPubkey(v.ConsensusPubkey),
				Jailed:            v.Jailed,
				Status:            uint8(stakingtypes.BondStatus_value[v.Status.String()]), //#nosec G115 // enum will always be convertible to uint8
				Tokens:            v.Tokens.BigInt(),
				DelegatorShares:   v.DelegatorShares.BigInt(),
				Description:       v.Description.Details,
				UnbondingHeight:   v.UnbondingHeight,
				UnbondingTime:     v.UnbondingTime.UTC().Unix(),
				Commission:        v.Commission.CommissionRates.Rate.BigInt(),
				MinSelfDelegation: v.MinSelfDelegation.BigInt(),
			}
		}
	}

	if res.Pagination != nil {
		vo.PageResponse.Total = res.Pagination.Total
		vo.PageResponse.NextKey = res.Pagination.NextKey
	}

	return vo
}

// Pack packs a given slice of abi arguments into a byte array.
func (vo *ValidatorsOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(vo.Validators, vo.PageResponse)
}

// FromResponse populates the ValidatorOutput from a QueryValidatorResponse.
func (vo *ValidatorOutput) FromResponse(res *multistakingtypes.QueryValidatorResponse) ValidatorOutput {
	operatorAddress, err := sdk.ValAddressFromBech32(res.Validator.OperatorAddress)
	if err != nil {
		return DefaultValidatorOutput()
	}

	return ValidatorOutput{
		Validator: ValidatorInfo{
			OperatorAddress: common.BytesToAddress(operatorAddress.Bytes()).String(),
			ConsensusPubkey: FormatConsensusPubkey(res.Validator.ConsensusPubkey),
			Jailed:          res.Validator.Jailed,
			Status:          uint8(stakingtypes.BondStatus_value[res.Validator.Status.String()]), //#nosec G115 // enum will always be convertible to uint8
			Tokens:          res.Validator.Tokens.BigInt(),
			DelegatorShares: res.Validator.DelegatorShares.BigInt(), // TODO: Decimal
			// TODO: create description type,
			Description:       res.Validator.Description.Details,
			UnbondingHeight:   res.Validator.UnbondingHeight,
			UnbondingTime:     res.Validator.UnbondingTime.UTC().Unix(),
			Commission:        res.Validator.Commission.CommissionRates.Rate.BigInt(),
			MinSelfDelegation: res.Validator.MinSelfDelegation.BigInt(),
		},
	}
}

// NewValidatorsRequest create a new QueryValidatorsRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewValidatorsRequest(method *abi.Method, args []interface{}) (*multistakingtypes.QueryValidatorsRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	var input ValidatorsInput
	if err := method.Inputs.Copy(&input, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to ValidatorsInput struct: %s", err)
	}

	if bytes.Equal(input.PageRequest.Key, []byte{0}) {
		input.PageRequest.Key = nil
	}

	return &multistakingtypes.QueryValidatorsRequest{
		Status:     input.Status,
		Pagination: &input.PageRequest,
	}, nil
}

// RedelegationEntry is a struct to represent the key information from
// a redelegation entry response.
type RedelegationEntry struct {
	CreationHeight int64
	CompletionTime int64
	InitialBalance *big.Int
	SharesDst      *big.Int
}

// RedelegationEntryResponse is equivalent to a RedelegationEntry except that it
// contains a balance in addition to shares which is more suitable for client
// responses.
type RedelegationEntryResponse struct {
	RedelegationEntry RedelegationEntry
	Balance           *big.Int
}

// Redelegation contains the list of a particular delegator's redelegating bonds
// from a particular source validator to a particular destination validator.
type Redelegation struct {
	DelegatorAddress    string
	ValidatorSrcAddress string
	ValidatorDstAddress string
	Entries             []RedelegationEntry
}

// RedelegationResponse is equivalent to a Redelegation except that its entries
// contain a balance in addition to shares which is more suitable for client
// responses.
type RedelegationResponse struct {
	Redelegation Redelegation
	Entries      []RedelegationEntryResponse
}

// RedelegationsInput is a struct to represent the input information for
// the redelegations query. Needed to unpack arguments into the PageRequest struct.
type RedelegationsInput struct {
	DelegatorAddress    common.Address
	SrcValidatorAddress string
	DstValidatorAddress string
	PageRequest         query.PageRequest
}

// RedelegationsOutput is a struct to represent the key information from
// a redelegations response.
type RedelegationsOutput struct {
	Response     []RedelegationResponse
	PageResponse query.PageResponse
}

// RedelegationValues is a struct to represent the key information from
// a redelegation response.
type RedelegationValues struct {
	DelegatorAddress    string
	ValidatorSrcAddress string
	ValidatorDstAddress string
	Entries             []RedelegationEntry
}

// RedelegationOutput returns the output for a redelegation query.
type RedelegationOutput struct {
	Redelegation RedelegationValues
}

// FromResponse populates the RedelegationOutput from a QueryRedelegationsResponse.
func (ro *RedelegationOutput) FromResponse(res stakingtypes.Redelegation) *RedelegationOutput {
	ro.Redelegation.Entries = make([]RedelegationEntry, len(res.Entries))
	ro.Redelegation.DelegatorAddress = res.DelegatorAddress
	ro.Redelegation.ValidatorSrcAddress = res.ValidatorSrcAddress
	ro.Redelegation.ValidatorDstAddress = res.ValidatorDstAddress
	for i, entry := range res.Entries {
		ro.Redelegation.Entries[i] = RedelegationEntry{
			CreationHeight: entry.CreationHeight,
			CompletionTime: entry.CompletionTime.UTC().Unix(),
			InitialBalance: entry.InitialBalance.BigInt(),
			SharesDst:      entry.SharesDst.BigInt(),
		}
	}
	return ro
}

// NewRedelegationsRequest create a new QueryRedelegationsRequest instance and does sanity checks
// on the given arguments before populating the request.
func NewRedelegationsRequest(method *abi.Method, args []interface{}) (*stakingtypes.QueryRedelegationsRequest, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	// delAddr, srcValAddr & dstValAddr
	// can be empty strings. The query will return the
	// corresponding redelegations according to the addresses specified
	// however, cannot pass all as empty strings, need to provide at least
	// the delegator address or the source validator address
	var input RedelegationsInput
	if err := method.Inputs.Copy(&input, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to RedelegationsInput struct: %s", err)
	}

	var (
		// delegatorAddr is the string representation of the delegator address
		delegatorAddr = ""
		// emptyAddr is an empty address
		emptyAddr = common.Address{}.Hex()
	)
	if input.DelegatorAddress.Hex() != emptyAddr {
		delegatorAddr = sdk.AccAddress(input.DelegatorAddress.Bytes()).String() // bech32 formatted
	}

	if delegatorAddr == "" && input.SrcValidatorAddress == "" && input.DstValidatorAddress == "" ||
		delegatorAddr == "" && input.SrcValidatorAddress == "" && input.DstValidatorAddress != "" {
		return nil, errors.New("invalid query. Need to specify at least a source validator address or delegator address")
	}

	return &stakingtypes.QueryRedelegationsRequest{
		DelegatorAddr:    delegatorAddr, // bech32 formatted
		SrcValidatorAddr: input.SrcValidatorAddress,
		DstValidatorAddr: input.DstValidatorAddress,
		Pagination:       &input.PageRequest,
	}, nil
}

// FromResponse populates the RedelgationsOutput from a QueryRedelegationsResponse.
func (ro *RedelegationsOutput) FromResponse(res *stakingtypes.QueryRedelegationsResponse) *RedelegationsOutput {
	ro.Response = make([]RedelegationResponse, len(res.RedelegationResponses))
	for i, resp := range res.RedelegationResponses {
		// for each RedelegationResponse
		// there's a RedelegationEntryResponse array ('Entries' field)
		entries := make([]RedelegationEntryResponse, len(resp.Entries))
		for j, e := range resp.Entries {
			entries[j] = RedelegationEntryResponse{
				RedelegationEntry: RedelegationEntry{
					CreationHeight: e.RedelegationEntry.CreationHeight,
					CompletionTime: e.RedelegationEntry.CompletionTime.Unix(),
					InitialBalance: e.RedelegationEntry.InitialBalance.BigInt(),
					SharesDst:      e.RedelegationEntry.SharesDst.BigInt(),
				},
				Balance: e.Balance.BigInt(),
			}
		}

		// the Redelegation field has also an 'Entries' field of type RedelegationEntry
		redelEntries := make([]RedelegationEntry, len(resp.Redelegation.Entries))
		for j, e := range resp.Redelegation.Entries {
			redelEntries[j] = RedelegationEntry{
				CreationHeight: e.CreationHeight,
				CompletionTime: e.CompletionTime.Unix(),
				InitialBalance: e.InitialBalance.BigInt(),
				SharesDst:      e.SharesDst.BigInt(),
			}
		}

		ro.Response[i] = RedelegationResponse{
			Entries: entries,
			Redelegation: Redelegation{
				DelegatorAddress:    resp.Redelegation.DelegatorAddress,
				ValidatorSrcAddress: resp.Redelegation.ValidatorSrcAddress,
				ValidatorDstAddress: resp.Redelegation.ValidatorDstAddress,
				Entries:             redelEntries,
			},
		}
	}

	if res.Pagination != nil {
		ro.PageResponse.Total = res.Pagination.Total
		ro.PageResponse.NextKey = res.Pagination.NextKey
	}

	return ro
}

// Pack packs a given slice of abi arguments into a byte array.
func (ro *RedelegationsOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(ro.Response, ro.PageResponse)
}

// FormatConsensusPubkey format ConsensusPubkey into a base64 string
func FormatConsensusPubkey(consensusPubkey *codectypes.Any) string {
	ed25519pk, ok := consensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	if ok {
		return base64.StdEncoding.EncodeToString(ed25519pk.Bytes())
	}
	return consensusPubkey.String()
}
