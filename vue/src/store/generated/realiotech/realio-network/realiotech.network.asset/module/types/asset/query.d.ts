import { Reader, Writer } from 'protobufjs/minimal';
import { Token } from '../asset/token';
import { PageRequest, PageResponse } from '../cosmos/base/query/v1beta1/pagination';
export declare const protobufPackage = "realiotech.network.asset";
export interface QueryGetTokenRequest {
    index: string;
}
export interface QueryGetTokenResponse {
    token: Token | undefined;
}
export interface QueryAllTokenRequest {
    pagination: PageRequest | undefined;
}
export interface QueryAllTokenResponse {
    token: Token[];
    pagination: PageResponse | undefined;
}
export declare const QueryGetTokenRequest: {
    encode(message: QueryGetTokenRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryGetTokenRequest;
    fromJSON(object: any): QueryGetTokenRequest;
    toJSON(message: QueryGetTokenRequest): unknown;
    fromPartial(object: DeepPartial<QueryGetTokenRequest>): QueryGetTokenRequest;
};
export declare const QueryGetTokenResponse: {
    encode(message: QueryGetTokenResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryGetTokenResponse;
    fromJSON(object: any): QueryGetTokenResponse;
    toJSON(message: QueryGetTokenResponse): unknown;
    fromPartial(object: DeepPartial<QueryGetTokenResponse>): QueryGetTokenResponse;
};
export declare const QueryAllTokenRequest: {
    encode(message: QueryAllTokenRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryAllTokenRequest;
    fromJSON(object: any): QueryAllTokenRequest;
    toJSON(message: QueryAllTokenRequest): unknown;
    fromPartial(object: DeepPartial<QueryAllTokenRequest>): QueryAllTokenRequest;
};
export declare const QueryAllTokenResponse: {
    encode(message: QueryAllTokenResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryAllTokenResponse;
    fromJSON(object: any): QueryAllTokenResponse;
    toJSON(message: QueryAllTokenResponse): unknown;
    fromPartial(object: DeepPartial<QueryAllTokenResponse>): QueryAllTokenResponse;
};
/** Query defines the gRPC querier service. */
export interface Query {
    /** Queries a token by index. */
    Token(request: QueryGetTokenRequest): Promise<QueryGetTokenResponse>;
    /** Queries a list of token items. */
    TokenAll(request: QueryAllTokenRequest): Promise<QueryAllTokenResponse>;
}
export declare class QueryClientImpl implements Query {
    private readonly rpc;
    constructor(rpc: Rpc);
    Token(request: QueryGetTokenRequest): Promise<QueryGetTokenResponse>;
    TokenAll(request: QueryAllTokenRequest): Promise<QueryAllTokenResponse>;
}
interface Rpc {
    request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
