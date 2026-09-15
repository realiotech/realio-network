<!--
order: 0
-->

# Claim

`x/claim` is the recovery counterpart to [`x/blacklist`](../blacklist/README.md):
once a holder of a leaked/blacklisted address finishes re-doing KYC on the
off-chain app and registers a new address, the module's admin links the old
address to the new one with `MsgLinkAddress`. Linking a pair also migrates
every active staking delegation the old address holds over to the new
address, in the same message — without unbonding and rebonding, so voting
power, the validator's total stake, and the unbonding/redelegation queues
are completely unaffected.

## State

- `Admin` (`collections.Item[string]`) — the bech32 address authorized to
  submit `MsgLinkAddress`. Set at genesis, or by a chain-upgrade handler
  poking the keeper directly (the same way `x/blacklist`'s leaked-address
  list and `x/asset`'s manager rotations are seeded on a live chain — see
  `app/migrations`). There is no message to change it.
- `AddressLinks` (`collections.Map[sdk.AccAddress, string]`) — old address
  (raw bytes) → new address (bech32 string). Many old addresses may point
  at the same new address (consolidating several leaked wallets into one
  safe one); an old address can only ever be linked once.

## Msg — `MsgLinkAddress`

The only message the module exposes. `admin` must match the configured
`Admin`. `old_address` must not already be linked, and must not itself be
someone else's `new_address` (no chaining). On success it:

1. Records `old_address -> new_address` in `AddressLinks`.
2. Calls `Keeper.MigrateDelegations` (`keeper/migrate.go`) to re-key every
   active delegation `old_address` holds.

Both steps happen in the same transaction, so a failure in step 2 rolls
back step 1 too — a link is never recorded without its delegations having
also moved (or there being none to move).

## Delegation migration — what moves and what doesn't

**In scope:** active (bonded) delegations only, for both native `x/staking`
and `x/multi-staking`. For each validator the old address delegates to,
`migrateOneDelegation`:

1. Redirects the delegation's distribution withdraw address to the new
   address (`SetDelegatorWithdrawAddr`), so any reward payout below is
   credited there, never to the compromised old address.
2. Flushes pending rewards and the old `(validator, oldAddr)`
   `DelegatorStartingInfo` via `distrKeeper.Hooks().BeforeDelegationSharesModified`
   — the same hook native `x/staking`'s `Unbond` calls before removing a
   delegation.
3. Removes the old delegation (`stakingKeeper.RemoveDelegation` — not
   `Unbond`/`Undelegate`, which would also shrink the validator's total
   tokens/shares).
4. Re-creates the shares under the new address with `SetDelegation`
   (merging into an existing delegation to the same validator if the new
   address already has one), bracketed by the matching
   `BeforeDelegationCreated`/`BeforeDelegationSharesModified` and
   `AfterDelegationModified` hooks so `DelegatorStartingInfo` and the
   historical-rewards reference counts stay correct. `SetDelegation` alone
   moves no coins and doesn't touch the validator's aggregate
   tokens/shares — only the bookkeeping key changes.
5. Moves the backing `MultiStakingLock` (the escrowed original-denom coin a
   bond-coin delegation represents), if any, from the old `LockID` to the
   new one, merging into an existing lock if the new address already has
   one for that validator.

This hook sequencing is load-bearing, not cosmetic — see the comment on
`migrateOneDelegation` for exactly which cosmos-sdk source it mirrors.
Getting the order wrong doesn't corrupt state silently; it surfaces later
as a wrong reward calculation or a panic in distribution's reference-count
bookkeeping (`panic("reference count should never exceed 2")`).

**Deliberately out of scope:** any unbonding delegation or in-flight
redelegation the old address already has when it's linked is left
untouched. It matures and pays out to the old address exactly as it would
have otherwise. This is safe because `BlacklistSendRestriction` only blocks
transfers *from* a blacklisted address — a module account paying matured
unbonding funds *to* one still succeeds (see `x/blacklist/README.md`). It's
also the simpler and lower-risk choice: `x/multi-staking`'s unbonding
payout (`keeper/unlock.go`) correlates its own `MultiStakingUnlock` record
with the native `UnbondingDelegation` purely by matching
`(DelegatorAddress, ValidatorAddress, CreationHeight)` at maturity time,
with no independent cross-reference — re-keying one side without the other
in lockstep would either silently orphan the lock or panic `EndBlocker`
with `"unlock entry not found"` when the entry matures. Migrating active
delegations only avoids that failure mode entirely.

## What this module does not do

It doesn't touch bank balances, vesting schedules, governance votes already
cast, or anything outside active staking/multi-staking delegations. It also
doesn't check whether `new_address` is itself blacklisted — that's left to
`x/blacklist`'s own enforcement layers.
