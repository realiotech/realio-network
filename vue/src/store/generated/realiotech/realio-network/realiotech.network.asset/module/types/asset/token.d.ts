import { Writer, Reader } from 'protobufjs/minimal';
export declare const protobufPackage = "realiotech.network.asset";
export interface Token {
    index: string;
    name: string;
    symbol: string;
    total: number;
    decimals: string;
    authorizationRequired: boolean;
    creator: string;
    authorized: {
        [key: string]: TokenAuthorization;
    };
}
export interface Token_AuthorizedEntry {
    key: string;
    value: TokenAuthorization | undefined;
}
export interface TokenAuthorization {
    tokenIndex: string;
    address: string;
    authorized: boolean;
}
export declare const Token: {
    encode(message: Token, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Token;
    fromJSON(object: any): Token;
    toJSON(message: Token): unknown;
    fromPartial(object: DeepPartial<Token>): Token;
};
export declare const Token_AuthorizedEntry: {
    encode(message: Token_AuthorizedEntry, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Token_AuthorizedEntry;
    fromJSON(object: any): Token_AuthorizedEntry;
    toJSON(message: Token_AuthorizedEntry): unknown;
    fromPartial(object: DeepPartial<Token_AuthorizedEntry>): Token_AuthorizedEntry;
};
export declare const TokenAuthorization: {
    encode(message: TokenAuthorization, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): TokenAuthorization;
    fromJSON(object: any): TokenAuthorization;
    toJSON(message: TokenAuthorization): unknown;
    fromPartial(object: DeepPartial<TokenAuthorization>): TokenAuthorization;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
