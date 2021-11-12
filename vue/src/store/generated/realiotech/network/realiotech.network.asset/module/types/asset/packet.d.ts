import { Writer, Reader } from 'protobufjs/minimal';
export declare const protobufPackage = "realiotech.network.asset";
export interface AssetPacketData {
    noData: NoData | undefined;
    /** this line is used by starport scaffolding # ibc/packet/proto/field */
    fungibleTokenTransferPacket: FungibleTokenTransferPacketData | undefined;
}
export interface NoData {
}
/** FungibleTokenTransferPacketData defines a struct for the packet payload */
export interface FungibleTokenTransferPacketData {
    denom: string;
    amount: number;
    receiver: string;
    sender: string;
}
/** FungibleTokenTransferPacketAck defines a struct for the packet acknowledgment */
export interface FungibleTokenTransferPacketAck {
}
export declare const AssetPacketData: {
    encode(message: AssetPacketData, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): AssetPacketData;
    fromJSON(object: any): AssetPacketData;
    toJSON(message: AssetPacketData): unknown;
    fromPartial(object: DeepPartial<AssetPacketData>): AssetPacketData;
};
export declare const NoData: {
    encode(_: NoData, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): NoData;
    fromJSON(_: any): NoData;
    toJSON(_: NoData): unknown;
    fromPartial(_: DeepPartial<NoData>): NoData;
};
export declare const FungibleTokenTransferPacketData: {
    encode(message: FungibleTokenTransferPacketData, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): FungibleTokenTransferPacketData;
    fromJSON(object: any): FungibleTokenTransferPacketData;
    toJSON(message: FungibleTokenTransferPacketData): unknown;
    fromPartial(object: DeepPartial<FungibleTokenTransferPacketData>): FungibleTokenTransferPacketData;
};
export declare const FungibleTokenTransferPacketAck: {
    encode(_: FungibleTokenTransferPacketAck, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): FungibleTokenTransferPacketAck;
    fromJSON(_: any): FungibleTokenTransferPacketAck;
    toJSON(_: FungibleTokenTransferPacketAck): unknown;
    fromPartial(_: DeepPartial<FungibleTokenTransferPacketAck>): FungibleTokenTransferPacketAck;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
