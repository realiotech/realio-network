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
and `x/multi-staking`. `MigrateDelegations` sets the old address's
distribution withdraw address to the new one up front (`SetDelegatorWithdrawAddr`,
once per `LinkAddress` call, not per validator — it's a per-account
setting) so that any reward payout this migration does trigger always
lands on the new address, never the compromised old one. Then, for each
validator the old address delegates to, `migrateOneDelegation` takes one
of two paths depending on whether the new address already delegates to
that same validator:

**No existing delegation on the new side (`migrateDelegationFresh`,
the common case).** Rather than paying out the old delegation's pending
reward as a side effect of the admin's `LinkAddress` call, its exact
reward-accrual ledger — the `DelegatorStartingInfo` recording which
historical reward period it last settled against — is carried over to the
new address verbatim (`GetDelegatorStartingInfo` → `SetDelegatorStartingInfo`
→ `DeleteDelegatorStartingInfo`, all called directly, bypassing
distribution's hooks entirely). The new address can then withdraw
whenever it likes, via the ordinary `MsgWithdrawDelegatorReward`, for
exactly what had accrued — nothing paid out early, nothing lost. This is
safe without touching the historical-rewards reference count (its
increment/decrement functions are unexported in cosmos-sdk, and rightly
so — callers aren't meant to manage it manually): relocating the ledger
from the old address's key to the new one's doesn't change how many
`DelegatorStartingInfo` entries point at that historical period, still
exactly one, just under a different owner.

**New side already has a delegation to that validator
(`migrateDelegationMerge`).** Two independent reward-accrual ledgers can't
be combined losslessly into the single `DelegatorStartingInfo` a merged
delegation needs, so this path can't avoid a flush: both sides' pending
rewards are paid out (to the new address) and the combined position starts
fresh, via `distrKeeper.Hooks().BeforeDelegationSharesModified` /
`AfterDelegationModified` — the same hooks native `x/staking`'s `Unbond`
and `Delegate` call internally for the equivalent state changes. This
isn't a weaker guarantee than an ordinary delegation gets: native
`x/staking` flushes on every share change the same way.

Either path then removes the old delegation
(`stakingKeeper.RemoveDelegation` — not `Unbond`/`Undelegate`, which would
also shrink the validator's total tokens/shares) and re-creates the shares
under the new address with `SetDelegation`, which alone moves no coins and
doesn't touch the validator's aggregate tokens/shares — only the
bookkeeping key changes. Finally, `migrateMultiStakingLock` moves the
backing `MultiStakingLock` (the escrowed original-denom coin a bond-coin
delegation represents), if any, from the old `LockID` to the new one,
merging into an existing lock if the new address already has one for that
validator.

The merge path's hook sequencing is load-bearing, not cosmetic — see the
comment on `migrateDelegationMerge` for exactly which cosmos-sdk source it
mirrors. Getting the order wrong doesn't corrupt state silently; it
surfaces later as a wrong reward calculation or a panic in distribution's
reference-count bookkeeping (`panic("reference count should never exceed 2")`).

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
