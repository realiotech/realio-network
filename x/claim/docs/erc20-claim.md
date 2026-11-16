# ERC-20 claim

Design notes for `MsgClaimTokens`. Read before changing `keeper/evm.go` or debugging a
sweep. The summary is in the [README](../README.md#erc-20-claim).

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

## Why the contracts are listed by hand

The EVM has no way to answer "which contracts does this address hold a balance
in" — balances live in each contract's own storage under a layout only that
contract knows. So the contracts are supplied by the admin and discovered
off-chain.

## Blacklisted addresses can be swept

`x/blacklist`'s EVM guard is an EVM *hook* (`PostTxProcessing`), and hooks fire
only on the msg-server transaction path — never on a module's own `CallEVM`. A
blacklisted old address can therefore be swept, which is the whole point. The
ante decorators are likewise irrelevant: the admin signs this transaction, not
the holder.

## What it refuses

Both cases are rejected up front because neither would fail loudly on its own.

| Case | What would otherwise happen |
| --- | --- |
| A contract registered in `x/erc20` as the representation of a **native coin** | It is a precompile, not bytecode: "transferring" it runs bank's `SendCoins` and surfaces as a bare `address is blacklisted`, with no hint the balance was a bank balance. Use `MsgMigrateBankBalance`. |
| An address with no contract code | A call to a codeless address does **not** revert in the EVM — it succeeds and returns nothing, so a mistyped address would only show up later as an unpack failure somewhere less obvious. |

## Verification

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
> events in the [README](../README.md#events), or by querying balances.

One side effect worth knowing: if the old address has no auth account —
possible for an address that only ever *received* tokens — one is created,
exactly as the first transaction it sent would have done.

## No NFT support

There is no ERC-721 path. The chain has no NFT module, no deployed ERC-721 and
no registry, and the standard offers no way to list an owner's tokens unless
the contract implements `ERC721Enumerable` — so every token ID would have to be
supplied by hand anyway. Adding it later is a self-contained change: a second
ABI and a `transferFrom` call alongside `MigrateERC20`.
