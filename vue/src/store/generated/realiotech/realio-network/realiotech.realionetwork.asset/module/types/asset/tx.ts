/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";

export const protobufPackage = "realiotech.realionetwork.asset";

export interface MsgCreateToken {
  creator: string;
  name: string;
  symbol: string;
  total: number;
  decimals: number;
  authorizationRequired: boolean;
}

export interface MsgCreateTokenResponse {}

export interface MsgUpdateToken {
  creator: string;
  symbol: string;
  authorizationRequired: boolean;
}

export interface MsgUpdateTokenResponse {}

export interface MsgAuthorizeAddress {
  creator: string;
  symbol: string;
  address: string;
}

export interface MsgAuthorizeAddressResponse {}

export interface MsgUnAuthorizeAddress {
  creator: string;
  symbol: string;
  address: string;
}

export interface MsgUnAuthorizeAddressResponse {}

export interface MsgTransferToken {
  creator: string;
  symbol: string;
  from: string;
  to: string;
  amount: number;
}

export interface MsgTransferTokenResponse {}

const baseMsgCreateToken: object = {
  creator: "",
  name: "",
  symbol: "",
  total: 0,
  decimals: 0,
  authorizationRequired: false,
};

export const MsgCreateToken = {
  encode(message: MsgCreateToken, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.name !== "") {
      writer.uint32(18).string(message.name);
    }
    if (message.symbol !== "") {
      writer.uint32(26).string(message.symbol);
    }
    if (message.total !== 0) {
      writer.uint32(32).int64(message.total);
    }
    if (message.decimals !== 0) {
      writer.uint32(40).int64(message.decimals);
    }
    if (message.authorizationRequired === true) {
      writer.uint32(48).bool(message.authorizationRequired);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateToken {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateToken } as MsgCreateToken;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.name = reader.string();
          break;
        case 3:
          message.symbol = reader.string();
          break;
        case 4:
          message.total = longToNumber(reader.int64() as Long);
          break;
        case 5:
          message.decimals = longToNumber(reader.int64() as Long);
          break;
        case 6:
          message.authorizationRequired = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgCreateToken {
    const message = { ...baseMsgCreateToken } as MsgCreateToken;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
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
    return message;
  },

  toJSON(message: MsgCreateToken): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.name !== undefined && (obj.name = message.name);
    message.symbol !== undefined && (obj.symbol = message.symbol);
    message.total !== undefined && (obj.total = message.total);
    message.decimals !== undefined && (obj.decimals = message.decimals);
    message.authorizationRequired !== undefined &&
      (obj.authorizationRequired = message.authorizationRequired);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgCreateToken>): MsgCreateToken {
    const message = { ...baseMsgCreateToken } as MsgCreateToken;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
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
    return message;
  },
};

const baseMsgCreateTokenResponse: object = {};

export const MsgCreateTokenResponse = {
  encode(_: MsgCreateTokenResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateTokenResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateTokenResponse } as MsgCreateTokenResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgCreateTokenResponse {
    const message = { ...baseMsgCreateTokenResponse } as MsgCreateTokenResponse;
    return message;
  },

  toJSON(_: MsgCreateTokenResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgCreateTokenResponse>): MsgCreateTokenResponse {
    const message = { ...baseMsgCreateTokenResponse } as MsgCreateTokenResponse;
    return message;
  },
};

const baseMsgUpdateToken: object = {
  creator: "",
  symbol: "",
  authorizationRequired: false,
};

export const MsgUpdateToken = {
  encode(message: MsgUpdateToken, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.symbol !== "") {
      writer.uint32(18).string(message.symbol);
    }
    if (message.authorizationRequired === true) {
      writer.uint32(24).bool(message.authorizationRequired);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateToken {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateToken } as MsgUpdateToken;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.symbol = reader.string();
          break;
        case 3:
          message.authorizationRequired = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgUpdateToken {
    const message = { ...baseMsgUpdateToken } as MsgUpdateToken;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = String(object.symbol);
    } else {
      message.symbol = "";
    }
    if (
      object.authorizationRequired !== undefined &&
      object.authorizationRequired !== null
    ) {
      message.authorizationRequired = Boolean(object.authorizationRequired);
    } else {
      message.authorizationRequired = false;
    }
    return message;
  },

  toJSON(message: MsgUpdateToken): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.symbol !== undefined && (obj.symbol = message.symbol);
    message.authorizationRequired !== undefined &&
      (obj.authorizationRequired = message.authorizationRequired);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgUpdateToken>): MsgUpdateToken {
    const message = { ...baseMsgUpdateToken } as MsgUpdateToken;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = object.symbol;
    } else {
      message.symbol = "";
    }
    if (
      object.authorizationRequired !== undefined &&
      object.authorizationRequired !== null
    ) {
      message.authorizationRequired = object.authorizationRequired;
    } else {
      message.authorizationRequired = false;
    }
    return message;
  },
};

const baseMsgUpdateTokenResponse: object = {};

export const MsgUpdateTokenResponse = {
  encode(_: MsgUpdateTokenResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateTokenResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateTokenResponse } as MsgUpdateTokenResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgUpdateTokenResponse {
    const message = { ...baseMsgUpdateTokenResponse } as MsgUpdateTokenResponse;
    return message;
  },

  toJSON(_: MsgUpdateTokenResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgUpdateTokenResponse>): MsgUpdateTokenResponse {
    const message = { ...baseMsgUpdateTokenResponse } as MsgUpdateTokenResponse;
    return message;
  },
};

const baseMsgAuthorizeAddress: object = {
  creator: "",
  symbol: "",
  address: "",
};

export const MsgAuthorizeAddress = {
  encode(
    message: MsgAuthorizeAddress,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.symbol !== "") {
      writer.uint32(18).string(message.symbol);
    }
    if (message.address !== "") {
      writer.uint32(26).string(message.address);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgAuthorizeAddress {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgAuthorizeAddress } as MsgAuthorizeAddress;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.symbol = reader.string();
          break;
        case 3:
          message.address = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgAuthorizeAddress {
    const message = { ...baseMsgAuthorizeAddress } as MsgAuthorizeAddress;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = String(object.symbol);
    } else {
      message.symbol = "";
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = String(object.address);
    } else {
      message.address = "";
    }
    return message;
  },

  toJSON(message: MsgAuthorizeAddress): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.symbol !== undefined && (obj.symbol = message.symbol);
    message.address !== undefined && (obj.address = message.address);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgAuthorizeAddress>): MsgAuthorizeAddress {
    const message = { ...baseMsgAuthorizeAddress } as MsgAuthorizeAddress;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = object.symbol;
    } else {
      message.symbol = "";
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = object.address;
    } else {
      message.address = "";
    }
    return message;
  },
};

const baseMsgAuthorizeAddressResponse: object = {};

export const MsgAuthorizeAddressResponse = {
  encode(
    _: MsgAuthorizeAddressResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgAuthorizeAddressResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgAuthorizeAddressResponse,
    } as MsgAuthorizeAddressResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgAuthorizeAddressResponse {
    const message = {
      ...baseMsgAuthorizeAddressResponse,
    } as MsgAuthorizeAddressResponse;
    return message;
  },

  toJSON(_: MsgAuthorizeAddressResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgAuthorizeAddressResponse>
  ): MsgAuthorizeAddressResponse {
    const message = {
      ...baseMsgAuthorizeAddressResponse,
    } as MsgAuthorizeAddressResponse;
    return message;
  },
};

const baseMsgUnAuthorizeAddress: object = {
  creator: "",
  symbol: "",
  address: "",
};

export const MsgUnAuthorizeAddress = {
  encode(
    message: MsgUnAuthorizeAddress,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.symbol !== "") {
      writer.uint32(18).string(message.symbol);
    }
    if (message.address !== "") {
      writer.uint32(26).string(message.address);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUnAuthorizeAddress {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUnAuthorizeAddress } as MsgUnAuthorizeAddress;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.symbol = reader.string();
          break;
        case 3:
          message.address = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgUnAuthorizeAddress {
    const message = { ...baseMsgUnAuthorizeAddress } as MsgUnAuthorizeAddress;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = String(object.symbol);
    } else {
      message.symbol = "";
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = String(object.address);
    } else {
      message.address = "";
    }
    return message;
  },

  toJSON(message: MsgUnAuthorizeAddress): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.symbol !== undefined && (obj.symbol = message.symbol);
    message.address !== undefined && (obj.address = message.address);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgUnAuthorizeAddress>
  ): MsgUnAuthorizeAddress {
    const message = { ...baseMsgUnAuthorizeAddress } as MsgUnAuthorizeAddress;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = object.symbol;
    } else {
      message.symbol = "";
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = object.address;
    } else {
      message.address = "";
    }
    return message;
  },
};

const baseMsgUnAuthorizeAddressResponse: object = {};

export const MsgUnAuthorizeAddressResponse = {
  encode(
    _: MsgUnAuthorizeAddressResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgUnAuthorizeAddressResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgUnAuthorizeAddressResponse,
    } as MsgUnAuthorizeAddressResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgUnAuthorizeAddressResponse {
    const message = {
      ...baseMsgUnAuthorizeAddressResponse,
    } as MsgUnAuthorizeAddressResponse;
    return message;
  },

  toJSON(_: MsgUnAuthorizeAddressResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgUnAuthorizeAddressResponse>
  ): MsgUnAuthorizeAddressResponse {
    const message = {
      ...baseMsgUnAuthorizeAddressResponse,
    } as MsgUnAuthorizeAddressResponse;
    return message;
  },
};

const baseMsgTransferToken: object = {
  creator: "",
  symbol: "",
  from: "",
  to: "",
  amount: 0,
};

export const MsgTransferToken = {
  encode(message: MsgTransferToken, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.symbol !== "") {
      writer.uint32(18).string(message.symbol);
    }
    if (message.from !== "") {
      writer.uint32(26).string(message.from);
    }
    if (message.to !== "") {
      writer.uint32(34).string(message.to);
    }
    if (message.amount !== 0) {
      writer.uint32(40).int64(message.amount);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgTransferToken {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgTransferToken } as MsgTransferToken;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.symbol = reader.string();
          break;
        case 3:
          message.from = reader.string();
          break;
        case 4:
          message.to = reader.string();
          break;
        case 5:
          message.amount = longToNumber(reader.int64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgTransferToken {
    const message = { ...baseMsgTransferToken } as MsgTransferToken;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = String(object.symbol);
    } else {
      message.symbol = "";
    }
    if (object.from !== undefined && object.from !== null) {
      message.from = String(object.from);
    } else {
      message.from = "";
    }
    if (object.to !== undefined && object.to !== null) {
      message.to = String(object.to);
    } else {
      message.to = "";
    }
    if (object.amount !== undefined && object.amount !== null) {
      message.amount = Number(object.amount);
    } else {
      message.amount = 0;
    }
    return message;
  },

  toJSON(message: MsgTransferToken): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.symbol !== undefined && (obj.symbol = message.symbol);
    message.from !== undefined && (obj.from = message.from);
    message.to !== undefined && (obj.to = message.to);
    message.amount !== undefined && (obj.amount = message.amount);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgTransferToken>): MsgTransferToken {
    const message = { ...baseMsgTransferToken } as MsgTransferToken;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = object.symbol;
    } else {
      message.symbol = "";
    }
    if (object.from !== undefined && object.from !== null) {
      message.from = object.from;
    } else {
      message.from = "";
    }
    if (object.to !== undefined && object.to !== null) {
      message.to = object.to;
    } else {
      message.to = "";
    }
    if (object.amount !== undefined && object.amount !== null) {
      message.amount = object.amount;
    } else {
      message.amount = 0;
    }
    return message;
  },
};

const baseMsgTransferTokenResponse: object = {};

export const MsgTransferTokenResponse = {
  encode(
    _: MsgTransferTokenResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgTransferTokenResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgTransferTokenResponse,
    } as MsgTransferTokenResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgTransferTokenResponse {
    const message = {
      ...baseMsgTransferTokenResponse,
    } as MsgTransferTokenResponse;
    return message;
  },

  toJSON(_: MsgTransferTokenResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgTransferTokenResponse>
  ): MsgTransferTokenResponse {
    const message = {
      ...baseMsgTransferTokenResponse,
    } as MsgTransferTokenResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  CreateToken(request: MsgCreateToken): Promise<MsgCreateTokenResponse>;
  UpdateToken(request: MsgUpdateToken): Promise<MsgUpdateTokenResponse>;
  AuthorizeAddress(
    request: MsgAuthorizeAddress
  ): Promise<MsgAuthorizeAddressResponse>;
  UnAuthorizeAddress(
    request: MsgUnAuthorizeAddress
  ): Promise<MsgUnAuthorizeAddressResponse>;
  /** this line is used by starport scaffolding # proto/tx/rpc */
  TransferToken(request: MsgTransferToken): Promise<MsgTransferTokenResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  CreateToken(request: MsgCreateToken): Promise<MsgCreateTokenResponse> {
    const data = MsgCreateToken.encode(request).finish();
    const promise = this.rpc.request(
      "realiotech.realionetwork.asset.Msg",
      "CreateToken",
      data
    );
    return promise.then((data) =>
      MsgCreateTokenResponse.decode(new Reader(data))
    );
  }

  UpdateToken(request: MsgUpdateToken): Promise<MsgUpdateTokenResponse> {
    const data = MsgUpdateToken.encode(request).finish();
    const promise = this.rpc.request(
      "realiotech.realionetwork.asset.Msg",
      "UpdateToken",
      data
    );
    return promise.then((data) =>
      MsgUpdateTokenResponse.decode(new Reader(data))
    );
  }

  AuthorizeAddress(
    request: MsgAuthorizeAddress
  ): Promise<MsgAuthorizeAddressResponse> {
    const data = MsgAuthorizeAddress.encode(request).finish();
    const promise = this.rpc.request(
      "realiotech.realionetwork.asset.Msg",
      "AuthorizeAddress",
      data
    );
    return promise.then((data) =>
      MsgAuthorizeAddressResponse.decode(new Reader(data))
    );
  }

  UnAuthorizeAddress(
    request: MsgUnAuthorizeAddress
  ): Promise<MsgUnAuthorizeAddressResponse> {
    const data = MsgUnAuthorizeAddress.encode(request).finish();
    const promise = this.rpc.request(
      "realiotech.realionetwork.asset.Msg",
      "UnAuthorizeAddress",
      data
    );
    return promise.then((data) =>
      MsgUnAuthorizeAddressResponse.decode(new Reader(data))
    );
  }

  TransferToken(request: MsgTransferToken): Promise<MsgTransferTokenResponse> {
    const data = MsgTransferToken.encode(request).finish();
    const promise = this.rpc.request(
      "realiotech.realionetwork.asset.Msg",
      "TransferToken",
      data
    );
    return promise.then((data) =>
      MsgTransferTokenResponse.decode(new Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

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
