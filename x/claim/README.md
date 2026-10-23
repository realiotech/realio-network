<!--
order: 0
-->

# Claim

`x/claim` is the recovery counterpart to [`x/blacklist`](../blacklist/README.md).

When an address's private key leaks, `x/blacklist` freezes it — the funds stop
moving, but they also stop being usable by their rightful owner. `x/claim` is
how the owner gets them back: after re-doing KYC off-chain and registering a
fresh address, the module's admin **links** old → new, and then claims each
kind of asset across.

```
leaked addr ──freeze──> x/blacklist
     │
     └──re-KYC off-chain──> new addr registered
                                  │
                    admin: MsgLinkAddress(old, new)     ← records the destination
                                  │
         ┌────────────────────────┼────────────────────────┐
  ClaimDelegations          ClaimBalances             ClaimTokens
  staking shares,           bank coins,               ERC-20 balances,
  rewards ledger,           admin states the          drained in full
  multi-staking lock        amount
```

## Messages

Every message is gated on the module's configured admin. There is no
governance route, no self-service route, and no message that changes who the
admin is — a compromised key can never link itself anywhere or claim anything.

| Message | Does |
| --- | --- |
| `MsgLinkAddress` | Record `old_address` → `new_address`. Moves nothing. |
| `MsgLinkAddresses` | Batch of the above under one admin check. |
| `MsgClaimDelegations` | Re-key active delegations. See [Delegation claim](#delegation-claim). |
| `MsgClaimBalances` | Move bank coins the admin names. See [Bank claim](#bank-claim). |
| `MsgClaimTokens` | Drain ERC-20 balances. See [ERC-20 claim](#erc-20-claim). |

### Why linking and claiming are separate

The three claim steps fail for unrelated reasons — a paused token contract, a
validator being slashed, an amount keyed in wrong — and none of those should
be able to roll back the others, or undo the link itself. Keeping them apart
also lets the admin re-run one step after finding more contracts, without
touching anything already recovered.

Atomicity is not lost, only made opt-in: several of these messages in **one
transaction** still land together or not at all.

### Shared gate

Every `Claim*` message passes through the same check (`authorizeClaim`): the
signer must be the admin, and `old_address` must already be linked. **None of
them has a field for the destination** — it is read from the link, so the only
address an admin names is the one being emptied, never where it empties to.

| Rule | Error |
| --- | --- |
| `admin` must equal the configured `Admin` | `unauthorized` (also when no admin is configured) |
| `old_address` must be valid bech32 | `invalid address` |
| `old_address` must already be linked | `is not linked; submit MsgLinkAddress first` |

`MsgLinkAddress` itself checks instead that `old_address` ≠ `new_address`,
that `old_address` is not already linked, and that `new_address` is not itself
someone's `old_address` (no chains — link to the end of the chain instead).

> **No tx CLI.** `GetTxCmd()` returns `nil` — these messages are submitted as
> generated/signed transactions, not through `realio-networkd tx claim ...`.

## Queries

| RPC | CLI | REST |
| --- | --- | --- |
| `LinkedAddress` | `realio-networkd query claim linked-address [old-address]` | `/realionetwork/claim/v1/linked_address/{old_address}` |
| `Admin` | `realio-networkd query claim admin` | `/realionetwork/claim/v1/admin` |

`LinkedAddress` returns `found: false` rather than an error for an unlinked
address. Queries are open to everyone.

## State

Both fields store **raw address bytes**, not bech32 strings, so every
comparison is a byte comparison — immune to the prefix/casing quirks a string
equality check would be sensitive to.

| Field | Type | Notes |
| --- | --- | --- |
| `Admin` | `collections.Item[sdk.AccAddress]` | The sole address allowed to submit anything. No message changes it — it is set at genesis or by a chain-upgrade handler calling `SetAdmin` directly. An empty value reads back as "not configured". |
| `AddressLinks` | `collections.Map[sdk.AccAddress, sdk.AccAddress]` | old → new. Many old addresses may point at one new address (consolidating several leaked wallets); an old address links only once. |

## Delegation claim

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

### Path A — fresh (the common case)

The new address does not yet delegate to that validator. The reward-accrual
ledger (`DelegatorStartingInfo` — which historical period the delegation last
settled against) is carried over **verbatim** to the new address, then the
shares are re-keyed. Nothing is paid out early, nothing is lost: the new
address withdraws later via the ordinary `MsgWithdrawDelegatorReward` for
exactly what had accrued.

This path deliberately bypasses distribution's hooks — every one of them would
either flush rewards now or reset the ledger to a zero baseline.

### Path B — merge

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

### Path C — skipped: a validator's own self-delegation

If the old address *is* the validator's operator account, that delegation is
left exactly where it is and a `skipped_self_delegation` event is emitted.

Cosmos-sdk enforces `MinSelfDelegation` and jailing only inside `Unbond`, keyed
off the delegator being the operator — and this claim never calls `Unbond`.
Moving self-bond shares to the new address would silently and permanently
defeat that safety net, since the new address is never the operator. Such a
validator needs a deliberate follow-up (see `app/migrations/validator_rotation.go`).

Only that one `(operator, own validator)` pair is skipped — the same account's
delegations to *other* validators migrate normally.

### Either way: shares and locks

Paths A and B both finish by removing the old delegation with
`RemoveDelegation` — **not** `Unbond`/`Undelegate`, which would also shrink the
validator's tokens and shares — and re-creating the shares under the new
address with `SetDelegation`. No coins move; only the bookkeeping key changes.

Then `migrateMultiStakingLock` moves the backing `MultiStakingLock` (the
escrowed original-denom coin a bond-coin delegation represents) to the new
address's lock ID, merging into an existing lock if there is one. Delegations
funded directly in the native bond denom have no lock and skip this.

### Deliberately left behind

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

## Bank claim

`MsgClaimBalances` moves bank coins the admin names:

```
MsgClaimBalances{ admin, old_address, amount: [ {denom, amount}, ... ] }
```

The amount is **stated, not read off the chain**. A recovery can then be scoped
deliberately — to what the holder can evidence, or to what a dispute settled
on — instead of always being all-or-nothing. Coins beyond the old address's
*spendable* balance fail the message; a vesting schedule still locks what it
locked.

### Why it hops through the module account

A direct `old → new` send is refused. `BlacklistSendRestriction` blocks a
blacklisted address paying a **user** account, and that is exactly what this
transfer is. It exempts payments *into* a module account, because routine
protocol bookkeeping debits blacklisted balances that way and none of those
callers survive `SendCoins` failing.

So `MoveBalances` (`keeper/bank.go`) sends in two hops:

```
old ──SendCoinsFromAccountToModule──> claim module acct ──SendCoinsFromModuleToAccount──> new
     (allowed: destination is a module account)          (allowed: sender isn't blacklisted)
```

Nothing about the restriction is weakened for anyone else: only this module can
spend from its own module account, and it only ever does so here, in the same
transaction, for the amount it just escrowed.

`x/asset`'s restriction clears for the same reason — it skips any transfer
where either side is a module account. Note what that implies: a permissioned
token lands on the new address whether or not that address may hold it. See
[x/asset authorization](#x-asset-authorization).

## ERC-20 claim

`MsgClaimTokens` runs each transfer inside the EVM **with the old address as
the caller**, so the token contract sees exactly what it would have seen had
the holder sent the transaction themselves.

```
MsgClaimTokens{ admin, old_address, erc20_contracts: ["0xAbc...", "0xDef..."] }
```

Each listed contract moves its **entire** balance — unlike a bank balance
there is nothing to split between the old holder and the new one. A contract
where the old address holds nothing is skipped, not rejected.

All or nothing within the message: one bad contract fails the whole sweep, so
there is never a half-swept address to reconcile against the list submitted.

### Why the contracts are listed by hand

The EVM has no way to answer "which contracts does this address hold a balance
in" — balances live in each contract's own storage under a layout only that
contract knows. So the contracts are supplied by the admin and discovered
off-chain.

### Blacklisting does not get in the way

`x/blacklist`'s EVM guard is an EVM *hook* (`PostTxProcessing`), and hooks fire
only on the msg-server transaction path — never on a module's own `CallEVM`. A
blacklisted old address can therefore be swept, which is the whole point. The
ante decorators are likewise irrelevant: the admin signs this transaction, not
the holder.

### What it refuses

Both cases are rejected up front because neither would fail loudly on its own.

| Case | What would otherwise happen |
| --- | --- |
| A contract registered in `x/erc20` as the representation of a **native coin** | It is a precompile, not bytecode: "transferring" it runs bank's `SendCoins` and surfaces as a bare `address is blacklisted`, with no hint the balance was a bank balance. Use `MsgClaimBalances`. |
| An address with no contract code | A call to a codeless address does **not** revert in the EVM — it succeeds and returns nothing, so a mistyped address would only show up later as an unpack failure somewhere less obvious. |

### Verification

One check, after the transfer: **the old address's balance must read zero.**

It is deliberately not an equality check against what moved — a fee-on-transfer
token delivers less than it debited, which is fine when the goal is emptying
the old address. It also subsumes inspecting `transfer`'s return value: a token
that reports failure without reverting moves nothing, so the balance still
stands there. (Which matters, because ERC-20 predates the `returns (bool)`
requirement and tokens that return nothing at all are still common.)

> **Explorers will not show these transfers.** The EVM logs a `Transfer` event
> as usual, but because the call does not go through `ApplyTransaction` there
> is no Ethereum receipt for `eth_getLogs` to return. Verify against the Cosmos
> events below, or by querying balances.

One side effect worth knowing: if the old address has no auth account —
possible for an address that only ever *received* tokens — one is created,
exactly as the first transaction it sent would have done.

### No NFT support

There is no ERC-721 path. The chain has no NFT module, no deployed ERC-721 and
no registry, and the standard offers no way to list an owner's tokens unless
the contract implements `ERC721Enumerable` — so every token ID would have to be
supplied by hand anyway. Adding it later is a self-contained change: a second
ABI and a `transferFrom` call alongside `MigrateERC20`.

## x/asset authorization

**A recovery does not authorize the new address for anything.** If a claimed
denom is a permissioned `x/asset` Token, the new address ends up holding it
and still off that Token's authorized list — so it cannot send it on. This
applies to both `MsgClaimBalances` and the multi-staking lock that comes
across with `MsgClaimDelegations`.

That is deliberate. The authorized list answers to the **Token's manager**:
`x/asset`'s `MsgAuthorizeAddress` is signed by that manager, a signature
x/claim's admin does not hold. Writing into it from here would let the claim
admin grant a permission `x/asset` never gave them.

So the last step of such a recovery belongs to the token manager:

```
MsgAuthorizeAddress{ manager, symbol, address: <new_address> }
```

Put it in the same transaction as the claim and the two land together.

`keeper/bank_test.go` and `keeper/mainnet_migration_test.go` both pin this —
the latter against real mainnet holders of the real permissioned `rst` token.

## What the module never touches

Vesting schedules, governance votes already cast, `x/asset` authorization, and
anything not named in a claim message. It also does not check whether
`new_address` is itself blacklisted — that is `x/blacklist`'s job.

## Genesis

Plain JSON (not protobuf), mirroring `x/blacklist`:

```json
{
  "admin": "realio1...",
  "links": [{ "old_address": "realio1...", "new_address": "realio1..." }]
}
```

`admin: ""` means unset, and every message is rejected until one is configured.
Validation rejects malformed addresses, `old == new`, and duplicate
`old_address` entries.

## Events

| Event | Attributes | When |
| --- | --- | --- |
| `skipped_self_delegation` | `old_address`, `validator_address` | A validator's own self-bond was left in place (Path C). |
| `claimed_balance` | `old_address`, `new_address`, `amount` | A `MsgClaimBalances` completed. |
| `claimed_erc20` | `contract`, `old_address`, `new_address`, `amount` | An ERC-20 balance was swept. Contracts holding nothing emit nothing. |

## Module wiring

- Consensus version `1`; keeper constructed in `app/app.go`.
- The store is mounted by the **v1.8 upgrade**
  (`app/upgrades/v1.8/constants.go`), the first release to ship the module.
- The module holds a **module account** with no minting or burning rights,
  used only as the escrow waypoint in `MsgClaimBalances`. Adding it to
  `maccPerms` also puts it in `x/blacklist`'s and `x/asset`'s allow lists,
  which is what makes both hops legal; the account itself is created lazily on
  first use.

## Tests

`keeper/msg_server_test.go` covers linking, the admin gate across every
message, both delegation paths, exact reward preservation on the fresh path,
reward flush to the new address on the merge path, the self-delegation skip,
and batch atomicity. `keeper/bank_test.go` covers the balance claim, including
that it succeeds out of a blacklisted address where a direct send is refused,
and that it leaves `x/asset` authorization alone.
`keeper/evm_test.go` covers the ERC-20 sweep against real contracts deployed
into the test EVM, including an address with no auth account and the
all-or-nothing rollback. `keeper/mainnet_migration_test.go` replays the whole
flow against a real mainnet export (`testdata/exported_mainnet_after.json`).

Run them with the build tag the EVM test config requires:

```sh
go test -tags=test ./x/claim/...
```
