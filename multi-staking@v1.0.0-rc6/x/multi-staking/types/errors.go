package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/multistaking module sentinel errors
var (
	ErrInvalidAddMultiStakingCoinProposal       = sdkerrors.Register(ModuleName, 2, "invalid add multi staking coin proposal")
	ErrInvalidUpdateBondWeightProposal          = sdkerrors.Register(ModuleName, 3, "invalid update bond weight proposal")
	ErrInvalidTotalMultiStakingLocks            = sdkerrors.Register(ModuleName, 4, "invalid total multi-staking lock")
	ErrInvalidTotalMultiStakingUnlocks          = sdkerrors.Register(ModuleName, 5, "invalid total multi-staking unlock")
	ErrInvalidMultiStakingUnlocksCreationHeight = sdkerrors.Register(ModuleName, 6, "invalid unlock creation height")
)
