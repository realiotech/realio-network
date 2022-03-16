// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry } from "@cosmjs/proto-signing";
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
const registry = new Registry(types);
const defaultFee = {
    amount: [],
    gas: "200000",
};
const txClient = async (wallet, { addr: addr } = { addr: "http://localhost:26657" }) => {
    if (!wallet)
        throw MissingWalletError;
    const client = await SigningStargateClient.connectWithSigner(addr, wallet, { registry });
    const { address } = (await wallet.getAccounts())[0];
    return {
        signAndBroadcast: (msgs, { fee, memo } = { fee: defaultFee, memo: "" }) => client.signAndBroadcast(address, msgs, fee, memo),
        msgDeposit: (data) => ({ typeUrl: "/realiotech.network.gov.MsgDeposit", value: data }),
        msgSubmitProposal: (data) => ({ typeUrl: "/realiotech.network.gov.MsgSubmitProposal", value: data }),
        msgVote: (data) => ({ typeUrl: "/realiotech.network.gov.MsgVote", value: data }),
        msgVoteWeighted: (data) => ({ typeUrl: "/realiotech.network.gov.MsgVoteWeighted", value: data }),
    };
};
const queryClient = async ({ addr: addr } = { addr: "http://localhost:1317" }) => {
    return new Api({ baseUrl: addr });
};
export { txClient, queryClient, };
