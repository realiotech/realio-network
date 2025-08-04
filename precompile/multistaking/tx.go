// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package multistaking

import (
	"encoding/json"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	// codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	mstypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
)

// Delegate handles the delegation of ERC20 tokens to a validator.
// This method converts ERC20 tokens to SDK coins and delegates them.
func (p Precompile) DelegateEVM(
	ctx sdk.Context,
	sender common.Address,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	erc20Token, validatorAddress, amount, err := checkDelegationUndelegationArgs(args)
	if err != nil {
		return nil, err
	}

	delAddr, err := p.addrCodec.BytesToString(sender.Bytes())
	if err != nil {
		return nil, err
	}
	// Create multistaking delegation message

	msg := &mstypes.MsgDelegateEVM{
		DelegatorAddress: delAddr,
		ValidatorAddress: validatorAddress,
		Amount:           math.NewIntFromBigInt(amount),
		ContractAddress:  erc20Token.Hex(),
	}

	// Execute delegation using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	_, err = msgServer.DelegateEVM(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Return success
	return method.Outputs.Pack(true)
}

// Undelegate handles the undelegation of tokens from a validator.
// This method undelegates SDK coins and converts them back to ERC20 tokens.
func (p Precompile) UndelegateEVM(
	ctx sdk.Context,
	origin common.Address,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	erc20Token, validatorAddress, amount, err := checkDelegationUndelegationArgs(args)
	if err != nil {
		return nil, err
	}

	// Convert caller address to SDK address
	delegatorAddr, err := p.addrCodec.BytesToString(origin.Bytes())
	if err != nil {
		return nil, err
	}

	// Create multistaking undelegation message
	msg := &mstypes.MsgUndelegateEVM{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddress,
		Amount:           math.NewIntFromBigInt(amount),
		ContractAddress:  erc20Token.Hex(),
	}

	// Execute undelegation using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	resp, err := msgServer.UndelegateEVM(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("multistaking undelegation failed: %v", err)
	}

	// Return completion time
	return method.Outputs.Pack(resp.CompletionTime.Unix())
}

// BeginRedelegateEVM handles the redelegation of tokens from one validator to another.
func (p Precompile) BeginRedelegateEVM(
	ctx sdk.Context,
	origin common.Address,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	erc20Token, srcValidatorAddress, dstValidatorAddress, amount, err := parseRedelegateArgs(args)
	if err != nil {
		return nil, err
	}

	// Convert caller address to SDK address
	delegatorAddr, err := p.addrCodec.BytesToString(origin.Bytes())
	if err != nil {
		return nil, err
	}

	// Create multistaking redelegation message
	msg := &mstypes.MsgBeginRedelegateEVM{
		DelegatorAddress:    delegatorAddr,
		ValidatorSrcAddress: srcValidatorAddress,
		ValidatorDstAddress: dstValidatorAddress,
		Amount:              math.NewIntFromBigInt(amount),
		ContractAddress:     erc20Token.Hex(),
	}

	// Execute redelegation using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	resp, err := msgServer.BeginRedelegateEVM(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("multistaking redelegation failed: %v", err)
	}

	// Return completion time
	return method.Outputs.Pack(resp.CompletionTime.Unix())
}

// CancelUnbondingEVMDelegation handles the cancellation of an unbonding delegation using erc20 token.
func (p Precompile) CancelUnbondingEVMDelegation(
	ctx sdk.Context,
	origin common.Address,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	erc20Token, validatorAddress, amount, creationHeight, err := parseCancelUnbondingArgs(args)
	if err != nil {
		return nil, err
	}

	// Convert caller address to SDK address
	delegatorAddr, err := p.addrCodec.BytesToString(origin.Bytes())
	if err != nil {
		return nil, err
	}

	// Create multistaking cancel unbonding evm delegation message
	msg := &mstypes.MsgCancelUnbondingEVMDelegation{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddress,
		Amount:           math.NewIntFromBigInt(amount),
		CreationHeight:   creationHeight,
		ContractAddress:  erc20Token.Hex(),
	}

	// Execute cancel unbonding delegation using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	_, err = msgServer.CancelUnbondingEVMDelegation(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("multistaking cancel unbonding delegation failed: %v", err)
	}

	// Return success
	return method.Outputs.Pack(true)
}

// CreateValidator creates a new erc20 validator using the multistaking module.
func (p Precompile) CreateEVMValidator(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	validatorAddress, pubkey, contractAddress, amount, commission, description, minSelfDelegation, err := p.parseCreateValidatorArgs(args)
	if err != nil {
		return nil, err
	}

	var pkAny *codectypes.Any
	if pubkey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubkey); err != nil {
			return nil, err
		}
	}

	msg := &mstypes.MsgCreateEVMValidator{
		Description:       description,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
		ValidatorAddress:  validatorAddress,
		Pubkey:            pkAny,
		ContractAddress:   contractAddress.Hex(),
		Value:             math.NewIntFromBigInt(amount),
	}

	// Execute create validator using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	_, err = msgServer.CreateEVMValidator(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("multistaking create validator failed: %v", err)
	}

	return method.Outputs.Pack(true)
}

// EditValidator edits an existing validator using the multistaking module.
func (p Precompile) EditValidator(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	validatorAddress, description, commissionRate, minSelfDelegation, err := parseEditValidatorArgs(args)
	if err != nil {
		return nil, err
	}

	// Create multistaking edit validator message
	msg := &stakingtypes.MsgEditValidator{
		Description:       description,
		ValidatorAddress:  validatorAddress,
		CommissionRate:    commissionRate,
		MinSelfDelegation: minSelfDelegation,
	}

	// Execute edit validator using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	_, err = msgServer.EditValidator(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("multistaking edit validator failed: %v", err)
	}

	return method.Outputs.Pack(true)
}

// Helper functions for parsing arguments

func checkDelegationUndelegationArgs(args []interface{}) (common.Address, string, *big.Int, error) {
	if len(args) != 3 {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	contractAddressStr, ok := args[0].(string)
	if !ok {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}
	contractAddress := common.HexToAddress(contractAddressStr)

	validatorAddress, ok := args[1].(string)
	if !ok {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidType, "validatorAddress", "string", args[1])
	}

	amount, ok := new(big.Int).SetString(args[2].(string), 10)
	if !ok {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidAmount, args[2])
	}

	return contractAddress, validatorAddress, amount, nil
}

func parseRedelegateArgs(args []interface{}) (common.Address, string, string, *big.Int, error) {
	if len(args) != 4 {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid number of arguments for redelegate")
	}

	erc20TokenStr, ok := args[0].(string)
	if !ok {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid erc20Token address")
	}
	erc20Token := common.HexToAddress(erc20TokenStr)

	srcValidatorAddress, ok := args[1].(string)
	if !ok {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid source validator address")
	}

	dstValidatorAddress, ok := args[2].(string)
	if !ok {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid destination validator address")
	}

	amount, ok := new(big.Int).SetString(args[3].(string), 10)
	if !ok {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid amount")
	}

	return erc20Token, srcValidatorAddress, dstValidatorAddress, amount, nil
}

func parseCancelUnbondingArgs(args []interface{}) (common.Address, string, *big.Int, int64, error) {
	if len(args) != 4 {
		return common.Address{}, "", nil, 0, fmt.Errorf("invalid number of arguments for cancelUnbondingDelegation")
	}

	erc20TokenStr, ok := args[0].(string)
	if !ok {
		return common.Address{}, "", nil, 0, fmt.Errorf("invalid erc20Token address")
	}
	erc20Token := common.HexToAddress(erc20TokenStr)

	validatorAddress, ok := args[1].(string)
	if !ok {
		return common.Address{}, "", nil, 0, fmt.Errorf("invalid validator address")
	}

	amount, ok := new(big.Int).SetString(args[2].(string), 10)
	if !ok {
		return common.Address{}, "", nil, 0, fmt.Errorf("invalid amount")
	}

	creationHeight, ok := new(big.Int).SetString(args[3].(string), 10)
	if !ok {
		return common.Address{}, "", nil, 0, fmt.Errorf("invalid creation height")
	}

	return erc20Token, validatorAddress, amount, creationHeight.Int64(), nil
}

func (p Precompile) parseCreateValidatorArgs(args []interface{}) (string, cryptotypes.PubKey, common.Address, *big.Int, stakingtypes.CommissionRates, stakingtypes.Description, math.Int, error) {
	if len(args) != 13 {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid number of arguments for createValidator, expected 13, got %d", len(args))
	}

	// Parse validatorAddress
	validatorAddress, ok := args[0].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid validator address")
	}

	// Parse pubkey as string
	pubkeyStr, ok := args[1].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid pubkey")
	}
	pkStr := fmt.Sprintf(`{"@type":"/cosmos.crypto.ed25519.PubKey","key":"%s"}`, pubkeyStr)

	pubkeyJSON := json.RawMessage(pkStr)

	var pk cryptotypes.PubKey
	if err := p.Codec.UnmarshalInterfaceJSON(pubkeyJSON, &pk); err != nil {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, err
	}

	// Parse contractAddress
	contractAddressStr, ok := args[2].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid contract address")
	}
	contractAddress := common.HexToAddress(contractAddressStr)

	// Parse amount (self delegation amount)
	amount, ok := new(big.Int).SetString(args[3].(string), 10)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid amount")
	}

	// Parse description fields
	moniker, ok := args[4].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid moniker")
	}

	identity, ok := args[5].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid identity")
	}

	website, ok := args[6].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid website")
	}

	security, ok := args[7].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid security")
	}

	details, ok := args[8].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid details")
	}

	description := stakingtypes.Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: security,
		Details:         details,
	}

	// Parse commission fields
	commissionRate, ok := args[9].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid commission rate")
	}

	commissionMaxRate, ok := args[10].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid commission max rate")
	}

	commissionMaxChangeRate, ok := args[11].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid commission max change rate")
	}

	commission := stakingtypes.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr(commissionRate),
		MaxRate:       math.LegacyMustNewDecFromStr(commissionMaxRate),
		MaxChangeRate: math.LegacyMustNewDecFromStr(commissionMaxChangeRate),
	}

	// Parse minSelfDelegation
	minSelfDelegationStr, ok := args[12].(string)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid min self delegation")
	}

	minSelfDelegation, ok := math.NewIntFromString(minSelfDelegationStr)
	if !ok {
		return "", nil, common.Address{}, nil, stakingtypes.CommissionRates{}, stakingtypes.Description{}, math.Int{}, fmt.Errorf("invalid min self delegation format")
	}

	return validatorAddress, pk, contractAddress, amount, commission, description, minSelfDelegation, nil
}

func parseEditValidatorArgs(args []interface{}) (string, stakingtypes.Description, *math.LegacyDec, *math.Int, error) {
	if len(args) != 8 {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid number of arguments for editValidator, expected 8, got %d", len(args))
	}

	// Parse validatorAddress
	validatorAddress, ok := args[0].(string)
	if !ok {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid validator address")
	}

	// Parse description fields
	moniker, ok := args[1].(string)
	if !ok {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid moniker")
	}

	identity, ok := args[2].(string)
	if !ok {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid identity")
	}

	website, ok := args[3].(string)
	if !ok {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid website")
	}

	security, ok := args[4].(string)
	if !ok {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid security")
	}

	details, ok := args[5].(string)
	if !ok {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid details")
	}

	description := stakingtypes.Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: security,
		Details:         details,
	}

	// Parse commission rate as *math.LegacyDec
	commissionRateStr, ok := args[6].(string)
	if !ok {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid commission rate")
	}

	commissionRate := math.LegacyMustNewDecFromStr(commissionRateStr)

	// Parse minSelfDelegation as *math.Int
	minSelfDelegationStr, ok := args[7].(string)
	if !ok {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid min self delegation")
	}

	minSelfDelegation, ok := math.NewIntFromString(minSelfDelegationStr)
	if !ok {
		return "", stakingtypes.Description{}, nil, nil, fmt.Errorf("invalid min self delegation format")
	}

	return validatorAddress, description, &commissionRate, &minSelfDelegation, nil
}

// Helper struct for self delegation
type SelfDelegation struct {
	Token  common.Address
	Amount *big.Int
}
