import { StdFee } from "@cosmjs/launchpad";
import { OfflineSigner, EncodeObject } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgDeposit } from "./types/gov/tx";
import { MsgSubmitProposal } from "./types/gov/tx";
import { MsgVote } from "./types/gov/tx";
import { MsgVoteWeighted } from "./types/gov/tx";
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
    msgDeposit: (data: MsgDeposit) => EncodeObject;
    msgSubmitProposal: (data: MsgSubmitProposal) => EncodeObject;
    msgVote: (data: MsgVote) => EncodeObject;
    msgVoteWeighted: (data: MsgVoteWeighted) => EncodeObject;
}>;
interface QueryClientOptions {
    addr: string;
}
declare const queryClient: ({ addr: addr }?: QueryClientOptions) => Promise<Api<unknown>>;
export { txClient, queryClient, };
