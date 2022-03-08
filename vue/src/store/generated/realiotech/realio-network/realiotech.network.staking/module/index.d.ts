import { StdFee } from "@cosmjs/launchpad";
import { OfflineSigner, EncodeObject } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgCreateValidator } from "./types/staking/tx";
import { MsgUndelegate } from "./types/staking/tx";
import { MsgDelegate } from "./types/staking/tx";
import { MsgBeginRedelegate } from "./types/staking/tx";
import { MsgEditValidator } from "./types/staking/tx";
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
    msgCreateValidator: (data: MsgCreateValidator) => EncodeObject;
    msgUndelegate: (data: MsgUndelegate) => EncodeObject;
    msgDelegate: (data: MsgDelegate) => EncodeObject;
    msgBeginRedelegate: (data: MsgBeginRedelegate) => EncodeObject;
    msgEditValidator: (data: MsgEditValidator) => EncodeObject;
}>;
interface QueryClientOptions {
    addr: string;
}
declare const queryClient: ({ addr: addr }?: QueryClientOptions) => Promise<Api<unknown>>;
export { txClient, queryClient, };
