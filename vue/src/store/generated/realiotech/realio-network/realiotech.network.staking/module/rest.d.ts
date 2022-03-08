/**
* Validator defines a validator, together with the total amount of the
Validator's bond shares and their exchange rate to coins. Slashing results in
a decrease in the exchange rate, allowing correct calculation of future
undelegations without iterating over delegators. When coins are delegated to
this validator, the validator is credited with a delegation whose number of
bond shares is based on the amount of coins delegated divided by the current
exchange rate. Voting power can be calculated as total bonded shares
multiplied by exchange rate.
*/
export interface NetworkstakingValidator {
    /** operator_address defines the address of the validator's operator; bech encoded in JSON. */
    operatorAddress?: string;
    /** consensus_pubkey is the consensus public key of the validator, as a Protobuf Any. */
    consensusPubkey?: ProtobufAny;
    /** jailed defined whether the validator has been jailed from bonded status or not. */
    jailed?: boolean;
    /** status is the validator status (bonded/unbonding/unbonded). */
    status?: StakingBondStatus;
    /** tokens define the delegated tokens (incl. self-delegation). */
    tokens?: string;
    /** delegator_shares defines total shares issued to a validator's delegators. */
    delegatorShares?: string;
    /** description defines the description terms for the validator. */
    description?: StakingDescription;
    /**
     * unbonding_height defines, if unbonding, the height at which this validator has begun unbonding.
     * @format int64
     */
    unbondingHeight?: string;
    /**
     * unbonding_time defines, if unbonding, the min time for the validator to complete unbonding.
     * @format date-time
     */
    unbondingTime?: string;
    /** commission defines the commission parameters. */
    commission?: StakingCommission;
    /** min_self_delegation is the validator's self declared minimum self delegation. */
    minSelfDelegation?: string;
}
/**
* `Any` contains an arbitrary serialized protocol buffer message along with a
URL that describes the type of the serialized message.

Protobuf library provides support to pack/unpack Any values in the form
of utility functions or additional generated methods of the Any type.

Example 1: Pack and unpack a message in C++.

    Foo foo = ...;
    Any any;
    any.PackFrom(foo);
    ...
    if (any.UnpackTo(&foo)) {
      ...
    }

Example 2: Pack and unpack a message in Java.

    Foo foo = ...;
    Any any = Any.pack(foo);
    ...
    if (any.is(Foo.class)) {
      foo = any.unpack(Foo.class);
    }

 Example 3: Pack and unpack a message in Python.

    foo = Foo(...)
    any = Any()
    any.Pack(foo)
    ...
    if any.Is(Foo.DESCRIPTOR):
      any.Unpack(foo)
      ...

 Example 4: Pack and unpack a message in Go

     foo := &pb.Foo{...}
     any, err := anypb.New(foo)
     if err != nil {
       ...
     }
     ...
     foo := &pb.Foo{}
     if err := any.UnmarshalTo(foo); err != nil {
       ...
     }

The pack methods provided by protobuf library will by default use
'type.googleapis.com/full.type.name' as the type URL and the unpack
methods only use the fully qualified type name after the last '/'
in the type URL, for example "foo.bar.com/x/y.z" will yield type
name "y.z".


JSON
====
The JSON representation of an `Any` value uses the regular
representation of the deserialized, embedded message, with an
additional field `@type` which contains the type URL. Example:

    package google.profile;
    message Person {
      string first_name = 1;
      string last_name = 2;
    }

    {
      "@type": "type.googleapis.com/google.profile.Person",
      "firstName": <string>,
      "lastName": <string>
    }

If the embedded message type is well-known and has a custom JSON
representation, that representation will be embedded adding a field
`value` which holds the custom JSON in addition to the `@type`
field. Example (for message [google.protobuf.Duration][]):

    {
      "@type": "type.googleapis.com/google.protobuf.Duration",
      "value": "1.212s"
    }
*/
export interface ProtobufAny {
    /**
     * A URL/resource name that uniquely identifies the type of the serialized
     * protocol buffer message. This string must contain at least
     * one "/" character. The last segment of the URL's path must represent
     * the fully qualified name of the type (as in
     * `path/google.protobuf.Duration`). The name should be in a canonical form
     * (e.g., leading "." is not accepted).
     *
     * In practice, teams usually precompile into the binary all types that they
     * expect it to use in the context of Any. However, for URLs which use the
     * scheme `http`, `https`, or no scheme, one can optionally set up a type
     * server that maps type URLs to message definitions as follows:
     *
     * * If no scheme is provided, `https` is assumed.
     * * An HTTP GET on the URL must yield a [google.protobuf.Type][]
     *   value in binary format, or produce an error.
     * * Applications are allowed to cache lookup results based on the
     *   URL, or have them precompiled into a binary to avoid any
     *   lookup. Therefore, binary compatibility needs to be preserved
     *   on changes to types. (Use versioned type names to manage
     *   breaking changes.)
     *
     * Note: this functionality is not currently available in the official
     * protobuf release, and it is not used for type URLs beginning with
     * type.googleapis.com.
     *
     * Schemes other than `http`, `https` (or the empty scheme) might be
     * used with implementation specific semantics.
     */
    "@type"?: string;
}
export interface RpcStatus {
    /** @format int32 */
    code?: number;
    message?: string;
    details?: ProtobufAny[];
}
/**
* BondStatus is the status of a validator.

 - BOND_STATUS_UNSPECIFIED: UNSPECIFIED defines an invalid validator status.
 - BOND_STATUS_UNBONDED: UNBONDED defines a validator that is not bonded.
 - BOND_STATUS_UNBONDING: UNBONDING defines a validator that is unbonding.
 - BOND_STATUS_BONDED: BONDED defines a validator that is bonded.
*/
export declare enum StakingBondStatus {
    BOND_STATUS_UNSPECIFIED = "BOND_STATUS_UNSPECIFIED",
    BOND_STATUS_UNBONDED = "BOND_STATUS_UNBONDED",
    BOND_STATUS_UNBONDING = "BOND_STATUS_UNBONDING",
    BOND_STATUS_BONDED = "BOND_STATUS_BONDED"
}
/**
 * Commission defines commission parameters for a given validator.
 */
export interface StakingCommission {
    /** commission_rates defines the initial commission rates to be used for creating a validator. */
    commissionRates?: StakingCommissionRates;
    /**
     * update_time is the last time the commission rate was changed.
     * @format date-time
     */
    updateTime?: string;
}
/**
* CommissionRates defines the initial commission rates to be used for creating
a validator.
*/
export interface StakingCommissionRates {
    /** rate is the commission rate charged to delegators, as a fraction. */
    rate?: string;
    /** max_rate defines the maximum commission rate which validator can ever charge, as a fraction. */
    maxRate?: string;
    /** max_change_rate defines the maximum daily increase of the validator commission, as a fraction. */
    maxChangeRate?: string;
}
/**
* Delegation represents the bond with tokens held by an account. It is
owned by one delegator, and is associated with the voting power of one
validator.
*/
export interface StakingDelegation {
    /** delegator_address is the bech32-encoded address of the delegator. */
    delegatorAddress?: string;
    /** validator_address is the bech32-encoded address of the validator. */
    validatorAddress?: string;
    /** shares define the delegation shares received. */
    shares?: string;
}
/**
* DelegationResponse is equivalent to Delegation except that it contains a
balance in addition to shares which is more suitable for client responses.
*/
export interface StakingDelegationResponse {
    /**
     * Delegation represents the bond with tokens held by an account. It is
     * owned by one delegator, and is associated with the voting power of one
     * validator.
     */
    delegation?: StakingDelegation;
    /**
     * Coin defines a token with a denomination and an amount.
     *
     * NOTE: The amount field is an Int which implements the custom method
     * signatures required by gogoproto.
     */
    balance?: V1Beta1Coin;
}
/**
 * Description defines a validator description.
 */
export interface StakingDescription {
    /** moniker defines a human-readable name for the validator. */
    moniker?: string;
    /** identity defines an optional identity signature (ex. UPort or Keybase). */
    identity?: string;
    /** website defines an optional website link. */
    website?: string;
    /** security_contact defines an optional email for security contact. */
    securityContact?: string;
    /** details define other optional details. */
    details?: string;
}
/**
* HistoricalInfo contains header and validator information for a given block.
It is stored as part of staking module's state, which persists the `n` most
recent HistoricalInfo
(`n` is set by the staking module's `historical_entries` parameter).
*/
export interface StakingHistoricalInfo {
    /** Header defines the structure of a Tendermint block header. */
    header?: TypesHeader;
    valset?: NetworkstakingValidator[];
}
/**
 * MsgBeginRedelegateResponse defines the Msg/BeginRedelegate response type.
 */
export interface StakingMsgBeginRedelegateResponse {
    /** @format date-time */
    completionTime?: string;
}
/**
 * MsgCreateValidatorResponse defines the Msg/CreateValidator response type.
 */
export declare type StakingMsgCreateValidatorResponse = object;
/**
 * MsgDelegateResponse defines the Msg/Delegate response type.
 */
export declare type StakingMsgDelegateResponse = object;
/**
 * MsgEditValidatorResponse defines the Msg/EditValidator response type.
 */
export declare type StakingMsgEditValidatorResponse = object;
/**
 * MsgUndelegateResponse defines the Msg/Undelegate response type.
 */
export interface StakingMsgUndelegateResponse {
    /** @format date-time */
    completionTime?: string;
}
/**
 * Params defines the parameters for the staking module.
 */
export interface StakingParams {
    /** unbonding_time is the time duration of unbonding. */
    unbondingTime?: string;
    /**
     * max_validators is the maximum number of validators.
     * @format int64
     */
    maxValidators?: number;
    /**
     * max_entries is the max entries for either unbonding delegation or redelegation (per pair/trio).
     * @format int64
     */
    maxEntries?: number;
    /**
     * historical_entries is the number of historical entries to persist.
     * @format int64
     */
    historicalEntries?: number;
    /** bond_denom defines the bondable coin denomination. */
    bondDenom?: string;
}
/**
* Pool is used for tracking bonded and not-bonded token supply of the bond
denomination.
*/
export interface StakingPool {
    notBondedTokens?: string;
    bondedTokens?: string;
}
/**
 * QueryDelegationResponse is response type for the Query/Delegation RPC method.
 */
export interface StakingQueryDelegationResponse {
    /** delegation_responses defines the delegation info of a delegation. */
    delegationResponse?: StakingDelegationResponse;
}
/**
* QueryDelegatorDelegationsResponse is response type for the
Query/DelegatorDelegations RPC method.
*/
export interface StakingQueryDelegatorDelegationsResponse {
    /** delegation_responses defines all the delegations' info of a delegator. */
    delegationResponses?: StakingDelegationResponse[];
    /** pagination defines the pagination in the response. */
    pagination?: V1Beta1PageResponse;
}
/**
* QueryUnbondingDelegatorDelegationsResponse is response type for the
Query/UnbondingDelegatorDelegations RPC method.
*/
export interface StakingQueryDelegatorUnbondingDelegationsResponse {
    unbondingResponses?: StakingUnbondingDelegation[];
    /** pagination defines the pagination in the response. */
    pagination?: V1Beta1PageResponse;
}
/**
* QueryDelegatorValidatorResponse response type for the
Query/DelegatorValidator RPC method.
*/
export interface StakingQueryDelegatorValidatorResponse {
    /** validator defines the the validator info. */
    validator?: NetworkstakingValidator;
}
/**
* QueryDelegatorValidatorsResponse is response type for the
Query/DelegatorValidators RPC method.
*/
export interface StakingQueryDelegatorValidatorsResponse {
    /** validators defines the the validators' info of a delegator. */
    validators?: NetworkstakingValidator[];
    /** pagination defines the pagination in the response. */
    pagination?: V1Beta1PageResponse;
}
/**
* QueryHistoricalInfoResponse is response type for the Query/HistoricalInfo RPC
method.
*/
export interface StakingQueryHistoricalInfoResponse {
    /** hist defines the historical info at the given height. */
    hist?: StakingHistoricalInfo;
}
/**
 * QueryParamsResponse is response type for the Query/Params RPC method.
 */
export interface StakingQueryParamsResponse {
    /** params holds all the parameters of this module. */
    params?: StakingParams;
}
/**
 * QueryPoolResponse is response type for the Query/Pool RPC method.
 */
export interface StakingQueryPoolResponse {
    /** pool defines the pool info. */
    pool?: StakingPool;
}
/**
* QueryRedelegationsResponse is response type for the Query/Redelegations RPC
method.
*/
export interface StakingQueryRedelegationsResponse {
    redelegationResponses?: StakingRedelegationResponse[];
    /** pagination defines the pagination in the response. */
    pagination?: V1Beta1PageResponse;
}
/**
* QueryDelegationResponse is response type for the Query/UnbondingDelegation
RPC method.
*/
export interface StakingQueryUnbondingDelegationResponse {
    /** unbond defines the unbonding information of a delegation. */
    unbond?: StakingUnbondingDelegation;
}
export interface StakingQueryValidatorDelegationsResponse {
    delegationResponses?: StakingDelegationResponse[];
    /** pagination defines the pagination in the response. */
    pagination?: V1Beta1PageResponse;
}
export interface StakingQueryValidatorResponse {
    /** validator defines the the validator info. */
    validator?: NetworkstakingValidator;
}
/**
* QueryValidatorUnbondingDelegationsResponse is response type for the
Query/ValidatorUnbondingDelegations RPC method.
*/
export interface StakingQueryValidatorUnbondingDelegationsResponse {
    unbondingResponses?: StakingUnbondingDelegation[];
    /** pagination defines the pagination in the response. */
    pagination?: V1Beta1PageResponse;
}
export interface StakingQueryValidatorsResponse {
    /** validators contains all the queried validators. */
    validators?: NetworkstakingValidator[];
    /** pagination defines the pagination in the response. */
    pagination?: V1Beta1PageResponse;
}
/**
* Redelegation contains the list of a particular delegator's redelegating bonds
from a particular source validator to a particular destination validator.
*/
export interface StakingRedelegation {
    /** delegator_address is the bech32-encoded address of the delegator. */
    delegatorAddress?: string;
    /** validator_src_address is the validator redelegation source operator address. */
    validatorSrcAddress?: string;
    /** validator_dst_address is the validator redelegation destination operator address. */
    validatorDstAddress?: string;
    /** entries are the redelegation entries. */
    entries?: StakingRedelegationEntry[];
}
/**
 * RedelegationEntry defines a redelegation object with relevant metadata.
 */
export interface StakingRedelegationEntry {
    /**
     * creation_height  defines the height which the redelegation took place.
     * @format int64
     */
    creationHeight?: string;
    /**
     * completion_time defines the unix time for redelegation completion.
     * @format date-time
     */
    completionTime?: string;
    /** initial_balance defines the initial balance when redelegation started. */
    initialBalance?: string;
    /** shares_dst is the amount of destination-validator shares created by redelegation. */
    sharesDst?: string;
}
/**
* RedelegationEntryResponse is equivalent to a RedelegationEntry except that it
contains a balance in addition to shares which is more suitable for client
responses.
*/
export interface StakingRedelegationEntryResponse {
    /** RedelegationEntry defines a redelegation object with relevant metadata. */
    redelegationEntry?: StakingRedelegationEntry;
    balance?: string;
}
/**
* RedelegationResponse is equivalent to a Redelegation except that its entries
contain a balance in addition to shares which is more suitable for client
responses.
*/
export interface StakingRedelegationResponse {
    /**
     * Redelegation contains the list of a particular delegator's redelegating bonds
     * from a particular source validator to a particular destination validator.
     */
    redelegation?: StakingRedelegation;
    entries?: StakingRedelegationEntryResponse[];
}
/**
* UnbondingDelegation stores all of a single delegator's unbonding bonds
for a single validator in an time-ordered list.
*/
export interface StakingUnbondingDelegation {
    /** delegator_address is the bech32-encoded address of the delegator. */
    delegatorAddress?: string;
    /** validator_address is the bech32-encoded address of the validator. */
    validatorAddress?: string;
    /** entries are the unbonding delegation entries. */
    entries?: StakingUnbondingDelegationEntry[];
}
/**
 * UnbondingDelegationEntry defines an unbonding object with relevant metadata.
 */
export interface StakingUnbondingDelegationEntry {
    /**
     * creation_height is the height which the unbonding took place.
     * @format int64
     */
    creationHeight?: string;
    /**
     * completion_time is the unix time for unbonding completion.
     * @format date-time
     */
    completionTime?: string;
    /** initial_balance defines the tokens initially scheduled to receive at completion. */
    initialBalance?: string;
    /** balance defines the tokens to receive at completion. */
    balance?: string;
}
export interface TypesBlockID {
    /** @format byte */
    hash?: string;
    partSetHeader?: TypesPartSetHeader;
}
/**
 * Header defines the structure of a Tendermint block header.
 */
export interface TypesHeader {
    /**
     * Consensus captures the consensus rules for processing a block in the blockchain,
     * including all blockchain data structures and the rules of the application's
     * state transition machine.
     */
    version?: VersionConsensus;
    chainId?: string;
    /** @format int64 */
    height?: string;
    /** @format date-time */
    time?: string;
    lastBlockId?: TypesBlockID;
    /** @format byte */
    lastCommitHash?: string;
    /** @format byte */
    dataHash?: string;
    /** @format byte */
    validatorsHash?: string;
    /** @format byte */
    nextValidatorsHash?: string;
    /** @format byte */
    consensusHash?: string;
    /** @format byte */
    appHash?: string;
    /** @format byte */
    lastResultsHash?: string;
    /** @format byte */
    evidenceHash?: string;
    /** @format byte */
    proposerAddress?: string;
}
export interface TypesPartSetHeader {
    /** @format int64 */
    total?: number;
    /** @format byte */
    hash?: string;
}
/**
* Coin defines a token with a denomination and an amount.

NOTE: The amount field is an Int which implements the custom method
signatures required by gogoproto.
*/
export interface V1Beta1Coin {
    denom?: string;
    amount?: string;
}
/**
* message SomeRequest {
         Foo some_parameter = 1;
         PageRequest pagination = 2;
 }
*/
export interface V1Beta1PageRequest {
    /**
     * key is a value returned in PageResponse.next_key to begin
     * querying the next page most efficiently. Only one of offset or key
     * should be set.
     * @format byte
     */
    key?: string;
    /**
     * offset is a numeric offset that can be used when key is unavailable.
     * It is less efficient than using key. Only one of offset or key should
     * be set.
     * @format uint64
     */
    offset?: string;
    /**
     * limit is the total number of results to be returned in the result page.
     * If left empty it will default to a value to be set by each app.
     * @format uint64
     */
    limit?: string;
    /**
     * count_total is set to true  to indicate that the result set should include
     * a count of the total number of items available for pagination in UIs.
     * count_total is only respected when offset is used. It is ignored when key
     * is set.
     */
    countTotal?: boolean;
    /** reverse is set to true if results are to be returned in the descending order. */
    reverse?: boolean;
}
/**
* PageResponse is to be embedded in gRPC response messages where the
corresponding request message has used PageRequest.

 message SomeResponse {
         repeated Bar results = 1;
         PageResponse page = 2;
 }
*/
export interface V1Beta1PageResponse {
    /** @format byte */
    nextKey?: string;
    /** @format uint64 */
    total?: string;
}
/**
* Consensus captures the consensus rules for processing a block in the blockchain,
including all blockchain data structures and the rules of the application's
state transition machine.
*/
export interface VersionConsensus {
    /** @format uint64 */
    block?: string;
    /** @format uint64 */
    app?: string;
}
export declare type QueryParamsType = Record<string | number, any>;
export declare type ResponseFormat = keyof Omit<Body, "body" | "bodyUsed">;
export interface FullRequestParams extends Omit<RequestInit, "body"> {
    /** set parameter to `true` for call `securityWorker` for this request */
    secure?: boolean;
    /** request path */
    path: string;
    /** content type of request body */
    type?: ContentType;
    /** query params */
    query?: QueryParamsType;
    /** format of response (i.e. response.json() -> format: "json") */
    format?: keyof Omit<Body, "body" | "bodyUsed">;
    /** request body */
    body?: unknown;
    /** base url */
    baseUrl?: string;
    /** request cancellation token */
    cancelToken?: CancelToken;
}
export declare type RequestParams = Omit<FullRequestParams, "body" | "method" | "query" | "path">;
export interface ApiConfig<SecurityDataType = unknown> {
    baseUrl?: string;
    baseApiParams?: Omit<RequestParams, "baseUrl" | "cancelToken" | "signal">;
    securityWorker?: (securityData: SecurityDataType) => RequestParams | void;
}
export interface HttpResponse<D extends unknown, E extends unknown = unknown> extends Response {
    data: D;
    error: E;
}
declare type CancelToken = Symbol | string | number;
export declare enum ContentType {
    Json = "application/json",
    FormData = "multipart/form-data",
    UrlEncoded = "application/x-www-form-urlencoded"
}
export declare class HttpClient<SecurityDataType = unknown> {
    baseUrl: string;
    private securityData;
    private securityWorker;
    private abortControllers;
    private baseApiParams;
    constructor(apiConfig?: ApiConfig<SecurityDataType>);
    setSecurityData: (data: SecurityDataType) => void;
    private addQueryParam;
    protected toQueryString(rawQuery?: QueryParamsType): string;
    protected addQueryParams(rawQuery?: QueryParamsType): string;
    private contentFormatters;
    private mergeRequestParams;
    private createAbortSignal;
    abortRequest: (cancelToken: CancelToken) => void;
    request: <T = any, E = any>({ body, secure, path, type, query, format, baseUrl, cancelToken, ...params }: FullRequestParams) => Promise<HttpResponse<T, E>>;
}
/**
 * @title staking/authz.proto
 * @version version not set
 */
export declare class Api<SecurityDataType extends unknown> extends HttpClient<SecurityDataType> {
    /**
     * No description
     *
     * @tags Query
     * @name QueryDelegatorDelegations
     * @summary DelegatorDelegations queries all delegations of a given delegator address.
     * @request GET:/realiotech/network/staking/delegations/{delegatorAddr}
     */
    queryDelegatorDelegations: (delegatorAddr: string, query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
        "pagination.reverse"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<StakingQueryDelegatorDelegationsResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryRedelegations
     * @summary Redelegations queries redelegations of given address.
     * @request GET:/realiotech/network/staking/delegators/{delegatorAddr}/redelegations
     */
    queryRedelegations: (delegatorAddr: string, query?: {
        srcValidatorAddr?: string;
        dstValidatorAddr?: string;
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
        "pagination.reverse"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<StakingQueryRedelegationsResponse, RpcStatus>>;
    /**
   * No description
   *
   * @tags Query
   * @name QueryDelegatorUnbondingDelegations
   * @summary DelegatorUnbondingDelegations queries all unbonding delegations of a given
  delegator address.
   * @request GET:/realiotech/network/staking/delegators/{delegatorAddr}/unbonding_delegations
   */
    queryDelegatorUnbondingDelegations: (delegatorAddr: string, query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
        "pagination.reverse"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<StakingQueryDelegatorUnbondingDelegationsResponse, RpcStatus>>;
    /**
   * No description
   *
   * @tags Query
   * @name QueryDelegatorValidators
   * @summary DelegatorValidators queries all validators info for given delegator
  address.
   * @request GET:/realiotech/network/staking/delegators/{delegatorAddr}/validators
   */
    queryDelegatorValidators: (delegatorAddr: string, query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
        "pagination.reverse"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<StakingQueryDelegatorValidatorsResponse, RpcStatus>>;
    /**
   * No description
   *
   * @tags Query
   * @name QueryDelegatorValidator
   * @summary DelegatorValidator queries validator info for given delegator validator
  pair.
   * @request GET:/realiotech/network/staking/delegators/{delegatorAddr}/validators/{validatorAddr}
   */
    queryDelegatorValidator: (delegatorAddr: string, validatorAddr: string, params?: RequestParams) => Promise<HttpResponse<StakingQueryDelegatorValidatorResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryHistoricalInfo
     * @summary HistoricalInfo queries the historical info for given height.
     * @request GET:/realiotech/network/staking/historical_info/{height}
     */
    queryHistoricalInfo: (height: string, params?: RequestParams) => Promise<HttpResponse<StakingQueryHistoricalInfoResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryParams
     * @summary Parameters queries the staking parameters.
     * @request GET:/realiotech/network/staking/params
     */
    queryParams: (params?: RequestParams) => Promise<HttpResponse<StakingQueryParamsResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryPool
     * @summary Pool queries the pool info.
     * @request GET:/realiotech/network/staking/pool
     */
    queryPool: (params?: RequestParams) => Promise<HttpResponse<StakingQueryPoolResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryValidators
     * @summary Validators queries all validators that match the given status.
     * @request GET:/realiotech/network/staking/validators
     */
    queryValidators: (query?: {
        status?: string;
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
        "pagination.reverse"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<StakingQueryValidatorsResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryValidator
     * @summary Validator queries validator info for given validator address.
     * @request GET:/realiotech/network/staking/validators/{validatorAddr}
     */
    queryValidator: (validatorAddr: string, params?: RequestParams) => Promise<HttpResponse<StakingQueryValidatorResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryValidatorDelegations
     * @summary ValidatorDelegations queries delegate info for given validator.
     * @request GET:/realiotech/network/staking/validators/{validatorAddr}/delegations
     */
    queryValidatorDelegations: (validatorAddr: string, query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
        "pagination.reverse"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<StakingQueryValidatorDelegationsResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryDelegation
     * @summary Delegation queries delegate info for given validator delegator pair.
     * @request GET:/realiotech/network/staking/validators/{validatorAddr}/delegations/{delegatorAddr}
     */
    queryDelegation: (validatorAddr: string, delegatorAddr: string, params?: RequestParams) => Promise<HttpResponse<StakingQueryDelegationResponse, RpcStatus>>;
    /**
   * No description
   *
   * @tags Query
   * @name QueryUnbondingDelegation
   * @summary UnbondingDelegation queries unbonding info for given validator delegator
  pair.
   * @request GET:/realiotech/network/staking/validators/{validatorAddr}/delegations/{delegatorAddr}/unbonding_delegation
   */
    queryUnbondingDelegation: (validatorAddr: string, delegatorAddr: string, params?: RequestParams) => Promise<HttpResponse<StakingQueryUnbondingDelegationResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryValidatorUnbondingDelegations
     * @summary ValidatorUnbondingDelegations queries unbonding delegations of a validator.
     * @request GET:/realiotech/network/staking/validators/{validatorAddr}/unbonding_delegations
     */
    queryValidatorUnbondingDelegations: (validatorAddr: string, query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
        "pagination.reverse"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<StakingQueryValidatorUnbondingDelegationsResponse, RpcStatus>>;
}
export {};
