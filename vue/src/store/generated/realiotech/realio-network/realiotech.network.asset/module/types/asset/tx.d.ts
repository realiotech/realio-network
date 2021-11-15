import { Reader, Writer } from 'protobufjs/minimal';
export declare const protobufPackage = "realiotech.network.asset";
export interface MsgCreateToken {
    creator: string;
    index: string;
    name: string;
    symbol: string;
    total: number;
    decimals: string;
    authorizationRequired: boolean;
}
export interface MsgCreateTokenResponse {
}
export interface MsgUpdateToken {
    creator: string;
    index: string;
    authorizationRequired: boolean;
}
export interface MsgUpdateTokenResponse {
}
export interface MsgAuthorizeAddress {
    creator: string;
    index: string;
    address: string;
}
export interface MsgAuthorizeAddressResponse {
}
export interface MsgUnAuthorizeAddress {
    creator: string;
    index: string;
    address: string;
}
export interface MsgUnAuthorizeAddressResponse {
}
export interface MsgTransferToken {
    creator: string;
    index: string;
    symbol: string;
    from: string;
    to: string;
    amount: number;
}
export interface MsgTransferTokenResponse {
}
export declare const MsgCreateToken: {
    encode(message: MsgCreateToken, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateToken;
    fromJSON(object: any): MsgCreateToken;
    toJSON(message: MsgCreateToken): unknown;
    fromPartial(object: DeepPartial<MsgCreateToken>): MsgCreateToken;
};
export declare const MsgCreateTokenResponse: {
    encode(_: MsgCreateTokenResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateTokenResponse;
    fromJSON(_: any): MsgCreateTokenResponse;
    toJSON(_: MsgCreateTokenResponse): unknown;
    fromPartial(_: DeepPartial<MsgCreateTokenResponse>): MsgCreateTokenResponse;
};
export declare const MsgUpdateToken: {
    encode(message: MsgUpdateToken, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateToken;
    fromJSON(object: any): MsgUpdateToken;
    toJSON(message: MsgUpdateToken): unknown;
    fromPartial(object: DeepPartial<MsgUpdateToken>): MsgUpdateToken;
};
export declare const MsgUpdateTokenResponse: {
    encode(_: MsgUpdateTokenResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateTokenResponse;
    fromJSON(_: any): MsgUpdateTokenResponse;
    toJSON(_: MsgUpdateTokenResponse): unknown;
    fromPartial(_: DeepPartial<MsgUpdateTokenResponse>): MsgUpdateTokenResponse;
};
export declare const MsgAuthorizeAddress: {
    encode(message: MsgAuthorizeAddress, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgAuthorizeAddress;
    fromJSON(object: any): MsgAuthorizeAddress;
    toJSON(message: MsgAuthorizeAddress): unknown;
    fromPartial(object: DeepPartial<MsgAuthorizeAddress>): MsgAuthorizeAddress;
};
export declare const MsgAuthorizeAddressResponse: {
    encode(_: MsgAuthorizeAddressResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgAuthorizeAddressResponse;
    fromJSON(_: any): MsgAuthorizeAddressResponse;
    toJSON(_: MsgAuthorizeAddressResponse): unknown;
    fromPartial(_: DeepPartial<MsgAuthorizeAddressResponse>): MsgAuthorizeAddressResponse;
};
export declare const MsgUnAuthorizeAddress: {
    encode(message: MsgUnAuthorizeAddress, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUnAuthorizeAddress;
    fromJSON(object: any): MsgUnAuthorizeAddress;
    toJSON(message: MsgUnAuthorizeAddress): unknown;
    fromPartial(object: DeepPartial<MsgUnAuthorizeAddress>): MsgUnAuthorizeAddress;
};
export declare const MsgUnAuthorizeAddressResponse: {
    encode(_: MsgUnAuthorizeAddressResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUnAuthorizeAddressResponse;
    fromJSON(_: any): MsgUnAuthorizeAddressResponse;
    toJSON(_: MsgUnAuthorizeAddressResponse): unknown;
    fromPartial(_: DeepPartial<MsgUnAuthorizeAddressResponse>): MsgUnAuthorizeAddressResponse;
};
export declare const MsgTransferToken: {
    encode(message: MsgTransferToken, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgTransferToken;
    fromJSON(object: any): MsgTransferToken;
    toJSON(message: MsgTransferToken): unknown;
    fromPartial(object: DeepPartial<MsgTransferToken>): MsgTransferToken;
};
export declare const MsgTransferTokenResponse: {
    encode(_: MsgTransferTokenResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgTransferTokenResponse;
    fromJSON(_: any): MsgTransferTokenResponse;
    toJSON(_: MsgTransferTokenResponse): unknown;
    fromPartial(_: DeepPartial<MsgTransferTokenResponse>): MsgTransferTokenResponse;
};
/** Msg defines the Msg service. */
export interface Msg {
    CreateToken(request: MsgCreateToken): Promise<MsgCreateTokenResponse>;
    UpdateToken(request: MsgUpdateToken): Promise<MsgUpdateTokenResponse>;
    AuthorizeAddress(request: MsgAuthorizeAddress): Promise<MsgAuthorizeAddressResponse>;
    UnAuthorizeAddress(request: MsgUnAuthorizeAddress): Promise<MsgUnAuthorizeAddressResponse>;
    /** this line is used by starport scaffolding # proto/tx/rpc */
    TransferToken(request: MsgTransferToken): Promise<MsgTransferTokenResponse>;
}
export declare class MsgClientImpl implements Msg {
    private readonly rpc;
    constructor(rpc: Rpc);
    CreateToken(request: MsgCreateToken): Promise<MsgCreateTokenResponse>;
    UpdateToken(request: MsgUpdateToken): Promise<MsgUpdateTokenResponse>;
    AuthorizeAddress(request: MsgAuthorizeAddress): Promise<MsgAuthorizeAddressResponse>;
    UnAuthorizeAddress(request: MsgUnAuthorizeAddress): Promise<MsgUnAuthorizeAddressResponse>;
    TransferToken(request: MsgTransferToken): Promise<MsgTransferTokenResponse>;
}
interface Rpc {
    request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
