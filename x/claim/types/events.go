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
	// EventTypeClaimedBalance is emitted once per MsgClaimBalances, with
	// the full set of coins that moved.
	EventTypeClaimedBalance = "claimed_balance"

	// EventTypeClaimedERC20 is emitted once per ERC-20 contract whose
	// balance MsgClaimTokens actually moved. Contracts where the old
	// address held nothing emit no event.
	EventTypeClaimedERC20 = "claimed_erc20"
)

const (
	AttributeKeyOldAddress = "old_address"
	AttributeKeyNewAddress = "new_address"
	AttributeKeyValidator  = "validator_address"
	AttributeKeyContract   = "contract"
	AttributeKeyAmount     = "amount"
)
