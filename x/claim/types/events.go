package types

// EventTypeSkippedSelfDelegation is emitted once per validator whose
// self-delegation MigrateDelegations declined to touch because old_address
// is that validator's own operator account. Native cosmos-sdk only
// enforces MinSelfDelegation/jailing inside Unbond, which this migration
// deliberately does not call (see x/claim/README.md) -- so a validator's
// own stake needs a separate, deliberate remediation path instead of
// silently riding along with its regular delegations.
const EventTypeSkippedSelfDelegation = "skipped_self_delegation"

const (
	AttributeKeyOldAddress = "old_address"
	AttributeKeyValidator  = "validator_address"
)
