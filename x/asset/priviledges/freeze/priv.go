package freeze

import (
	"github.com/realiotech/realio-network/x/asset/keeper"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
)

var (
	_ keeper.RestrictionChecker = (*FreezePriviledge)(nil)
)

const priv_name = "transfer_auth"

type FreezePriviledge struct {
	storeKey storetypes.StoreKey
}

func NewFreezePriviledge(sk storetypes.StoreKey) FreezePriviledge {
	return FreezePriviledge{
		storeKey: sk,
	}
}

func (tp FreezePriviledge) Name() string {
	return priv_name
}

func (tp FreezePriviledge) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*assettypes.PrivilegeMsgI)(nil),
		&MsgFreezeToken{},
		&MsgUnfreezeToken{},
	)
}

func (tp FreezePriviledge) IsAllow(ctx sdk.Context, tokenID string, sender string) (bool, error) {
	return tp.CheckAddressIsFreezed(ctx, tokenID, sender), nil
}
