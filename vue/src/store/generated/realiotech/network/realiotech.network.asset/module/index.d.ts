import { StdFee } from "@cosmjs/launchpad";
import { OfflineSigner, EncodeObject } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgCreateToken } from "./types/asset/tx";
import { MsgAuthorizeAddress } from "./types/asset/tx";
import { MsgUpdateToken } from "./types/asset/tx";
import { MsgUnAuthorizeAddress } from "./types/asset/tx";
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
    msgCreateToken: (data: MsgCreateToken) => EncodeObject;
    msgAuthorizeAddress: (data: MsgAuthorizeAddress) => EncodeObject;
    msgUpdateToken: (data: MsgUpdateToken) => EncodeObject;
    msgUnAuthorizeAddress: (data: MsgUnAuthorizeAddress) => EncodeObject;
}>;
interface QueryClientOptions {
    addr: string;
}
declare const queryClient: ({ addr: addr }?: QueryClientOptions) => Promise<Api<unknown>>;
export { txClient, queryClient, };
