import { StdFee } from "@cosmjs/launchpad";
import { OfflineSigner, EncodeObject } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgWithdrawDelegatorReward } from "./types/distribution/tx";
import { MsgFundCommunityPool } from "./types/distribution/tx";
import { MsgSetWithdrawAddress } from "./types/distribution/tx";
import { MsgWithdrawValidatorCommission } from "./types/distribution/tx";
export declare const MissingWalletError: Error;
interface TxClientOptions {
    addr: string;
}
interface SignAndBroadcastOptions {
    fee: StdFee;
    memo?: string;
}
declare const txClient: (wallet: OfflineSigner, { addr: addr }?: TxClientOptions) => Promise<{
    signAndBroadcast: (msgs: EncodeObject[], { fee, memo }?: SignAndBroadcastOptions) => Promise<import("@cosmjs/stargate").BroadcastTxResponse>;
    msgWithdrawDelegatorReward: (data: MsgWithdrawDelegatorReward) => EncodeObject;
    msgFundCommunityPool: (data: MsgFundCommunityPool) => EncodeObject;
    msgSetWithdrawAddress: (data: MsgSetWithdrawAddress) => EncodeObject;
    msgWithdrawValidatorCommission: (data: MsgWithdrawValidatorCommission) => EncodeObject;
}>;
interface QueryClientOptions {
    addr: string;
}
declare const queryClient: ({ addr: addr }?: QueryClientOptions) => Promise<Api<unknown>>;
export { txClient, queryClient, };
