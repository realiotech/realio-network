<!--
order: 4
-->

# Messages

The asset module exposes the following messages:

## MessageCreateToken

```go
type MsgCreateToken struct {
    Creator               string  
    Manager               string
    Name                  string
    Symbol                string
    Decimal               string
    ExcludedPrivileges    []string
    AddNewPrivileges      bool
}
```

`MessageCreateToken` allows a whitelisted account to create a token with custom configuration.

## MessageAllocateToken

```go
type MsgAllocateToken struct {
    Manager string
    TokenID string
    Balances []Balance
    VestingBalances []VestingAccount
}
```

`MessageAllocateToken` can only be executed by the token manager once after their token is successfully created. It will allocate tokens (either vesting or liquid) to the list of accounts sepecifed in the message.

## MessageAssignPrivilege

```go
type MsgAssignPrivilege struct {
    Manager string
    TokenID string
    AssignedTo []string
    Privilege string
}
```

`MessageAssignPrivilege` allows the token manager to assign a privilege to the chosen addresses. This message will fail if the privilege is in the list of `ExcludedPrivileges` specified when creating the token.

## MessageUnassignPrivilege

```go
type MsgAssignPrivilege struct {
    Manager string
    TokenID string
    UnassignedFrom []string
    Privilege string
}
```

`MessageUnassignPrivilege` allows the token manager to unassign a privilege from the chosen addresses.

## MessageDisablePrivilege

```go
type MsgDisablePrivilege struct {
    Manager string
    TokenID string
    DisabledPrivilege string 
}
```

`MessageDisablePrivilege` allows the token manager to disable a privilege permanently, it will also unassigns all the accounts with that privilege.

## MessageExecutePrivilege

```go
type PrivilegeMsg interface {
    Privilege()    string
}

type MsgExecute struct {
    Address string
    TokenID string
    PrivilegeMsg  PrivilegeMsg
}
```

`PrivilegeMsg` allows privileged accounts to execute logic of its privilege. For that reason, it has different implementations defined by each types of privilege instead of the `Asset Module`. These implementations and the logic to handle them are registered into the module via `RegisterPrivilege` method.
