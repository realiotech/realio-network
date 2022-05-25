<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [asset/params.proto](#asset/params.proto)
    - [Params](#realiotech.realionetwork.asset.Params)
  
- [asset/genesis.proto](#asset/genesis.proto)
    - [GenesisState](#realiotech.realionetwork.asset.GenesisState)
  
- [asset/packet.proto](#asset/packet.proto)
    - [AssetPacketData](#realiotech.realionetwork.asset.AssetPacketData)
    - [NoData](#realiotech.realionetwork.asset.NoData)
  
- [asset/query.proto](#asset/query.proto)
    - [QueryParamsRequest](#realiotech.realionetwork.asset.QueryParamsRequest)
    - [QueryParamsResponse](#realiotech.realionetwork.asset.QueryParamsResponse)
  
    - [Query](#realiotech.realionetwork.asset.Query)
  
- [asset/token.proto](#asset/token.proto)
    - [Token](#realiotech.realionetwork.asset.Token)
    - [Token.AuthorizedEntry](#realiotech.realionetwork.asset.Token.AuthorizedEntry)
    - [TokenAuthorization](#realiotech.realionetwork.asset.TokenAuthorization)
  
- [asset/tx.proto](#asset/tx.proto)
    - [MsgAuthorizeAddress](#realiotech.realionetwork.asset.MsgAuthorizeAddress)
    - [MsgAuthorizeAddressResponse](#realiotech.realionetwork.asset.MsgAuthorizeAddressResponse)
    - [MsgCreateToken](#realiotech.realionetwork.asset.MsgCreateToken)
    - [MsgCreateTokenResponse](#realiotech.realionetwork.asset.MsgCreateTokenResponse)
    - [MsgTransferToken](#realiotech.realionetwork.asset.MsgTransferToken)
    - [MsgTransferTokenResponse](#realiotech.realionetwork.asset.MsgTransferTokenResponse)
    - [MsgUnAuthorizeAddress](#realiotech.realionetwork.asset.MsgUnAuthorizeAddress)
    - [MsgUnAuthorizeAddressResponse](#realiotech.realionetwork.asset.MsgUnAuthorizeAddressResponse)
    - [MsgUpdateToken](#realiotech.realionetwork.asset.MsgUpdateToken)
    - [MsgUpdateTokenResponse](#realiotech.realionetwork.asset.MsgUpdateTokenResponse)
  
    - [Msg](#realiotech.realionetwork.asset.Msg)
  
- [mint/v1beta1/mint.proto](#mint/v1beta1/mint.proto)
    - [Minter](#realiotech.realionetwork.mint.v1beta1.Minter)
    - [Params](#realiotech.realionetwork.mint.v1beta1.Params)
  
- [mint/v1beta1/genesis.proto](#mint/v1beta1/genesis.proto)
    - [GenesisState](#realiotech.realionetwork.mint.v1beta1.GenesisState)
  
- [mint/v1beta1/query.proto](#mint/v1beta1/query.proto)
    - [QueryAnnualProvisionsRequest](#realiotech.realionetwork.mint.v1beta1.QueryAnnualProvisionsRequest)
    - [QueryAnnualProvisionsResponse](#realiotech.realionetwork.mint.v1beta1.QueryAnnualProvisionsResponse)
    - [QueryInflationRequest](#realiotech.realionetwork.mint.v1beta1.QueryInflationRequest)
    - [QueryInflationResponse](#realiotech.realionetwork.mint.v1beta1.QueryInflationResponse)
    - [QueryParamsRequest](#realiotech.realionetwork.mint.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#realiotech.realionetwork.mint.v1beta1.QueryParamsResponse)
  
    - [Query](#realiotech.realionetwork.mint.v1beta1.Query)
  
- [Scalar Value Types](#scalar-value-types)



<a name="asset/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## asset/params.proto



<a name="realiotech.realionetwork.asset.Params"></a>

### Params
Params defines the parameters for the module.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="asset/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## asset/genesis.proto



<a name="realiotech.realionetwork.asset.GenesisState"></a>

### GenesisState
GenesisState defines the asset module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#realiotech.realionetwork.asset.Params) |  |  |
| `port_id` | [string](#string) |  | this line is used by starport scaffolding # genesis/proto/state |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="asset/packet.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## asset/packet.proto



<a name="realiotech.realionetwork.asset.AssetPacketData"></a>

### AssetPacketData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `noData` | [NoData](#realiotech.realionetwork.asset.NoData) |  | this line is used by starport scaffolding # ibc/packet/proto/field |






<a name="realiotech.realionetwork.asset.NoData"></a>

### NoData






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="asset/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## asset/query.proto



<a name="realiotech.realionetwork.asset.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is request type for the Query/Params RPC method.






<a name="realiotech.realionetwork.asset.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#realiotech.realionetwork.asset.Params) |  | params holds all the parameters of this module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="realiotech.realionetwork.asset.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#realiotech.realionetwork.asset.QueryParamsRequest) | [QueryParamsResponse](#realiotech.realionetwork.asset.QueryParamsResponse) | Parameters queries the parameters of the module.

this line is used by starport scaffolding # 2 | GET|/realiotech/realionetwork/asset/params|

 <!-- end services -->



<a name="asset/token.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## asset/token.proto



<a name="realiotech.realionetwork.asset.Token"></a>

### Token



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `total` | [int64](#int64) |  |  |
| `decimals` | [int64](#int64) |  |  |
| `authorizationRequired` | [bool](#bool) |  |  |
| `creator` | [string](#string) |  |  |
| `authorized` | [Token.AuthorizedEntry](#realiotech.realionetwork.asset.Token.AuthorizedEntry) | repeated |  |
| `created` | [int64](#int64) |  |  |






<a name="realiotech.realionetwork.asset.Token.AuthorizedEntry"></a>

### Token.AuthorizedEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [TokenAuthorization](#realiotech.realionetwork.asset.TokenAuthorization) |  |  |






<a name="realiotech.realionetwork.asset.TokenAuthorization"></a>

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



<a name="asset/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## asset/tx.proto



<a name="realiotech.realionetwork.asset.MsgAuthorizeAddress"></a>

### MsgAuthorizeAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `address` | [string](#string) |  |  |






<a name="realiotech.realionetwork.asset.MsgAuthorizeAddressResponse"></a>

### MsgAuthorizeAddressResponse







<a name="realiotech.realionetwork.asset.MsgCreateToken"></a>

### MsgCreateToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `total` | [int64](#int64) |  |  |
| `decimals` | [int64](#int64) |  |  |
| `authorizationRequired` | [bool](#bool) |  |  |






<a name="realiotech.realionetwork.asset.MsgCreateTokenResponse"></a>

### MsgCreateTokenResponse







<a name="realiotech.realionetwork.asset.MsgTransferToken"></a>

### MsgTransferToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `from` | [string](#string) |  |  |
| `to` | [string](#string) |  |  |
| `amount` | [int64](#int64) |  |  |






<a name="realiotech.realionetwork.asset.MsgTransferTokenResponse"></a>

### MsgTransferTokenResponse







<a name="realiotech.realionetwork.asset.MsgUnAuthorizeAddress"></a>

### MsgUnAuthorizeAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `address` | [string](#string) |  |  |






<a name="realiotech.realionetwork.asset.MsgUnAuthorizeAddressResponse"></a>

### MsgUnAuthorizeAddressResponse







<a name="realiotech.realionetwork.asset.MsgUpdateToken"></a>

### MsgUpdateToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `authorizationRequired` | [bool](#bool) |  |  |






<a name="realiotech.realionetwork.asset.MsgUpdateTokenResponse"></a>

### MsgUpdateTokenResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="realiotech.realionetwork.asset.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateToken` | [MsgCreateToken](#realiotech.realionetwork.asset.MsgCreateToken) | [MsgCreateTokenResponse](#realiotech.realionetwork.asset.MsgCreateTokenResponse) |  | |
| `UpdateToken` | [MsgUpdateToken](#realiotech.realionetwork.asset.MsgUpdateToken) | [MsgUpdateTokenResponse](#realiotech.realionetwork.asset.MsgUpdateTokenResponse) |  | |
| `AuthorizeAddress` | [MsgAuthorizeAddress](#realiotech.realionetwork.asset.MsgAuthorizeAddress) | [MsgAuthorizeAddressResponse](#realiotech.realionetwork.asset.MsgAuthorizeAddressResponse) |  | |
| `UnAuthorizeAddress` | [MsgUnAuthorizeAddress](#realiotech.realionetwork.asset.MsgUnAuthorizeAddress) | [MsgUnAuthorizeAddressResponse](#realiotech.realionetwork.asset.MsgUnAuthorizeAddressResponse) |  | |
| `TransferToken` | [MsgTransferToken](#realiotech.realionetwork.asset.MsgTransferToken) | [MsgTransferTokenResponse](#realiotech.realionetwork.asset.MsgTransferTokenResponse) | this line is used by starport scaffolding # proto/tx/rpc | |

 <!-- end services -->



<a name="mint/v1beta1/mint.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mint/v1beta1/mint.proto



<a name="realiotech.realionetwork.mint.v1beta1.Minter"></a>

### Minter
Minter represents the minting state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `inflation` | [string](#string) |  | current annual inflation rate |
| `annual_provisions` | [string](#string) |  | current annual expected provisions |






<a name="realiotech.realionetwork.mint.v1beta1.Params"></a>

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



<a name="mint/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mint/v1beta1/genesis.proto



<a name="realiotech.realionetwork.mint.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the mint module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minter` | [Minter](#realiotech.realionetwork.mint.v1beta1.Minter) |  | minter is a space for holding current inflation information. |
| `params` | [Params](#realiotech.realionetwork.mint.v1beta1.Params) |  | params defines all the paramaters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="mint/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mint/v1beta1/query.proto



<a name="realiotech.realionetwork.mint.v1beta1.QueryAnnualProvisionsRequest"></a>

### QueryAnnualProvisionsRequest
QueryAnnualProvisionsRequest is the request type for the
Query/AnnualProvisions RPC method.






<a name="realiotech.realionetwork.mint.v1beta1.QueryAnnualProvisionsResponse"></a>

### QueryAnnualProvisionsResponse
QueryAnnualProvisionsResponse is the response type for the
Query/AnnualProvisions RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `annual_provisions` | [bytes](#bytes) |  | annual_provisions is the current minting annual provisions value. |






<a name="realiotech.realionetwork.mint.v1beta1.QueryInflationRequest"></a>

### QueryInflationRequest
QueryInflationRequest is the request type for the Query/Inflation RPC method.






<a name="realiotech.realionetwork.mint.v1beta1.QueryInflationResponse"></a>

### QueryInflationResponse
QueryInflationResponse is the response type for the Query/Inflation RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `inflation` | [bytes](#bytes) |  | inflation is the current minting inflation value. |






<a name="realiotech.realionetwork.mint.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="realiotech.realionetwork.mint.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#realiotech.realionetwork.mint.v1beta1.Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="realiotech.realionetwork.mint.v1beta1.Query"></a>

### Query
Query provides defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#realiotech.realionetwork.mint.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#realiotech.realionetwork.mint.v1beta1.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/realiotech/realionetwork/v1beta1/params|
| `Inflation` | [QueryInflationRequest](#realiotech.realionetwork.mint.v1beta1.QueryInflationRequest) | [QueryInflationResponse](#realiotech.realionetwork.mint.v1beta1.QueryInflationResponse) | Inflation returns the current minting inflation value. | GET|/realiotech/realionetwork/mint/v1beta1/inflation|
| `AnnualProvisions` | [QueryAnnualProvisionsRequest](#realiotech.realionetwork.mint.v1beta1.QueryAnnualProvisionsRequest) | [QueryAnnualProvisionsResponse](#realiotech.realionetwork.mint.v1beta1.QueryAnnualProvisionsResponse) | AnnualProvisions current minting annual provisions value. | GET|/realiotech/realionetwork/mint/v1beta1/annual_provisions|

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
