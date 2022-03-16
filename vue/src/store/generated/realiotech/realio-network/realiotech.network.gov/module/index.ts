// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import { StdFee } from "@cosmjs/launchpad";
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry, OfflineSigner, EncodeObject, DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgDeposit } from "./types/gov/tx";
import { MsgSubmitProposal } from "./types/gov/tx";
import { MsgVote } from "./types/gov/tx";
import { MsgVoteWeighted } from "./types/gov/tx";


const types = [
  ["/realiotech.network.gov.MsgDeposit", MsgDeposit],
  ["/realiotech.network.gov.MsgSubmitProposal", MsgSubmitProposal],
  ["/realiotech.network.gov.MsgVote", MsgVote],
  ["/realiotech.network.gov.MsgVoteWeighted", MsgVoteWeighted],
  
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
    msgDeposit: (data: MsgDeposit): EncodeObject => ({ typeUrl: "/realiotech.network.gov.MsgDeposit", value: data }),
    msgSubmitProposal: (data: MsgSubmitProposal): EncodeObject => ({ typeUrl: "/realiotech.network.gov.MsgSubmitProposal", value: data }),
    msgVote: (data: MsgVote): EncodeObject => ({ typeUrl: "/realiotech.network.gov.MsgVote", value: data }),
    msgVoteWeighted: (data: MsgVoteWeighted): EncodeObject => ({ typeUrl: "/realiotech.network.gov.MsgVoteWeighted", value: data }),
    
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
