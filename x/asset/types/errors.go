package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/asset module sentinel errors
var (
	ErrInvalidVersion       = sdkerrors.Register(ModuleName, 1001, "invalid version")
	ErrMintingOnCreate      = sdkerrors.Register(ModuleName, 1002, "error miniting creating asset")
	ErrDistributingOnCreate = sdkerrors.Register(ModuleName, 1003, "error distributing created asset")
)
