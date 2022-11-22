<!--
order: 2
-->

# State

## State Objects

The `x/asset` module keeps the following objects in state:

| State Object         | Description                    | Key                      | Value           | Store |
|----------------------|--------------------------------|--------------------------| --------------- |-------|
| `Token`              | Token bytecode                 | `[]byte{1} + []byte(id)` | `[]byte{token}` | KV    |
| `TokenAuthorization` | Token Authorization bytecode   | `[]byte{2} + []byte(id)` | `[]byte(id)`    | KV    |

### Token 

Allows creation of tokens with optional user authorization.  

```go
type Token struct {
    Name                  string                         `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
    Symbol                string                         `protobuf:"bytes,2,opt,name=symbol,proto3" json:"symbol,omitempty"`
    Total                 int64                          `protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
    Decimals              int64                          `protobuf:"varint,4,opt,name=decimals,proto3" json:"decimals,omitempty"`
    AuthorizationRequired bool                           `protobuf:"varint,5,opt,name=authorizationRequired,proto3" json:"authorizationRequired,omitempty"`
    Creator               string                         `protobuf:"bytes,6,opt,name=creator,proto3" json:"creator,omitempty"`
    Authorized            map[string]*TokenAuthorization `protobuf:"bytes,7,rep,name=authorized,proto3" json:"authorized,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
    Created               int64                          `protobuf:"varint,8,opt,name=created,proto3" json:"created,omitempty"`
}
```

### Token Authorization

A Token authorization struct represents a single addresses current authorization state for a token

```go
type TokenAuthorization struct {
    TokenSymbol string `protobuf:"bytes,1,opt,name=tokenSymbol,proto3" json:"tokenSymbol,omitempty"`
    Address     string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
    Authorized  bool   `protobuf:"varint,3,opt,name=authorized,proto3" json:"authorized,omitempty"`
}
```



## Genesis State

The `x/asset` module's `GenesisState` defines the state necessary for initializing the chain from a previous exported height. It contains the module parameters and the registered token pairs :

```go
// GenesisState defines the module's genesis state.
type GenesisState struct {
    Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
    PortId string `protobuf:"bytes,2,opt,name=port_id,json=portId,proto3" json:"port_id,omitempty"`
}
```