import { StdFee } from "@cosmjs/launchpad";
import { Registry, OfflineSigner, EncodeObject } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgTransferToken } from "./types/asset/tx";
import { MsgUnAuthorizeAddress } from "./types/asset/tx";
import { MsgAuthorizeAddress } from "./types/asset/tx";
import { MsgCreateToken } from "./types/asset/tx";
import { MsgUpdateToken } from "./types/asset/tx";
export declare const MissingWalletError: Error;
export declare const registry: Registry;
interface TxClientOptions {
    addr: string;
}
interface SignAndBroadcastOptions {
    fee: StdFee;
    memo?: string;
}
declare const txClient: (wallet: OfflineSigner, { addr: addr }?: TxClientOptions) => Promise<{
    signAndBroadcast: (msgs: EncodeObject[], { fee, memo }?: SignAndBroadcastOptions) => any;
    msgTransferToken: (data: MsgTransferToken) => EncodeObject;
    msgUnAuthorizeAddress: (data: MsgUnAuthorizeAddress) => EncodeObject;
    msgAuthorizeAddress: (data: MsgAuthorizeAddress) => EncodeObject;
    msgCreateToken: (data: MsgCreateToken) => EncodeObject;
    msgUpdateToken: (data: MsgUpdateToken) => EncodeObject;
}>;
interface QueryClientOptions {
    addr: string;
}
declare const queryClient: ({ addr: addr }?: QueryClientOptions) => Promise<Api<unknown>>;
export { txClient, queryClient, };
