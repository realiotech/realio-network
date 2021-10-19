// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgUnAuthorizeAddress } from "./types/asset/tx";
import { MsgUpdateToken } from "./types/asset/tx";
import { MsgCreateToken } from "./types/asset/tx";
import { MsgAuthorizeAddress } from "./types/asset/tx";
const types = [
    ["/realiotech.network.asset.MsgUnAuthorizeAddress", MsgUnAuthorizeAddress],
    ["/realiotech.network.asset.MsgUpdateToken", MsgUpdateToken],
    ["/realiotech.network.asset.MsgCreateToken", MsgCreateToken],
    ["/realiotech.network.asset.MsgAuthorizeAddress", MsgAuthorizeAddress],
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
        msgUnAuthorizeAddress: (data) => ({ typeUrl: "/realiotech.network.asset.MsgUnAuthorizeAddress", value: data }),
        msgUpdateToken: (data) => ({ typeUrl: "/realiotech.network.asset.MsgUpdateToken", value: data }),
        msgCreateToken: (data) => ({ typeUrl: "/realiotech.network.asset.MsgCreateToken", value: data }),
        msgAuthorizeAddress: (data) => ({ typeUrl: "/realiotech.network.asset.MsgAuthorizeAddress", value: data }),
    };
};
const queryClient = async ({ addr: addr } = { addr: "http://localhost:1317" }) => {
    return new Api({ baseUrl: addr });
};
export { txClient, queryClient, };
