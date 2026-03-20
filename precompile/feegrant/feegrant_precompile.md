

# Feegrant Precompiles

The feegrant module provides an Ethereum precompile that enables MetaMask and other EVM wallets to interact with the Cosmos SDK feegrant functionality directly through smart contract calls. This precompile bridges the gap between EVM-based applications and the Cosmos SDK feegrant module.

## Overview

The feegrant precompile allows users to:
- Grant fee allowances from a granter to a grantee (Basic, Periodic, or AllowedMsg allowances)
- Revoke existing fee allowances

## Precompile Address

The feegrant precompile is deployed at the following address:

```
0x0000000000000000000000000000000000000901
```

This address can be used to interact with the feegrant functionality from MetaMask, web3 applications, or any EVM-compatible wallet.

## Precompile Interface

The precompile implements the `IFeeGrant` interface defined in Solidity:

```solidity
interface IFeeGrant {
    function grant(
        address grantee,
        string calldata spendLimit,
        string calldata expiration,
        int64 period,
        string calldata periodLimit,
        string[] calldata allowedMessages
    ) external returns (bool success);

    function revoke(
        address grantee
    ) external returns (bool success);
}
```

## Transaction Methods

### grant

Grants a fee allowance from the caller (granter) to a grantee. The allowance type is determined automatically based on the provided parameters, following the same logic as the Cosmos SDK CLI `NewCmdFeeGrant` command:

- **BasicAllowance**: Created when no `period` is set
- **PeriodicAllowance**: Created when `period > 0` and `periodLimit` is provided
- **AllowedMsgAllowance**: Wraps the above allowance when `allowedMessages` is non-empty

**Parameters:**
- `grantee` (address): The Ethereum address of the grantee. Currently tx will be failed if grantee was not inited or received any fund.
- `spendLimit` (string): The maximum amount of coins that can be spent, as a comma-separated string (e.g. `"1000000ario"`). Empty string for unlimited
- `expiration` (string): The expiration time in RFC3339 format (e.g. `"2026-03-22T00:00:00Z"`). Empty string for no expiration
- `period` (int64): The period duration in seconds. `0` for no periodic allowance
- `periodLimit` (string): The maximum amount of coins per period, as a comma-separated string (e.g. `"1000000ario"`). Empty string if no periodic allowance
- `allowedMessages` (string[]): The list of allowed message type URLs (e.g. `["/cosmos.evm.vm.v1.MsgEthereumTx"]`). Empty array for no restriction

**Returns:**
- `success` (bool): True if the grant was successful

**Logic Flow:**
1. Parses and validates all input parameters
2. Constructs a `BasicAllowance` with optional `spendLimit` and `expiration`
3. If `period > 0`, wraps in a `PeriodicAllowance` with `periodLimit`
4. If `allowedMessages` is non-empty, wraps in an `AllowedMsgAllowance`
5. Builds a `MsgGrantAllowance` and executes it through the feegrant message server
6. Returns success status

### revoke

Revokes an existing fee allowance from the caller (granter) to a grantee.

**Parameters:**
- `grantee` (address): The Ethereum address of the grantee whose allowance should be revoked

**Returns:**
- `success` (bool): True if the revocation was successful

**Logic Flow:**
1. Parses the grantee address from the arguments
2. Builds a `MsgRevokeAllowance` with the caller as granter
3. Executes the revocation through the feegrant message server
4. Returns success status

## Allowance Types

### BasicAllowance

The simplest allowance type. Optionally limits the total amount the grantee can spend and/or sets an expiration time.

| Parameter | Effect |
|-----------|--------|
| `spendLimit = ""` | Unlimited spending |
| `spendLimit = "100ario"` | Grantee can spend up to 100 ario |
| `expiration = ""` | No expiration |
| `expiration = "2026-12-31T23:59:59Z"` | Expires at the specified time |

### PeriodicAllowance

Extends `BasicAllowance` with a periodic reset of the spend limit. The grantee can spend up to `periodLimit` per `period`.

| Parameter | Effect |
|-----------|--------|
| `period = 3600` | Resets every 3600 seconds (1 hour) |
| `periodLimit = "10ario"` | Grantee can spend up to 10 ario per period |

### AllowedMsgAllowance

Wraps any of the above allowances and restricts the grant to only cover fees for specific message types.

| Parameter | Effect |
|-----------|--------|
| `allowedMessages = ["/cosmos.evm.vm.v1.MsgEthereumTx"]` | Only covers fees for EVM transactions |
| `allowedMessages = ["/cosmos.bank.v1beta1.MsgSend"]` | Only covers fees for bank send messages |

## Usage Examples

### Grant unlimited fee allowance

```javascript
const feegrant = new ethers.Contract("0x0000000000000000000000000000000000000901", IFeeGrantABI, signer);
await feegrant.grant(granteeAddress, "", "", 0, "", []);
```

### Grant with spend limit and expiration

```javascript
await feegrant.grant(granteeAddress, "1000000ario", "2026-12-31T23:59:59Z", 0, "", []);
```

### Grant periodic allowance

```javascript
await feegrant.grant(granteeAddress, "", "", 3600, "100000ario", []);
```

### Grant only for EVM transactions

```javascript
await feegrant.grant(granteeAddress, "", "", 0, "", ["/cosmos.evm.vm.v1.MsgEthereumTx"]);
```

### Revoke a fee allowance

```javascript
await feegrant.revoke(granteeAddress);
```

## Integration with MetaMask

The precompile enables seamless integration with MetaMask and other EVM wallets:

1. **Contract Interaction**: Users can interact with the precompile as if it were a regular smart contract
2. **Transaction Signing**: All transactions are signed using standard Ethereum transaction signing
3. **Gas Estimation**: Gas costs are calculated and displayed in MetaMask
4. **Fee Sponsorship**: Combined with the feesponsor module, granters can sponsor EVM transaction fees for grantees

