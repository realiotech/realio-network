// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import { StdFee } from "@cosmjs/launchpad";
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry, OfflineSigner, EncodeObject, DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgUnAuthorizeAddress } from "./types/asset/tx";
import { MsgCreateToken } from "./types/asset/tx";
import { MsgTransferToken } from "./types/asset/tx";
import { MsgAuthorizeAddress } from "./types/asset/tx";
import { MsgUpdateToken } from "./types/asset/tx";


const types = [
  ["/realiotech.realionetwork.asset.MsgUnAuthorizeAddress", MsgUnAuthorizeAddress],
  ["/realiotech.realionetwork.asset.MsgCreateToken", MsgCreateToken],
  ["/realiotech.realionetwork.asset.MsgTransferToken", MsgTransferToken],
  ["/realiotech.realionetwork.asset.MsgAuthorizeAddress", MsgAuthorizeAddress],
  ["/realiotech.realionetwork.asset.MsgUpdateToken", MsgUpdateToken],
  
];
export const MissingWalletError = new Error("wallet is required");

export const registry = new Registry(<any>types);

const defaultFee = {
  amount: [],
  gas: "200000",
};

interface TxClientOptions {
  addr: string
}

interface SignAndBroadcastOptions {
  fee: StdFee,
  memo?: string
}

const txClient = async (wallet: OfflineSigner, { addr: addr }: TxClientOptions = { addr: "http://localhost:26657" }) => {
  if (!wallet) throw MissingWalletError;
  let client;
  if (addr) {
    client = await SigningStargateClient.connectWithSigner(addr, wallet, { registry });
  }else{
    client = await SigningStargateClient.offline( wallet, { registry });
  }
  const { address } = (await wallet.getAccounts())[0];

  return {
    signAndBroadcast: (msgs: EncodeObject[], { fee, memo }: SignAndBroadcastOptions = {fee: defaultFee, memo: ""}) => client.signAndBroadcast(address, msgs, fee,memo),
    msgUnAuthorizeAddress: (data: MsgUnAuthorizeAddress): EncodeObject => ({ typeUrl: "/realiotech.realionetwork.asset.MsgUnAuthorizeAddress", value: MsgUnAuthorizeAddress.fromPartial( data ) }),
    msgCreateToken: (data: MsgCreateToken): EncodeObject => ({ typeUrl: "/realiotech.realionetwork.asset.MsgCreateToken", value: MsgCreateToken.fromPartial( data ) }),
    msgTransferToken: (data: MsgTransferToken): EncodeObject => ({ typeUrl: "/realiotech.realionetwork.asset.MsgTransferToken", value: MsgTransferToken.fromPartial( data ) }),
    msgAuthorizeAddress: (data: MsgAuthorizeAddress): EncodeObject => ({ typeUrl: "/realiotech.realionetwork.asset.MsgAuthorizeAddress", value: MsgAuthorizeAddress.fromPartial( data ) }),
    msgUpdateToken: (data: MsgUpdateToken): EncodeObject => ({ typeUrl: "/realiotech.realionetwork.asset.MsgUpdateToken", value: MsgUpdateToken.fromPartial( data ) }),
    
  };
};

interface QueryClientOptions {
  addr: string
}

const queryClient = async ({ addr: addr }: QueryClientOptions = { addr: "http://localhost:1317" }) => {
  return new Api({ baseUrl: addr });
};

export {
  txClient,
  queryClient,
};
