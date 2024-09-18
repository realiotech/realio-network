package transfer_auth

import (
	"github.com/realiotech/realio-network/x/asset/keeper"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
)

var (
	_ keeper.RestrictionChecker = (*TransferAuthPriviledge)(nil)
)

const priv_name = "transfer_auth"

type TransferAuthPriviledge struct {
	storeKey storetypes.StoreKey
}

func NewTransferAuthPriviledge(sk storetypes.StoreKey) TransferAuthPriviledge {
	return TransferAuthPriviledge{
		storeKey: sk,
	}
}

func (tp TransferAuthPriviledge) Name() string {
	return priv_name
}

func (tp TransferAuthPriviledge) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*assettypes.PrivilegeMsgI)(nil),
		&MsgUpdateAllowList{},
	)
}

func (tp TransferAuthPriviledge) IsAllow(ctx sdk.Context, tokenID string, sender string) (bool, error) {
	return tp.CheckAddressIsWhitelisted(ctx, tokenID, sender), nil
}
