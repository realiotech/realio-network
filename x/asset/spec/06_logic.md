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

1. The denom for the token will be derived from `Creator` and `Symbol` with the format of `asset/{Manager}/{Symbol}`
2. Save the token basic information (name, symbol, decimal and description) in the x/bank metadata store
3. Save the token management info (`Manager`, `ExcludedPrivileges` and `AddNewPrivilege`) in the x/asset store.

Note that here we prefixed the token denom with the manager address in order to allow many different creators to create token with the same symbol, differentiate their denom by including in their creator.

## Register a privilege

Each type of privilege enables a set of messages to be executed.

MintPrivilege {
    storeKey
    mintKeeper MintKeeper
    bankKeeper BankKeeper
}

func Wrap

```go
type handler func(ctx, store, msg)


MintPrivilege func Handle (ctx, store, msg) error{
    func 


}

type Privilege interface {
    
    RegisterMsgHandlers(msgType string, MsgHandler func(ctx, store, msg) )
    RegisterCodec()
    RegisterQuerier(queryType string, )
    
}

```