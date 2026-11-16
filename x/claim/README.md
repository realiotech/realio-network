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
  ClaimDelegations     MigrateBankBalance      ClaimTokens
  staking shares,      own bank coins in       ERC-20 balances,
  rewards ledger,      full, plus the          drained in full
  multi-staking lock   hacked amount
                       recorded on chain,
                       paid by the attacker
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
| `MsgMigrateBankBalance` | Sweep own bank coins, plus the hacked amount recorded on chain, paid by the attacker. No amount input. See [Bank claim](#bank-claim). |
| `MsgClaimTokens` | Drain ERC-20 balances. See [ERC-20 claim](#erc-20-claim). |

### Why linking and claiming are separate

The three claim steps fail for unrelated reasons — a paused token contract, a
validator being slashed, a bad contract address — and none of those should
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

All open to everyone.

| RPC | CLI | REST |
| --- | --- | --- |
| `LinkedAddress` | `realio-networkd query claim linked-address [old-address]` | `/realionetwork/claim/v1/linked_address/{old_address}` |
| `HackedAmount` | `realio-networkd query claim hacked-amount [old-address]` | `/realionetwork/claim/v1/hacked_amount/{old_address}` |
| `Admin` | `realio-networkd query claim admin` | `/realionetwork/claim/v1/admin` |

`LinkedAddress` returns `found: false` rather than an error for an unlinked
address. `HackedAmount` returns an empty amount when nothing is recorded or the
payout has already been made.

### Data for the claim dashboard

What a claim dashboard shows for a connected old address, and where each number
comes from:

| Dashboard shows | State | Query |
| --- | --- | --- |
| Linked address | claim | `LinkedAddress(old)` |
| Bank amount to claim | claim + bank | Sweep: bank `SpendableBalances(old)` (`/cosmos/bank/v1beta1/spendable_balances/{address}`). Payout: `HackedAmount(old)`. Total per denom = sweep + payout. |
| Delegation amount | staking + multi-staking | Staking `DelegatorDelegations(old)` (`/cosmos/staking/v1beta1/delegations/{delegator_addr}`), plus multi-staking `MultiStakingLock(old, validator)` (`/realiotech/multistaking/v1/lock/{validator_address}/{multi_staker_address}`) for the original-denom coin behind each delegation. |
| ERC-20 amount to claim | EVM | `balanceOf(old)` on each candidate contract, via JSON-RPC `eth_call` (or `/cosmos/evm/vm/v1/eth_call`). |

Notes that change what the dashboard should count:

- **Bank:** vesting-locked coins are not spendable and are not claimed. The
  payout is only paid out of the attacker's pool; if the pool cannot cover it,
  the claim fails (see [Bank claim](#bank-claim)).
- **Delegations:** skip a delegation where the old address is the operator of
  that validator (case C), and do not count unbonding delegations or
  redelegations; neither is claimed.
- **ERC-20:** no on-chain index says which contracts an address holds, so the
  contract list is supplied from outside. `x/erc20` `TokenPairs`
  (`/cosmos/evm/erc20/v1/token_pairs`) lists registered tokens, but tokens not
  registered there will not appear, so combine it with the admin's off-chain
  list. Leave out pairs that represent a native coin (`contract_owner =
  OWNER_MODULE`): those are claimed as bank balances and `MsgClaimTokens`
  rejects them. `balanceOf` is selector `0x70a08231` followed by the address
  left-padded to 32 bytes.

## State

Addresses are stored as **raw bytes**, not bech32 strings, so every comparison
is a byte comparison — immune to prefix/casing quirks.

| Field | Type | Notes |
| --- | --- | --- |
| `Admin` | `collections.Item[sdk.AccAddress]` | The sole address allowed to submit anything. Set at genesis or by an upgrade handler (`SetAdmin`); no message changes it. Empty means "not configured". |
| `Attacker` | `collections.Item[sdk.AccAddress]` | The account payouts are drawn from. State, not a message field, so no admin slip can aim a payout at a victim. No message changes it. |
| `AddressLinks` | `collections.Map[sdk.AccAddress, sdk.AccAddress]` | old → new. Several old addresses may share one new address; an old address links only once. |
| `HackedAmounts` | `collections.Map[(sdk.AccAddress, denom), math.Int]` | What the attacker took from each victim and has not yet repaid, keyed by old address and denom. Set at genesis or by an upgrade handler (`SetHackedAmount`); no message writes it. A payout deletes the entry. |

## Delegation claim

`MsgClaimDelegations` re-keys every active delegation of the old address to the
new one. Shares are rewritten under a new owner; nothing unbonds or rebonds, so
voting power and each validator's total stake are untouched. Per
`(old, validator)` pair:

| Case | What happens |
| --- | --- |
| **A** — the new address does not delegate to that validator | Reward ledger carried over verbatim, shares re-keyed. Rewards accrue as before. |
| **B** — the new address already delegates to it | Positions merge; both sides' pending rewards are paid out to the new address. |
| **C** — the old address is that validator's own operator | Skipped, with a `skipped_self_delegation` event. Needs a deliberate follow-up (`app/upgrades/v1.8/validator_rotation.go`). |

Also: the old address's distribution withdraw address is redirected to the new
one (permanently); the backing multi-staking lock moves with the delegation;
unbonding delegations and in-flight redelegations stay behind and mature to the
old address.

Why each choice, and the hook ordering that makes B correct:
[docs/delegation-claim.md](docs/delegation-claim.md). Read it before changing
`keeper/migrate.go`.

## Bank claim

```
MsgMigrateBankBalance{ admin, old_address }        ← no amount

  1 sweep    old_address ──> new_address   everything spendable
  2 payout   attacker    ──> new_address   the hacked amount recorded for old_address
```

Both parts land together or not at all.

- **Sweep** takes the whole balance; coins a vesting schedule still locks stay
  put. An empty address sweeps successfully *for zero*. On mainnet this is the
  only way LMX and RST come back: the attacker holds neither.
- **Payout** pays the hacked amount recorded in `HackedAmounts` out of the attacker,
  then deletes the record. The figure comes from the incident analysis and goes
  on chain at genesis or in an upgrade handler (`SetHackedAmount`), so there is
  nothing for an admin to mistype. No record (79 mainnet victims lost nothing)
  means sweep only.
- **Safe to send again.** The sweep picks up coins that arrived since; the
  payout already deleted its record and pays nothing more. If the attacker's
  pool cannot cover a record, the message fails and the record stays, so it can
  be retried after a top-up.
- **The attacker is state, not input.** The message also refuses the attacker
  as `old_address`, otherwise the sweep would move the whole pool to the
  attacker's link.

The parts stay two keeper calls, two events and two response fields, so what
came back from where is never ambiguous: the sweep moves the holder's own
coins, the payout draws on one finite pool shared across every victim.

### An empty address sweeps for zero

Of 2,454 mainnet victims, 2,373 need both parts, 79 need only the sweep, and
one needs only the payout:

```
realio1zypukaapv7etkevjm8ur2t5hj5ww0uu9lvma09   owed 58,041.338 RIO, bank empty
```

If an empty sweep failed, that address could never be paid. (`ClaimTokens`
likewise skips a contract the holder holds none of.)

### Why both parts hop through the module account

`BlacklistSendRestriction` blocks a blacklisted address paying a **user**
account, and here the old address and the attacker are both frozen while the
destination is a plain account. It exempts payments *into* a module account,
because routine protocol bookkeeping debits blacklisted balances that way. So
each part sends in two hops (`keeper/bank.go`):

```
frozen ──SendCoinsFromAccountToModule──> claim module acct ──SendCoinsFromModuleToAccount──> new
       (allowed: destination is a module account)          (allowed: sender isn't blacklisted)
```

Nothing is weakened for anyone else: only this module can spend from its module
account, and only here, in the same transaction, for what it just escrowed.

`x/asset`'s restriction clears the same way (it skips transfers touching a
module account), so a permissioned token lands on the new address whether or
not that address may hold it. See [x/asset authorization](#x-asset-authorization).

## ERC-20 claim

`MsgClaimTokens` runs each transfer inside the EVM **with the old address as
the caller**.

- Each listed contract moves its **entire** balance; a contract where the old
  address holds nothing is skipped.
- All or nothing: one bad contract fails the whole message.
- The admin lists the contracts, found off-chain — the EVM cannot enumerate an
  address's holdings.
- Rejected up front: a contract that is the ERC-20 representation of a native
  coin (use `MsgMigrateBankBalance`) and an address with no contract code.
- Success check: the old address's balance reads zero afterwards.
- A blacklisted old address can be swept. There is no ERC-721 support.
- Explorers do not show these transfers (no Ethereum receipt); verify via the
  [events](#events) or balances.

Details: [docs/erc20-claim.md](docs/erc20-claim.md). Read it before changing
`keeper/evm.go` or debugging a sweep.

## x/asset authorization

**A recovery does not authorize the new address for anything.** If a claimed
denom is a permissioned `x/asset` Token, the new address ends up holding it
and still off that Token's authorized list — so it cannot send it on. This
applies to both `MsgMigrateBankBalance` and the multi-staking lock that comes
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

## What the module leaves alone

Vesting schedules, governance votes already cast, `x/asset` authorization, and
anything not named in a claim message. It also does not check whether
`new_address` is itself blacklisted — that is `x/blacklist`'s job.

## Genesis

Plain JSON (not protobuf), mirroring `x/blacklist`:

```json
{
  "admin": "realio1...",
  "attacker": "realio1...",
  "links": [{ "old_address": "realio1...", "new_address": "realio1..." }],
  "hacked_amounts": [
    { "address": "realio1...", "amount": [{ "denom": "ario", "amount": "58041338000000000000000" }] }
  ]
}
```

- `admin: ""` — unset; every message is rejected until one is configured.
- `attacker: ""` — unset; a payout fails until one is configured.
- `hacked_amounts` — only what is still unpaid; an export after payouts omits
  what was already repaid.

Validation rejects malformed addresses, `old == new`, `attacker == admin`,
duplicate `old_address` entries, and in `hacked_amounts` a malformed or empty
amount, a duplicate address, or the attacker itself.

## Events

| Event | Attributes | When |
| --- | --- | --- |
| `skipped_self_delegation` | `old_address`, `validator_address` | A validator's own self-bond was left in place (case C). |
| `migrated_bank_balance` | `old_address`, `new_address`, `swept` | The sweep part moved something. |
| `claimed_from_attacker` | `old_address`, `new_address`, `attacker`, `amount` | The payout part moved something. |
| `claimed_erc20` | `contract`, `old_address`, `new_address`, `amount` | An ERC-20 balance was swept. Contracts holding nothing emit nothing. |

## Module wiring

- The v1.8 upgrade mounts the store (`app/upgrades/v1.8/constants.go`), the
  first release to ship the module.
- The module account has no mint or burn rights and serves only as the escrow
  waypoint for `MsgMigrateBankBalance`. Adding it to `maccPerms` also puts it in
  `x/blacklist`'s and `x/asset`'s allow lists, which makes both hops legal; the
  account is created lazily on first use.

## Tests

The EVM test config needs a build tag; without it the suite panics in
`app.Setup`:

```sh
go test -tags=test ./x/claim/...
```

`keeper/mainnet_recovery_test.go` is the end-to-end rehearsal against a real
mainnet export: it links and claims every sampled victim on all three asset
kinds and checks the attacker ends at exactly zero.
