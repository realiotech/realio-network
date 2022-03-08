// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgCreateValidator } from "./types/staking/tx";
import { MsgUndelegate } from "./types/staking/tx";
import { MsgDelegate } from "./types/staking/tx";
import { MsgBeginRedelegate } from "./types/staking/tx";
import { MsgEditValidator } from "./types/staking/tx";
const types = [
    ["/realiotech.network.staking.MsgCreateValidator", MsgCreateValidator],
    ["/realiotech.network.staking.MsgUndelegate", MsgUndelegate],
    ["/realiotech.network.staking.MsgDelegate", MsgDelegate],
    ["/realiotech.network.staking.MsgBeginRedelegate", MsgBeginRedelegate],
    ["/realiotech.network.staking.MsgEditValidator", MsgEditValidator],
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
        msgCreateValidator: (data) => ({ typeUrl: "/realiotech.network.staking.MsgCreateValidator", value: data }),
        msgUndelegate: (data) => ({ typeUrl: "/realiotech.network.staking.MsgUndelegate", value: data }),
        msgDelegate: (data) => ({ typeUrl: "/realiotech.network.staking.MsgDelegate", value: data }),
        msgBeginRedelegate: (data) => ({ typeUrl: "/realiotech.network.staking.MsgBeginRedelegate", value: data }),
        msgEditValidator: (data) => ({ typeUrl: "/realiotech.network.staking.MsgEditValidator", value: data }),
    };
};
const queryClient = async ({ addr: addr } = { addr: "http://localhost:1317" }) => {
    return new Api({ baseUrl: addr });
};
export { txClient, queryClient, };
