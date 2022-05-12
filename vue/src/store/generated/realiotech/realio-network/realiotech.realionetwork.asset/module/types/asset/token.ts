/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "realiotech.realionetwork.asset";

export interface Token {
  name: string;
  symbol: string;
  total: number;
  decimals: number;
  authorizationRequired: boolean;
  creator: string;
  authorized: { [key: string]: TokenAuthorization };
  created: number;
}

export interface Token_AuthorizedEntry {
  key: string;
  value: TokenAuthorization | undefined;
}

export interface TokenAuthorization {
  tokenSymbol: string;
  address: string;
  authorized: boolean;
}

const baseToken: object = {
  name: "",
  symbol: "",
  total: 0,
  decimals: 0,
  authorizationRequired: false,
  creator: "",
  created: 0,
};

export const Token = {
  encode(message: Token, writer: Writer = Writer.create()): Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.symbol !== "") {
      writer.uint32(18).string(message.symbol);
    }
    if (message.total !== 0) {
      writer.uint32(24).int64(message.total);
    }
    if (message.decimals !== 0) {
      writer.uint32(32).int64(message.decimals);
    }
    if (message.authorizationRequired === true) {
      writer.uint32(40).bool(message.authorizationRequired);
    }
    if (message.creator !== "") {
      writer.uint32(50).string(message.creator);
    }
    Object.entries(message.authorized).forEach(([key, value]) => {
      Token_AuthorizedEntry.encode(
        { key: key as any, value },
        writer.uint32(58).fork()
      ).ldelim();
    });
    if (message.created !== 0) {
      writer.uint32(64).int64(message.created);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Token {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseToken } as Token;
    message.authorized = {};
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.symbol = reader.string();
          break;
        case 3:
          message.total = longToNumber(reader.int64() as Long);
          break;
        case 4:
          message.decimals = longToNumber(reader.int64() as Long);
          break;
        case 5:
          message.authorizationRequired = reader.bool();
          break;
        case 6:
          message.creator = reader.string();
          break;
        case 7:
          const entry7 = Token_AuthorizedEntry.decode(reader, reader.uint32());
          if (entry7.value !== undefined) {
            message.authorized[entry7.key] = entry7.value;
          }
          break;
        case 8:
          message.created = longToNumber(reader.int64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Token {
    const message = { ...baseToken } as Token;
    message.authorized = {};
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = String(object.symbol);
    } else {
      message.symbol = "";
    }
    if (object.total !== undefined && object.total !== null) {
      message.total = Number(object.total);
    } else {
      message.total = 0;
    }
    if (object.decimals !== undefined && object.decimals !== null) {
      message.decimals = Number(object.decimals);
    } else {
      message.decimals = 0;
    }
    if (
      object.authorizationRequired !== undefined &&
      object.authorizationRequired !== null
    ) {
      message.authorizationRequired = Boolean(object.authorizationRequired);
    } else {
      message.authorizationRequired = false;
    }
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.authorized !== undefined && object.authorized !== null) {
      Object.entries(object.authorized).forEach(([key, value]) => {
        message.authorized[key] = TokenAuthorization.fromJSON(value);
      });
    }
    if (object.created !== undefined && object.created !== null) {
      message.created = Number(object.created);
    } else {
      message.created = 0;
    }
    return message;
  },

  toJSON(message: Token): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.symbol !== undefined && (obj.symbol = message.symbol);
    message.total !== undefined && (obj.total = message.total);
    message.decimals !== undefined && (obj.decimals = message.decimals);
    message.authorizationRequired !== undefined &&
      (obj.authorizationRequired = message.authorizationRequired);
    message.creator !== undefined && (obj.creator = message.creator);
    obj.authorized = {};
    if (message.authorized) {
      Object.entries(message.authorized).forEach(([k, v]) => {
        obj.authorized[k] = TokenAuthorization.toJSON(v);
      });
    }
    message.created !== undefined && (obj.created = message.created);
    return obj;
  },

  fromPartial(object: DeepPartial<Token>): Token {
    const message = { ...baseToken } as Token;
    message.authorized = {};
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = object.symbol;
    } else {
      message.symbol = "";
    }
    if (object.total !== undefined && object.total !== null) {
      message.total = object.total;
    } else {
      message.total = 0;
    }
    if (object.decimals !== undefined && object.decimals !== null) {
      message.decimals = object.decimals;
    } else {
      message.decimals = 0;
    }
    if (
      object.authorizationRequired !== undefined &&
      object.authorizationRequired !== null
    ) {
      message.authorizationRequired = object.authorizationRequired;
    } else {
      message.authorizationRequired = false;
    }
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.authorized !== undefined && object.authorized !== null) {
      Object.entries(object.authorized).forEach(([key, value]) => {
        if (value !== undefined) {
          message.authorized[key] = TokenAuthorization.fromPartial(value);
        }
      });
    }
    if (object.created !== undefined && object.created !== null) {
      message.created = object.created;
    } else {
      message.created = 0;
    }
    return message;
  },
};

const baseToken_AuthorizedEntry: object = { key: "" };

export const Token_AuthorizedEntry = {
  encode(
    message: Token_AuthorizedEntry,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value !== undefined) {
      TokenAuthorization.encode(
        message.value,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Token_AuthorizedEntry {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseToken_AuthorizedEntry } as Token_AuthorizedEntry;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.key = reader.string();
          break;
        case 2:
          message.value = TokenAuthorization.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Token_AuthorizedEntry {
    const message = { ...baseToken_AuthorizedEntry } as Token_AuthorizedEntry;
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = TokenAuthorization.fromJSON(object.value);
    } else {
      message.value = undefined;
    }
    return message;
  },

  toJSON(message: Token_AuthorizedEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined &&
      (obj.value = message.value
        ? TokenAuthorization.toJSON(message.value)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<Token_AuthorizedEntry>
  ): Token_AuthorizedEntry {
    const message = { ...baseToken_AuthorizedEntry } as Token_AuthorizedEntry;
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = TokenAuthorization.fromPartial(object.value);
    } else {
      message.value = undefined;
    }
    return message;
  },
};

const baseTokenAuthorization: object = {
  tokenSymbol: "",
  address: "",
  authorized: false,
};

export const TokenAuthorization = {
  encode(
    message: TokenAuthorization,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.tokenSymbol !== "") {
      writer.uint32(10).string(message.tokenSymbol);
    }
    if (message.address !== "") {
      writer.uint32(18).string(message.address);
    }
    if (message.authorized === true) {
      writer.uint32(24).bool(message.authorized);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): TokenAuthorization {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseTokenAuthorization } as TokenAuthorization;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.tokenSymbol = reader.string();
          break;
        case 2:
          message.address = reader.string();
          break;
        case 3:
          message.authorized = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): TokenAuthorization {
    const message = { ...baseTokenAuthorization } as TokenAuthorization;
    if (object.tokenSymbol !== undefined && object.tokenSymbol !== null) {
      message.tokenSymbol = String(object.tokenSymbol);
    } else {
      message.tokenSymbol = "";
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = String(object.address);
    } else {
      message.address = "";
    }
    if (object.authorized !== undefined && object.authorized !== null) {
      message.authorized = Boolean(object.authorized);
    } else {
      message.authorized = false;
    }
    return message;
  },

  toJSON(message: TokenAuthorization): unknown {
    const obj: any = {};
    message.tokenSymbol !== undefined &&
      (obj.tokenSymbol = message.tokenSymbol);
    message.address !== undefined && (obj.address = message.address);
    message.authorized !== undefined && (obj.authorized = message.authorized);
    return obj;
  },

  fromPartial(object: DeepPartial<TokenAuthorization>): TokenAuthorization {
    const message = { ...baseTokenAuthorization } as TokenAuthorization;
    if (object.tokenSymbol !== undefined && object.tokenSymbol !== null) {
      message.tokenSymbol = object.tokenSymbol;
    } else {
      message.tokenSymbol = "";
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = object.address;
    } else {
      message.address = "";
    }
    if (object.authorized !== undefined && object.authorized !== null) {
      message.authorized = object.authorized;
    } else {
      message.authorized = false;
    }
    return message;
  },
};

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
