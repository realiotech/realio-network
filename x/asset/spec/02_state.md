<!--
order: 2
-->

# State

## Store

### Token Store

Map: `0x00 | {Token ID} -> Token`

```go
type Token struct {
    Name string
    Symbol string
    Decimal string
    Description string
}
```

Note that these infos are also stored in the `metadata` store of bank module.

### Token Management

Map: `0x01 | {Token ID} -> TokenManagement`

Token management holds these information about the token:

* the token's manager
* the excluded privileges (privileges that are permenantly disable)
* if we can add newly introduced privilege to the token later on

```go
type TokenManagement struct {
    Manager               string                         
    AddNewPrivilege       bool
    ExcludedPrivileges    []string
}
```

### Privileged Accounts

Map: `0x02 | {Token ID} | {Privilege Name} -> Addresses`

### Privilege Store

Sub stores: `0x03 | {Token ID} | {Privilege Name}`

Since each type of privilege has its own logic, we need to leave a seprate space for each of them to store their data. A privilege should manage its own store provided by the asset module, prefixed with `0x03 | {Token ID} | {Privilege Name}`

## Genesis State

The `x/asset` module's `GenesisState` defines the state necessary for initializing the chain from a previous exported height. It contains the module parameters and the registered token pairs :

```go
// GenesisState defines the module's genesis state.
type GenesisState struct {
    Params Params
    Tokens []Token
}
```
