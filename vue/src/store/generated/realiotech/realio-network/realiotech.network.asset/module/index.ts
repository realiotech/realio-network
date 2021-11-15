// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import { StdFee } from "@cosmjs/launchpad";
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry, OfflineSigner, EncodeObject, DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgUpdateToken } from "./types/asset/tx";
import { MsgTransferToken } from "./types/asset/tx";
import { MsgCreateToken } from "./types/asset/tx";
import { MsgAuthorizeAddress } from "./types/asset/tx";
import { MsgUnAuthorizeAddress } from "./types/asset/tx";


const types = [
  ["/realiotech.network.asset.MsgUpdateToken", MsgUpdateToken],
  ["/realiotech.network.asset.MsgTransferToken", MsgTransferToken],
  ["/realiotech.network.asset.MsgCreateToken", MsgCreateToken],
  ["/realiotech.network.asset.MsgAuthorizeAddress", MsgAuthorizeAddress],
  ["/realiotech.network.asset.MsgUnAuthorizeAddress", MsgUnAuthorizeAddress],
  
];
export const MissingWalletError = new Error("wallet is required");

const registry = new Registry(<any>types);

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

  const client = await SigningStargateClient.connectWithSigner(addr, wallet, { registry });
  const { address } = (await wallet.getAccounts())[0];

  return {
    signAndBroadcast: (msgs: EncodeObject[], { fee, memo }: SignAndBroadcastOptions = {fee: defaultFee, memo: ""}) => client.signAndBroadcast(address, msgs, fee,memo),
    msgUpdateToken: (data: MsgUpdateToken): EncodeObject => ({ typeUrl: "/realiotech.network.asset.MsgUpdateToken", value: data }),
    msgTransferToken: (data: MsgTransferToken): EncodeObject => ({ typeUrl: "/realiotech.network.asset.MsgTransferToken", value: data }),
    msgCreateToken: (data: MsgCreateToken): EncodeObject => ({ typeUrl: "/realiotech.network.asset.MsgCreateToken", value: data }),
    msgAuthorizeAddress: (data: MsgAuthorizeAddress): EncodeObject => ({ typeUrl: "/realiotech.network.asset.MsgAuthorizeAddress", value: data }),
    msgUnAuthorizeAddress: (data: MsgUnAuthorizeAddress): EncodeObject => ({ typeUrl: "/realiotech.network.asset.MsgUnAuthorizeAddress", value: data }),
    
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
