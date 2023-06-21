package types

import (
	errorsmod "cosmossdk.io/errors"
)

// DONTCOVER

// x/asset module sentinel errors
var (
	ErrSample               = errorsmod.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout = errorsmod.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion       = errorsmod.Register(ModuleName, 1501, "invalid version")
	ErrNotAuthorized        = errorsmod.Register(ModuleName, 1502, "transaction not authorized")
)
