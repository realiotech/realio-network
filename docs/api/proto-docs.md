<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [realionetwork/asset/params.proto](#realionetwork/asset/params.proto)
    - [Params](#realionetwork.asset.Params)
  
- [realionetwork/asset/genesis.proto](#realionetwork/asset/genesis.proto)
    - [GenesisState](#realionetwork.asset.GenesisState)
  
- [realionetwork/asset/packet.proto](#realionetwork/asset/packet.proto)
    - [AssetPacketData](#realionetwork.asset.AssetPacketData)
    - [NoData](#realionetwork.asset.NoData)
  
- [realionetwork/asset/query.proto](#realionetwork/asset/query.proto)
    - [QueryParamsRequest](#realionetwork.asset.QueryParamsRequest)
    - [QueryParamsResponse](#realionetwork.asset.QueryParamsResponse)
  
    - [Query](#realionetwork.asset.Query)
  
- [realionetwork/asset/token.proto](#realionetwork/asset/token.proto)
    - [Token](#realionetwork.asset.Token)
    - [Token.AuthorizedEntry](#realionetwork.asset.Token.AuthorizedEntry)
    - [TokenAuthorization](#realionetwork.asset.TokenAuthorization)
  
- [realionetwork/asset/tx.proto](#realionetwork/asset/tx.proto)
    - [MsgAuthorizeAddress](#realionetwork.asset.MsgAuthorizeAddress)
    - [MsgAuthorizeAddressResponse](#realionetwork.asset.MsgAuthorizeAddressResponse)
    - [MsgCreateToken](#realionetwork.asset.MsgCreateToken)
    - [MsgCreateTokenResponse](#realionetwork.asset.MsgCreateTokenResponse)
    - [MsgTransferToken](#realionetwork.asset.MsgTransferToken)
    - [MsgTransferTokenResponse](#realionetwork.asset.MsgTransferTokenResponse)
    - [MsgUnAuthorizeAddress](#realionetwork.asset.MsgUnAuthorizeAddress)
    - [MsgUnAuthorizeAddressResponse](#realionetwork.asset.MsgUnAuthorizeAddressResponse)
    - [MsgUpdateToken](#realionetwork.asset.MsgUpdateToken)
    - [MsgUpdateTokenResponse](#realionetwork.asset.MsgUpdateTokenResponse)
  
    - [Msg](#realionetwork.asset.Msg)
  
- [realionetwork/mint/v1beta1/mint.proto](#realionetwork/mint/v1beta1/mint.proto)
    - [Minter](#realionetwork.mint.v1beta1.Minter)
    - [Params](#realionetwork.mint.v1beta1.Params)
  
- [realionetwork/mint/v1beta1/genesis.proto](#realionetwork/mint/v1beta1/genesis.proto)
    - [GenesisState](#realionetwork.mint.v1beta1.GenesisState)
  
- [realionetwork/mint/v1beta1/query.proto](#realionetwork/mint/v1beta1/query.proto)
    - [QueryAnnualProvisionsRequest](#realionetwork.mint.v1beta1.QueryAnnualProvisionsRequest)
    - [QueryAnnualProvisionsResponse](#realionetwork.mint.v1beta1.QueryAnnualProvisionsResponse)
    - [QueryInflationRequest](#realionetwork.mint.v1beta1.QueryInflationRequest)
    - [QueryInflationResponse](#realionetwork.mint.v1beta1.QueryInflationResponse)
    - [QueryParamsRequest](#realionetwork.mint.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#realionetwork.mint.v1beta1.QueryParamsResponse)
  
    - [Query](#realionetwork.mint.v1beta1.Query)
  
- [Scalar Value Types](#scalar-value-types)



<a name="realionetwork/asset/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## realionetwork/asset/params.proto



<a name="realionetwork.asset.Params"></a>

### Params
Params defines the parameters for the module.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="realionetwork/asset/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## realionetwork/asset/genesis.proto



<a name="realionetwork.asset.GenesisState"></a>

### GenesisState
GenesisState defines the asset module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#realionetwork.asset.Params) |  |  |
| `port_id` | [string](#string) |  | this line is used by starport scaffolding # genesis/proto/state |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="realionetwork/asset/packet.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## realionetwork/asset/packet.proto



<a name="realionetwork.asset.AssetPacketData"></a>

### AssetPacketData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `noData` | [NoData](#realionetwork.asset.NoData) |  | this line is used by starport scaffolding # ibc/packet/proto/field |






<a name="realionetwork.asset.NoData"></a>

### NoData






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="realionetwork/asset/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## realionetwork/asset/query.proto



<a name="realionetwork.asset.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is request type for the Query/Params RPC method.






<a name="realionetwork.asset.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#realionetwork.asset.Params) |  | params holds all the parameters of this module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="realionetwork.asset.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#realionetwork.asset.QueryParamsRequest) | [QueryParamsResponse](#realionetwork.asset.QueryParamsResponse) | Parameters queries the parameters of the module.

this line is used by starport scaffolding # 2 | GET|/realionetwork/asset/params|

 <!-- end services -->



<a name="realionetwork/asset/token.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## realionetwork/asset/token.proto



<a name="realionetwork.asset.Token"></a>

### Token



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `total` | [int64](#int64) |  |  |
| `decimals` | [int64](#int64) |  |  |
| `authorizationRequired` | [bool](#bool) |  |  |
| `creator` | [string](#string) |  |  |
| `authorized` | [Token.AuthorizedEntry](#realionetwork.asset.Token.AuthorizedEntry) | repeated |  |
| `created` | [int64](#int64) |  |  |






<a name="realionetwork.asset.Token.AuthorizedEntry"></a>

### Token.AuthorizedEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [TokenAuthorization](#realionetwork.asset.TokenAuthorization) |  |  |






<a name="realionetwork.asset.TokenAuthorization"></a>

### TokenAuthorization



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokenSymbol` | [string](#string) |  |  |
| `address` | [string](#string) |  |  |
| `authorized` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="realionetwork/asset/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## realionetwork/asset/tx.proto



<a name="realionetwork.asset.MsgAuthorizeAddress"></a>

### MsgAuthorizeAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `address` | [string](#string) |  |  |






<a name="realionetwork.asset.MsgAuthorizeAddressResponse"></a>

### MsgAuthorizeAddressResponse







<a name="realionetwork.asset.MsgCreateToken"></a>

### MsgCreateToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `total` | [int64](#int64) |  |  |
| `decimals` | [int64](#int64) |  |  |
| `authorizationRequired` | [bool](#bool) |  |  |






<a name="realionetwork.asset.MsgCreateTokenResponse"></a>

### MsgCreateTokenResponse







<a name="realionetwork.asset.MsgTransferToken"></a>

### MsgTransferToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `from` | [string](#string) |  |  |
| `to` | [string](#string) |  |  |
| `amount` | [int64](#int64) |  |  |






<a name="realionetwork.asset.MsgTransferTokenResponse"></a>

### MsgTransferTokenResponse







<a name="realionetwork.asset.MsgUnAuthorizeAddress"></a>

### MsgUnAuthorizeAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `address` | [string](#string) |  |  |






<a name="realionetwork.asset.MsgUnAuthorizeAddressResponse"></a>

### MsgUnAuthorizeAddressResponse







<a name="realionetwork.asset.MsgUpdateToken"></a>

### MsgUpdateToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `authorizationRequired` | [bool](#bool) |  |  |






<a name="realionetwork.asset.MsgUpdateTokenResponse"></a>

### MsgUpdateTokenResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="realionetwork.asset.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateToken` | [MsgCreateToken](#realionetwork.asset.MsgCreateToken) | [MsgCreateTokenResponse](#realionetwork.asset.MsgCreateTokenResponse) |  | |
| `UpdateToken` | [MsgUpdateToken](#realionetwork.asset.MsgUpdateToken) | [MsgUpdateTokenResponse](#realionetwork.asset.MsgUpdateTokenResponse) |  | |
| `AuthorizeAddress` | [MsgAuthorizeAddress](#realionetwork.asset.MsgAuthorizeAddress) | [MsgAuthorizeAddressResponse](#realionetwork.asset.MsgAuthorizeAddressResponse) |  | |
| `UnAuthorizeAddress` | [MsgUnAuthorizeAddress](#realionetwork.asset.MsgUnAuthorizeAddress) | [MsgUnAuthorizeAddressResponse](#realionetwork.asset.MsgUnAuthorizeAddressResponse) |  | |
| `TransferToken` | [MsgTransferToken](#realionetwork.asset.MsgTransferToken) | [MsgTransferTokenResponse](#realionetwork.asset.MsgTransferTokenResponse) | this line is used by starport scaffolding # proto/tx/rpc | |

 <!-- end services -->



<a name="realionetwork/mint/v1beta1/mint.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## realionetwork/mint/v1beta1/mint.proto



<a name="realionetwork.mint.v1beta1.Minter"></a>

### Minter
Minter represents the minting state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `inflation` | [string](#string) |  | current annual inflation rate |
| `annual_provisions` | [string](#string) |  | current annual expected provisions |






<a name="realionetwork.mint.v1beta1.Params"></a>

### Params
Params holds parameters for the mint module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `mint_denom` | [string](#string) |  | type of coin to mint |
| `inflation_rate_change` | [string](#string) |  | maximum annual change in inflation rate |
| `inflation_max` | [string](#string) |  | maximum inflation rate |
| `inflation_min` | [string](#string) |  | minimum inflation rate |
| `goal_bonded` | [string](#string) |  | goal of percent bonded atoms |
| `blocks_per_year` | [uint64](#uint64) |  | expected blocks per year |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="realionetwork/mint/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## realionetwork/mint/v1beta1/genesis.proto



<a name="realionetwork.mint.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the mint module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minter` | [Minter](#realionetwork.mint.v1beta1.Minter) |  | minter is a space for holding current inflation information. |
| `params` | [Params](#realionetwork.mint.v1beta1.Params) |  | params defines all the paramaters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="realionetwork/mint/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## realionetwork/mint/v1beta1/query.proto



<a name="realionetwork.mint.v1beta1.QueryAnnualProvisionsRequest"></a>

### QueryAnnualProvisionsRequest
QueryAnnualProvisionsRequest is the request type for the
Query/AnnualProvisions RPC method.






<a name="realionetwork.mint.v1beta1.QueryAnnualProvisionsResponse"></a>

### QueryAnnualProvisionsResponse
QueryAnnualProvisionsResponse is the response type for the
Query/AnnualProvisions RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `annual_provisions` | [bytes](#bytes) |  | annual_provisions is the current minting annual provisions value. |






<a name="realionetwork.mint.v1beta1.QueryInflationRequest"></a>

### QueryInflationRequest
QueryInflationRequest is the request type for the Query/Inflation RPC method.






<a name="realionetwork.mint.v1beta1.QueryInflationResponse"></a>

### QueryInflationResponse
QueryInflationResponse is the response type for the Query/Inflation RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `inflation` | [bytes](#bytes) |  | inflation is the current minting inflation value. |






<a name="realionetwork.mint.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="realionetwork.mint.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#realionetwork.mint.v1beta1.Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="realionetwork.mint.v1beta1.Query"></a>

### Query
Query provides defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#realionetwork.mint.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#realionetwork.mint.v1beta1.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/realionetwork/v1beta1/params|
| `Inflation` | [QueryInflationRequest](#realionetwork.mint.v1beta1.QueryInflationRequest) | [QueryInflationResponse](#realionetwork.mint.v1beta1.QueryInflationResponse) | Inflation returns the current minting inflation value. | GET|/realionetwork/mint/v1beta1/inflation|
| `AnnualProvisions` | [QueryAnnualProvisionsRequest](#realionetwork.mint.v1beta1.QueryAnnualProvisionsRequest) | [QueryAnnualProvisionsResponse](#realionetwork.mint.v1beta1.QueryAnnualProvisionsResponse) | AnnualProvisions current minting annual provisions value. | GET|/realionetwork/mint/v1beta1/annual_provisions|

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
