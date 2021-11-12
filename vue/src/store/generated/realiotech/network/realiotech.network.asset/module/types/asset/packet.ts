/* eslint-disable */
import * as Long from 'long'
import { util, configure, Writer, Reader } from 'protobufjs/minimal'

export const protobufPackage = 'realiotech.network.asset'

export interface AssetPacketData {
  noData: NoData | undefined
  /** this line is used by starport scaffolding # ibc/packet/proto/field */
  fungibleTokenTransferPacket: FungibleTokenTransferPacketData | undefined
}

export interface NoData {}

/** FungibleTokenTransferPacketData defines a struct for the packet payload */
export interface FungibleTokenTransferPacketData {
  denom: string
  amount: number
  receiver: string
  sender: string
}

/** FungibleTokenTransferPacketAck defines a struct for the packet acknowledgment */
export interface FungibleTokenTransferPacketAck {}

const baseAssetPacketData: object = {}

export const AssetPacketData = {
  encode(message: AssetPacketData, writer: Writer = Writer.create()): Writer {
    if (message.noData !== undefined) {
      NoData.encode(message.noData, writer.uint32(10).fork()).ldelim()
    }
    if (message.fungibleTokenTransferPacket !== undefined) {
      FungibleTokenTransferPacketData.encode(message.fungibleTokenTransferPacket, writer.uint32(18).fork()).ldelim()
    }
    return writer
  },

  decode(input: Reader | Uint8Array, length?: number): AssetPacketData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input
    let end = length === undefined ? reader.len : reader.pos + length
    const message = { ...baseAssetPacketData } as AssetPacketData
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.noData = NoData.decode(reader, reader.uint32())
          break
        case 2:
          message.fungibleTokenTransferPacket = FungibleTokenTransferPacketData.decode(reader, reader.uint32())
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  fromJSON(object: any): AssetPacketData {
    const message = { ...baseAssetPacketData } as AssetPacketData
    if (object.noData !== undefined && object.noData !== null) {
      message.noData = NoData.fromJSON(object.noData)
    } else {
      message.noData = undefined
    }
    if (object.fungibleTokenTransferPacket !== undefined && object.fungibleTokenTransferPacket !== null) {
      message.fungibleTokenTransferPacket = FungibleTokenTransferPacketData.fromJSON(object.fungibleTokenTransferPacket)
    } else {
      message.fungibleTokenTransferPacket = undefined
    }
    return message
  },

  toJSON(message: AssetPacketData): unknown {
    const obj: any = {}
    message.noData !== undefined && (obj.noData = message.noData ? NoData.toJSON(message.noData) : undefined)
    message.fungibleTokenTransferPacket !== undefined &&
      (obj.fungibleTokenTransferPacket = message.fungibleTokenTransferPacket
        ? FungibleTokenTransferPacketData.toJSON(message.fungibleTokenTransferPacket)
        : undefined)
    return obj
  },

  fromPartial(object: DeepPartial<AssetPacketData>): AssetPacketData {
    const message = { ...baseAssetPacketData } as AssetPacketData
    if (object.noData !== undefined && object.noData !== null) {
      message.noData = NoData.fromPartial(object.noData)
    } else {
      message.noData = undefined
    }
    if (object.fungibleTokenTransferPacket !== undefined && object.fungibleTokenTransferPacket !== null) {
      message.fungibleTokenTransferPacket = FungibleTokenTransferPacketData.fromPartial(object.fungibleTokenTransferPacket)
    } else {
      message.fungibleTokenTransferPacket = undefined
    }
    return message
  }
}

const baseNoData: object = {}

export const NoData = {
  encode(_: NoData, writer: Writer = Writer.create()): Writer {
    return writer
  },

  decode(input: Reader | Uint8Array, length?: number): NoData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input
    let end = length === undefined ? reader.len : reader.pos + length
    const message = { ...baseNoData } as NoData
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  fromJSON(_: any): NoData {
    const message = { ...baseNoData } as NoData
    return message
  },

  toJSON(_: NoData): unknown {
    const obj: any = {}
    return obj
  },

  fromPartial(_: DeepPartial<NoData>): NoData {
    const message = { ...baseNoData } as NoData
    return message
  }
}

const baseFungibleTokenTransferPacketData: object = { denom: '', amount: 0, receiver: '', sender: '' }

export const FungibleTokenTransferPacketData = {
  encode(message: FungibleTokenTransferPacketData, writer: Writer = Writer.create()): Writer {
    if (message.denom !== '') {
      writer.uint32(10).string(message.denom)
    }
    if (message.amount !== 0) {
      writer.uint32(16).uint64(message.amount)
    }
    if (message.receiver !== '') {
      writer.uint32(26).string(message.receiver)
    }
    if (message.sender !== '') {
      writer.uint32(34).string(message.sender)
    }
    return writer
  },

  decode(input: Reader | Uint8Array, length?: number): FungibleTokenTransferPacketData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input
    let end = length === undefined ? reader.len : reader.pos + length
    const message = { ...baseFungibleTokenTransferPacketData } as FungibleTokenTransferPacketData
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.denom = reader.string()
          break
        case 2:
          message.amount = longToNumber(reader.uint64() as Long)
          break
        case 3:
          message.receiver = reader.string()
          break
        case 4:
          message.sender = reader.string()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  fromJSON(object: any): FungibleTokenTransferPacketData {
    const message = { ...baseFungibleTokenTransferPacketData } as FungibleTokenTransferPacketData
    if (object.denom !== undefined && object.denom !== null) {
      message.denom = String(object.denom)
    } else {
      message.denom = ''
    }
    if (object.amount !== undefined && object.amount !== null) {
      message.amount = Number(object.amount)
    } else {
      message.amount = 0
    }
    if (object.receiver !== undefined && object.receiver !== null) {
      message.receiver = String(object.receiver)
    } else {
      message.receiver = ''
    }
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = String(object.sender)
    } else {
      message.sender = ''
    }
    return message
  },

  toJSON(message: FungibleTokenTransferPacketData): unknown {
    const obj: any = {}
    message.denom !== undefined && (obj.denom = message.denom)
    message.amount !== undefined && (obj.amount = message.amount)
    message.receiver !== undefined && (obj.receiver = message.receiver)
    message.sender !== undefined && (obj.sender = message.sender)
    return obj
  },

  fromPartial(object: DeepPartial<FungibleTokenTransferPacketData>): FungibleTokenTransferPacketData {
    const message = { ...baseFungibleTokenTransferPacketData } as FungibleTokenTransferPacketData
    if (object.denom !== undefined && object.denom !== null) {
      message.denom = object.denom
    } else {
      message.denom = ''
    }
    if (object.amount !== undefined && object.amount !== null) {
      message.amount = object.amount
    } else {
      message.amount = 0
    }
    if (object.receiver !== undefined && object.receiver !== null) {
      message.receiver = object.receiver
    } else {
      message.receiver = ''
    }
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = object.sender
    } else {
      message.sender = ''
    }
    return message
  }
}

const baseFungibleTokenTransferPacketAck: object = {}

export const FungibleTokenTransferPacketAck = {
  encode(_: FungibleTokenTransferPacketAck, writer: Writer = Writer.create()): Writer {
    return writer
  },

  decode(input: Reader | Uint8Array, length?: number): FungibleTokenTransferPacketAck {
    const reader = input instanceof Uint8Array ? new Reader(input) : input
    let end = length === undefined ? reader.len : reader.pos + length
    const message = { ...baseFungibleTokenTransferPacketAck } as FungibleTokenTransferPacketAck
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  fromJSON(_: any): FungibleTokenTransferPacketAck {
    const message = { ...baseFungibleTokenTransferPacketAck } as FungibleTokenTransferPacketAck
    return message
  },

  toJSON(_: FungibleTokenTransferPacketAck): unknown {
    const obj: any = {}
    return obj
  },

  fromPartial(_: DeepPartial<FungibleTokenTransferPacketAck>): FungibleTokenTransferPacketAck {
    const message = { ...baseFungibleTokenTransferPacketAck } as FungibleTokenTransferPacketAck
    return message
  }
}

declare var self: any | undefined
declare var window: any | undefined
var globalThis: any = (() => {
  if (typeof globalThis !== 'undefined') return globalThis
  if (typeof self !== 'undefined') return self
  if (typeof window !== 'undefined') return window
  if (typeof global !== 'undefined') return global
  throw 'Unable to locate global object'
})()

type Builtin = Date | Function | Uint8Array | string | number | undefined
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error('Value is larger than Number.MAX_SAFE_INTEGER')
  }
  return long.toNumber()
}

if (util.Long !== Long) {
  util.Long = Long as any
  configure()
}
