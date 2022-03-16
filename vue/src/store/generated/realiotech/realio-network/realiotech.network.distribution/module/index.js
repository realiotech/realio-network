// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgWithdrawDelegatorReward } from "./types/distribution/tx";
import { MsgFundCommunityPool } from "./types/distribution/tx";
import { MsgSetWithdrawAddress } from "./types/distribution/tx";
import { MsgWithdrawValidatorCommission } from "./types/distribution/tx";
const types = [
    ["/realiotech.network.distribution.MsgWithdrawDelegatorReward", MsgWithdrawDelegatorReward],
    ["/realiotech.network.distribution.MsgFundCommunityPool", MsgFundCommunityPool],
    ["/realiotech.network.distribution.MsgSetWithdrawAddress", MsgSetWithdrawAddress],
    ["/realiotech.network.distribution.MsgWithdrawValidatorCommission", MsgWithdrawValidatorCommission],
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
        msgWithdrawDelegatorReward: (data) => ({ typeUrl: "/realiotech.network.distribution.MsgWithdrawDelegatorReward", value: data }),
        msgFundCommunityPool: (data) => ({ typeUrl: "/realiotech.network.distribution.MsgFundCommunityPool", value: data }),
        msgSetWithdrawAddress: (data) => ({ typeUrl: "/realiotech.network.distribution.MsgSetWithdrawAddress", value: data }),
        msgWithdrawValidatorCommission: (data) => ({ typeUrl: "/realiotech.network.distribution.MsgWithdrawValidatorCommission", value: data }),
    };
};
const queryClient = async ({ addr: addr } = { addr: "http://localhost:1317" }) => {
    return new Api({ baseUrl: addr });
};
export { txClient, queryClient, };
