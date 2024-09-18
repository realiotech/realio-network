<!--
order: 6
-->

# Logic

This file describes the core logics in this module.

## Token creation process

This process is triggered by `MsgCreateToken`.

Validation:

- Check if `Creator` is whitelisted. We only allow some certain accounts to create tokens, these accounts is determined via gov proposal.
- Check if the token with the same denom has already existed.

Flow:

1. The denom for the token will be derived from `Creator` and `Symbol` with the format of `asset/{Creator}/{Symbol}`
2. Save the token basic information (name, symbol, decimal and description) in the x/bank metadata store
3. Save the token management info (`Manager`, `ExcludedPrivileges` and `AddNewPrivilege`) in the x/asset store.

Note that here we prefixed the token denom with the manager address in order to allow many different creators to create token with the same symbol, differentiate their denom by including in their creator.

## Register a privilege

To intergrate with the `asset module` Each type of privilege has to implement this interface

```go
type Privilege interface {
    RegisterInterfaces()
    MsgHandler() MsgHandler
    QueryHandler() QueryHandler
}

type MsgHandler func(context Context, privMsg PrivilegeMsg, tokenID string, privAcc sdk.AccAddress) error

type QueryHandler func(context Context, privQuery PrivilegeQuery, tokenID string) error
```

This interface provides all the functionality necessary for a privilege, including a message handler, query handler and cli

All the `PrivilegeMsg` of a privilege should return the name of that privilege when called `NeedPrivilege()`. A message handler should handle all the `PrivilegeMsg` of that privilege.

When adding a `Privilege`, we calls `PrivilegeManager.AddPrivilege()` in `app.go` which inturn maps all the `PrivilegeMsg` of that privilege to its `MsgHandler`. This mapping logic will later be used when running a `MsgExecutePrivilege`

## Flow of MsgExecutePrivilege

This process is triggered by the `MsgExecutePrivilege`.

Validation:

- Checks if the token specified in the msg exists.
- Checks if the privilege is supported.
- Checks if the `Msg.Address` has the corresponding `Privilege` specified by `PrivilegeMsg.NeedPrivilege()`

Flow:

- `PrivilegeManager` routes the `PrivilegeMsg` to the its `MsgHandler`.
- `MsgHandler` now handles the `PrivilegeMsg`.
