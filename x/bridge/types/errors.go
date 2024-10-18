package types

import (
	errorsmod "cosmossdk.io/errors"
)

// DONTCOVER

// x/bridge module sentinel errors
var (
	ErrCoinNotRegister          = errorsmod.Register(ModuleName, 1600, "coin not in register list")
	ErrCoinAlreadyRegister      = errorsmod.Register(ModuleName, 1601, "coin already in register list")
	ErrEpochDurationZero        = errorsmod.Register(ModuleName, 1602, "epoch duration should NOT be 0")
	ErrInflowThresholdExceeded  = errorsmod.Register(ModuleName, 1603, "inflow threshold exceeded")
	ErrOutflowThresholdExceeded = errorsmod.Register(ModuleName, 1604, "outflow threshold exceeded")
	ErrZeroRatelimit            = errorsmod.Register(ModuleName, 1605, "ratelimit must not be zero")
)
