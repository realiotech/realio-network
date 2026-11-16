<!--
order: 0
-->

# Blacklist

`x/blacklist` maintains a set of addresses barred from moving funds — the
hardening measure for accounts whose private keys are known to have leaked.
Membership is a simple on-chain set (`Blacklisted`, keyed by raw address
bytes); there is no richer per-entry state.

Entries are added or removed only via `MsgUpdateBlacklist`, gated to the
module's authority (`x/gov` by default). A compromised key can never remove
itself from the blacklist — that requires a governance decision. Entries can
also be seeded directly by a chain-upgrade fork handler (see
`app/migrations`), which is how the initial leaked-key list gets loaded
without a governance vote.

## Enforcement — three independent layers

Being on the list alone does nothing; three separate integration points each
check it, because no single point in the stack sees every way funds or EVM
state can move:

1. **Ante decorators** (`app/ante/blacklist.go`) — `CosmosBlacklistDecorator`
   and `EVMBlacklistDecorator` reject a transaction outright if a blacklisted
   address is one of its signers (including an authz `MsgExec`'s inner
   signers, and EIP-7702 authorization authorities), or is the tx's
   `FeeGranter`. This only ever sees the **outer transaction's** signers/fee
   payer.
2. **Bank send restriction** (`x/blacklist/keeper/restrictions.go`,
   `BlacklistSendRestriction`) — registered on `BankKeeper`, so it runs
   inside every `SendCoins` call regardless of what triggered it: a plain
   `MsgSend`, an authz `MsgExec`'s inner message, an ERC-20 precompile's
   `transferFrom`, IBC, anything. This is the one place all of those
   converge, which is exactly what the ante layer can't see.
3. **EVM token hook** (`app/evm_blacklist_hook.go`,
   `EVMTokenBlacklistHook`) — a post-EVM-execution hook that reverts a
   transaction if a standard ERC-20/721 `Transfer`/`Approval`/
   `ApprovalForAll` event shows a blacklisted owner losing tokens or
   delegating spending authority, including calls routed through arbitrary
   contracts.

### `BlacklistSendRestriction`'s carve-out is blanket, not scoped

The send restriction only checks `fromAddr`, and only when `toAddr` is **not**
a module account (`allowAddrs`, snapshotted once from `app.ModuleAccountAddrs()`
at startup). This exists because routine protocol bookkeeping — multistaking
burning a matured delegator's locked "representation" coin as an unbonding
payout step, for one — moves coins out of a blacklisted address's balance
into a module account automatically, and the calling code isn't written to
handle `SendCoins` failing. Blocking it caused a real `EndBlocker` panic
during testing; that's why the carve-out exists at all.

**The exemption is blanket: a blacklisted address can still successfully
send to *any* module account** (`gov`, `erc20`, `bridge`, ...) — not just
the specific ones behind that bookkeeping. This is safe today because the
actual threat this restriction closes — authz/ERC-20 moving a blacklisted
account's funds to an attacker-controlled wallet — always targets a plain
user account, never a module account, so the broader exemption doesn't
reopen it. It stops being safe the moment any module makes a module
account's balance directly extractable back out to an arbitrary address; if
that's ever true, revisit this carve-out before shipping it.
