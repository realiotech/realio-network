package transfer_auth

import (
	"github.com/realiotech/realio-network/x/asset/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ keeper.RestrictionChecker = (*TransferAuthPriviledge)(nil)
)

const priv_name = "transfer_auth"

type TransferAuthPriviledge struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

func NewTransferAuthPriviledge(cdc codec.BinaryCodec, sk storetypes.StoreKey) TransferAuthPriviledge {
	return TransferAuthPriviledge{
		cdc:      cdc,
		storeKey: sk,
	}
}

func (tp TransferAuthPriviledge) Name() string {
	return priv_name
}

func (tp TransferAuthPriviledge) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {}

func (tp TransferAuthPriviledge) IsAllow(ctx sdk.Context, tokenID string, sender string) (bool, error) {
	return tp.CheckAddressIsWhitelisted(ctx, tokenID, sender), nil
}
