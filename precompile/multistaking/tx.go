// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package multistaking

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	// codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
)

// Delegate handles the delegation of ERC20 tokens to a validator.
// This method converts ERC20 tokens to SDK coins and delegates them.
func (p Precompile) Delegate(
	ctx sdk.Context,
	sender common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	erc20Token, validatorAddress, amount, err := checkDelegationUndelegationArgs(args)
	if err != nil {
		return nil, err
	}

	// Convert ERC20 to SDK coin
	coin, err := p.convertERC20ToSDKCoin(ctx, sender, erc20Token, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ERC20 to SDK coin: %v", err)
	}

	// Create multistaking delegation message
	msg := &stakingtypes.MsgDelegate{
		DelegatorAddress: sdk.AccAddress(sender.Bytes()).String(),
		ValidatorAddress: validatorAddress,
		Amount:           coin,
	}

	// Execute delegation using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	_, err = msgServer.Delegate(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Return success
	return method.Outputs.Pack(true)
}

// Undelegate handles the undelegation of tokens from a validator.
// This method undelegates SDK coins and converts them back to ERC20 tokens.
func (p Precompile) Undelegate(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	erc20Token, validatorAddress, amount, err := checkDelegationUndelegationArgs(args)
	if err != nil {
		return nil, err
	}

	// Convert caller address to SDK address
	delegatorAddr := sdk.AccAddress(origin.Bytes())

	// Get the corresponding SDK coin denom for the ERC20 token
	denom, err := p.getSDKDenomForERC20(ctx, erc20Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get SDK denom for ERC20: %v", err)
	}

	// Create coin for undelegation
	coin := sdk.NewCoin(denom, math.NewIntFromBigInt(amount))

	// Create multistaking undelegation message
	msg := &stakingtypes.MsgUndelegate{
		DelegatorAddress: delegatorAddr.String(),
		ValidatorAddress: validatorAddress,
		Amount:           coin,
	}

	// Execute undelegation using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	resp, err := msgServer.Undelegate(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("multistaking undelegation failed: %v", err)
	}

	// Convert undelegated coins back to ERC20
	err = p.convertSDKCoinToERC20(ctx, delegatorAddr, coin)
	if err != nil {
		return nil, fmt.Errorf("failed to convert SDK coin back to ERC20: %v", err)
	}

	// Return completion time
	return method.Outputs.Pack(resp.CompletionTime.Unix())
}

// Redelegate handles the redelegation of tokens from one validator to another.
func (p Precompile) Redelegate(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	erc20Token, srcValidatorAddress, dstValidatorAddress, amount, err := parseRedelegateArgs(args)
	if err != nil {
		return nil, err
	}

	// Convert caller address to SDK address
	delegatorAddr := sdk.AccAddress(origin.Bytes())

	// Get the corresponding SDK coin denom for the ERC20 token
	denom, err := p.getSDKDenomForERC20(ctx, erc20Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get SDK denom for ERC20: %v", err)
	}

	// Create coin for redelegation
	coin := sdk.NewCoin(denom, math.NewIntFromBigInt(amount))

	// Create multistaking redelegation message
	msg := &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    delegatorAddr.String(),
		ValidatorSrcAddress: srcValidatorAddress,
		ValidatorDstAddress: dstValidatorAddress,
		Amount:              coin,
	}

	// Execute redelegation using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	resp, err := msgServer.BeginRedelegate(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("multistaking redelegation failed: %v", err)
	}

	// Return completion time
	return method.Outputs.Pack(resp.CompletionTime.Unix())
}

// CancelUnbondingDelegation handles the cancellation of an unbonding delegation.
func (p Precompile) CancelUnbondingDelegation(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	erc20Token, validatorAddress, amount, creationHeight, err := parseCancelUnbondingArgs(args)
	if err != nil {
		return nil, err
	}

	// Convert caller address to SDK address
	delegatorAddr := sdk.AccAddress(origin.Bytes())

	// Get the corresponding SDK coin denom for the ERC20 token
	denom, err := p.getSDKDenomForERC20(ctx, erc20Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get SDK denom for ERC20: %v", err)
	}

	// Create coin for cancel unbonding delegation
	coin := sdk.NewCoin(denom, math.NewIntFromBigInt(amount))

	// Create multistaking cancel unbonding delegation message
	msg := &stakingtypes.MsgCancelUnbondingDelegation{
		DelegatorAddress: delegatorAddr.String(),
		ValidatorAddress: validatorAddress,
		Amount:           coin,
		CreationHeight:   creationHeight,
	}

	// Execute cancel unbonding delegation using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	_, err = msgServer.CancelUnbondingDelegation(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("multistaking cancel unbonding delegation failed: %v", err)
	}

	// Return success
	return method.Outputs.Pack(true)
}

// CreateValidator creates a new validator using the multistaking module.
func (p Precompile) CreateValidator(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	validatorAddress, pubkey, contractAddress, amount, commission, description, minSelfDelegation, err := p.parseCreateValidatorArgs(args)
	if err != nil {
		return nil, err
	}

	// Get delegator address from origin
	// delegatorAddr := sdk.AccAddress(origin.Bytes())

	// Convert ERC20 to SDK coin
	coin, err := p.convertERC20ToSDKCoin(ctx, origin, contractAddress, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ERC20 to SDK coin: %v", err)
	}

	msg, err := stakingtypes.NewMsgCreateValidator(
		validatorAddress,
		pubkey,
		coin,
		description,
		commission,
		minSelfDelegation,
	)

	// Execute create validator using multistaking msgServer
	msgServer := multistakingkeeper.NewMsgServerImpl(p.multiStakingKeeper)
	_, err = msgServer.CreateValidator(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("multistaking create validator failed: %v", err)
	}

	return method.Outputs.Pack(true)
}

// EditValidator edits an existing validator using the multistaking module.
func (p Precompile) EditValidator(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
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

	delegatorAddr, ok := args[0].(common.Address)
	if !ok || delegatorAddr == (common.Address{}) {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	validatorAddress, ok := args[1].(string)
	if !ok {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidType, "validatorAddress", "string", args[1])
	}

	amount, ok := args[2].(*big.Int)
	if !ok {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidAmount, args[2])
	}

	return delegatorAddr, validatorAddress, amount, nil
}

func parseRedelegateArgs(args []interface{}) (common.Address, string, string, *big.Int, error) {
	if len(args) != 4 {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid number of arguments for redelegate")
	}

	erc20Token, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid erc20Token address")
	}

	srcValidatorAddress, ok := args[1].(string)
	if !ok {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid source validator address")
	}

	dstValidatorAddress, ok := args[2].(string)
	if !ok {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid destination validator address")
	}

	amount, ok := args[3].(*big.Int)
	if !ok {
		return common.Address{}, "", "", nil, fmt.Errorf("invalid amount")
	}

	return erc20Token, srcValidatorAddress, dstValidatorAddress, amount, nil
}

func parseCancelUnbondingArgs(args []interface{}) (common.Address, string, *big.Int, int64, error) {
	if len(args) != 4 {
		return common.Address{}, "", nil, 0, fmt.Errorf("invalid number of arguments for cancelUnbondingDelegation")
	}

	erc20Token, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, "", nil, 0, fmt.Errorf("invalid erc20Token address")
	}

	validatorAddress, ok := args[1].(string)
	if !ok {
		return common.Address{}, "", nil, 0, fmt.Errorf("invalid validator address")
	}

	amount, ok := args[2].(*big.Int)
	if !ok {
		return common.Address{}, "", nil, 0, fmt.Errorf("invalid amount")
	}

	creationHeight, ok := args[3].(*big.Int)
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

	var pk cryptotypes.PubKey
	if err := p.Codec.UnmarshalInterfaceJSON([]byte(pubkeyStr), &pk); err != nil {
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
