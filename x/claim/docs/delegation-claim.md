# Delegation claim

Design notes for `MsgClaimDelegations`. Read before changing `keeper/migrate.go`.
The summary is in the [README](../README.md#delegation-claim).

`MsgClaimDelegations` runs `Keeper.MigrateDelegations` (`keeper/migrate.go`),
which walks every active delegation the old address holds and moves it. The
move is a pure re-key: shares are rewritten under a new owner in the staking
store, nothing unbonds, nothing rebonds. **Voting power, each validator's
total stake, and the unbonding/redelegation queues are untouched.**

Before the walk it points the old address's distribution **withdraw address**
at the new one, once per call — so any reward this claim pays out lands on the
new address, never the compromised one. (The redirect is permanent and
account-wide, not scoped to this call: later withdrawals go to the new address
too, including from a self-bond left behind by Path C below.)

For each `(old, validator)` delegation, one of three things happens:

## Path A — fresh (the common case)

The new address does not yet delegate to that validator. The reward-accrual
ledger (`DelegatorStartingInfo` — which historical period the delegation last
settled against) is carried over **verbatim** to the new address, then the
shares are re-keyed. Nothing is paid out early, nothing is lost: the new
address withdraws later via the ordinary `MsgWithdrawDelegatorReward` for
exactly what had accrued.

This path deliberately bypasses distribution's hooks — every one of them would
either flush rewards now or reset the ledger to a zero baseline.

## Path B — merge

The new address *already* delegates to that validator, so the two positions
must combine into one delegation record — and one record can hold only one
reward ledger. Two ledgers cannot be combined losslessly, so this path flushes:
both sides' pending rewards are paid out (to the new address) and the merged
position starts fresh, driven by distribution's
`BeforeDelegationSharesModified` / `AfterDelegationModified` hooks.

This is not a weaker guarantee than an ordinary delegation gets — native
`x/staking` flushes the same way on every share change.

> The hook ordering here is load-bearing, not cosmetic. Getting it wrong
> doesn't corrupt state quietly; it surfaces later as a wrong reward figure or
> a panic in distribution's reference counting. The exact cosmos-sdk code it
> mirrors is documented on `migrateDelegationMerge` in `keeper/migrate.go`.

## Path C — skipped: a validator's own self-delegation

If the old address *is* the validator's operator account, that delegation is
left exactly where it is and a `skipped_self_delegation` event is emitted.

Cosmos-sdk enforces `MinSelfDelegation` and jailing only inside `Unbond`, keyed
off the delegator being the operator — and this claim never calls `Unbond`.
Moving self-bond shares to the new address would silently and permanently
defeat that safety net, since the new address is never the operator. Such a
validator needs a deliberate follow-up (see `app/upgrades/v1.8/validator_rotation.go`).

Only that one `(operator, own validator)` pair is skipped — the same account's
delegations to *other* validators migrate normally.

## Either way: shares and locks

Paths A and B both finish by removing the old delegation with
`RemoveDelegation` — **not** `Unbond`/`Undelegate`, which would also shrink the
validator's tokens and shares — and re-creating the shares under the new
address with `SetDelegation`. No coins move; only the bookkeeping key changes.

Then `migrateMultiStakingLock` moves the backing `MultiStakingLock` (the
escrowed original-denom coin a bond-coin delegation represents) to the new
address's lock ID, merging into an existing lock if there is one. Delegations
funded directly in the native bond denom have no lock and skip this.

## Deliberately left behind

**Unbonding delegations and in-flight redelegations are untouched.** They
mature and pay out to the *old* address exactly as they would have otherwise.

That is safe: `BlacklistSendRestriction` only blocks transfers *from* a
blacklisted address, so a module account paying matured unbonding funds *to*
one still succeeds.

It is also the lower-risk choice. `x/multi-staking`'s unbonding payout
correlates its `MultiStakingUnlock` with the native `UnbondingDelegation`
purely by matching `(delegator, validator, creation height)` at maturity, with
no independent cross-reference. Re-keying one side without the other in
lockstep would either silently orphan the lock or panic `EndBlocker` with
`"unlock entry not found"`. Claiming active delegations only avoids that
failure mode entirely.
