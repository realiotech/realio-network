import { StdFee } from "@cosmjs/launchpad";
import { OfflineSigner, EncodeObject } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgUpdateToken } from "./types/asset/tx";
import { MsgTransferToken } from "./types/asset/tx";
import { MsgCreateToken } from "./types/asset/tx";
import { MsgAuthorizeAddress } from "./types/asset/tx";
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
    msgUpdateToken: (data: MsgUpdateToken) => EncodeObject;
    msgTransferToken: (data: MsgTransferToken) => EncodeObject;
    msgCreateToken: (data: MsgCreateToken) => EncodeObject;
    msgAuthorizeAddress: (data: MsgAuthorizeAddress) => EncodeObject;
    msgUnAuthorizeAddress: (data: MsgUnAuthorizeAddress) => EncodeObject;
}>;
interface QueryClientOptions {
    addr: string;
}
declare const queryClient: ({ addr: addr }?: QueryClientOptions) => Promise<Api<unknown>>;
export { txClient, queryClient, };
