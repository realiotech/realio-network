// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package feegrant

import (
	"fmt"
	"time"

	feegranttypes "cosmossdk.io/x/feegrant"

	"github.com/ethereum/go-ethereum/common"

	cmn "github.com/cosmos/evm/precompiles/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	FeeGrantPrecompileAddress = "0x0000000000000000000000000000000000000901"
	// Transactions
	GrantMethod  = "grant"
	RevokeMethod = "revoke"
)

// NewGrantRequest parses ABI arguments and builds a MsgGrantAllowance directly.
// Based on NewCmdFeeGrant logic: determines the appropriate allowance type from inputs.
// - BasicAllowance if no period is set
// - PeriodicAllowance if period > 0 and periodLimit is provided
// - Wraps in AllowedMsgAllowance if allowedMessages is non-empty
// args: [grantee address, spendLimit string, expiration int64, period int64, periodLimit string, allowedMessages string[]]
func NewGrantRequest(origin common.Address, args []interface{}) (*feegranttypes.MsgGrantAllowance, error) {
	if len(args) != 6 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 6, len(args))
	}

	granteeAddr, ok := args[0].(common.Address)
	if !ok || granteeAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	// Parse spendLimit string (e.g. "100stake") using sdk.ParseCoinsNormalized
	// If empty string, limit will be nil — no spend limit for the grantee.
	spendLimitStr, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "spendLimit", "string", args[1])
	}

	var spendLimit sdk.Coins
	if spendLimitStr != "" {
		var err error
		spendLimit, err = sdk.ParseCoinsNormalized(spendLimitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid spendLimit: %v", err)
		}
	}

	// Parse expiration string in RFC3339 format (e.g. "2026-01-01T00:00:00Z")
	// If empty string, no expiration is set.
	expirationStr, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "expiration", "string", args[2])
	}

	basic := feegranttypes.BasicAllowance{
		SpendLimit: spendLimit,
	}

	var expiresAtTime time.Time
	if expirationStr != "" {
		expiresAtTime, err := time.Parse(time.RFC3339, expirationStr)
		if err != nil {
			return nil, fmt.Errorf("invalid expiration: %v", err)
		}
		basic.Expiration = &expiresAtTime
	}

	period, ok := args[3].(int64)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "period", "int64", args[3])
	}

	// Parse periodLimit string (e.g. "10stake") using sdk.ParseCoinsNormalized
	periodLimitStr, ok := args[4].(string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "periodLimit", "string", args[4])
	}

	allowedMessages, ok := args[5].([]string)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "allowedMessages", "[]string", args[5])
	}

	// Determine allowance type
	var grant feegranttypes.FeeAllowanceI
	grant = &basic

	// If period is set, create PeriodicAllowance
	if period > 0 || periodLimitStr != "" {
		periodLimit, err := sdk.ParseCoinsNormalized(periodLimitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid period limit: %v", err)
		}

		if period <= 0 {
			return nil, fmt.Errorf("period clock was not set")
		}

		if periodLimit == nil {
			return nil, fmt.Errorf("period limit was not set")
		}

		periodReset := getPeriodReset(period)
		if expirationStr != "" && periodReset.Sub(expiresAtTime) > 0 {
			return nil, fmt.Errorf("period (%d) cannot reset after expiration (%v)", period, expirationStr)
		}

		periodic := feegranttypes.PeriodicAllowance{
			Basic:            basic,
			Period:           getPeriod(period),
			PeriodSpendLimit: periodLimit,
			PeriodCanSpend:   periodLimit,
			PeriodReset:      getPeriodReset(period),
		}
		grant = &periodic
	}

	// If allowedMessages is non-empty, wrap in AllowedMsgAllowance
	var err error
	if len(allowedMessages) > 0 {
		grant, err = feegranttypes.NewAllowedMsgAllowance(grant, allowedMessages)
		if err != nil {
			return nil, fmt.Errorf("failed to create allowed msg allowance: %v", err)
		}
	}

	// Build MsgGrantAllowance
	msg, err := feegranttypes.NewMsgGrantAllowance(grant, sdk.AccAddress(origin.Bytes()), sdk.AccAddress(granteeAddr.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("failed to create grant msg: %v", err)
	}

	return msg, nil
}

// NewRevokeRequest creates a new revoke request from ABI arguments
func NewRevokeRequest(args []interface{}) (sdk.AccAddress, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	granteeAddr, ok := args[0].(common.Address)
	if !ok || granteeAddr == (common.Address{}) {
		return nil, fmt.Errorf(cmn.ErrInvalidDelegator, args[0])
	}

	return sdk.AccAddress(granteeAddr.Bytes()), nil
}

// GrantOutput is a struct to represent the output of a grant operation
type GrantOutput struct {
	Success bool
}

// Pack packs a given slice of abi arguments into a byte array
func (g *GrantOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(g.Success)
}

// RevokeOutput is a struct to represent the output of a revoke operation
type RevokeOutput struct {
	Success bool
}

// Pack packs a given slice of abi arguments into a byte array
func (ro *RevokeOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(ro.Success)
}

func getPeriodReset(duration int64) time.Time {
	return time.Now().Add(getPeriod(duration))
}

func getPeriod(duration int64) time.Duration {
	return time.Duration(duration) * time.Second
}
