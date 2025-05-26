// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package multistaking

// Multistaking precompile method names
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
	DelegationMethod                    = "delegation"
	UnbondingDelegationMethod           = "unbondingDelegation"
	ValidatorMethod                     = "validator"
	ValidatorsMethod                    = "validators"
	DelegatorDelegationsMethod          = "delegatorDelegations"
	DelegatorUnbondingDelegationsMethod = "delegatorUnbondingDelegations"
)
