package transfer_auth

import (
	"fmt"

	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
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

func (tp TransferAuthPriviledge) SetupAllowListForToken(ctx sdk.Context, tokenId string, list map[string]bool) error {
	store := prefix.NewStore(ctx.KVStore(tp.storeKey), types.TokenKey)
	key := []byte(tokenId)
	bz := store.Get(key)

	if bz != nil {
		return fmt.Errorf("token ID %s already have an allow list", tokenId)
	}

	allowAddrs := AllowAddrs{
		Addrs: list,
	}

	bz, err := tp.cdc.Marshal(&allowAddrs)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

func (tp TransferAuthPriviledge) GetAddrList(ctx sdk.Context, tokenId string) (AllowAddrs, error) {
	store := prefix.NewStore(ctx.KVStore(tp.storeKey), types.TokenKey)
	key := []byte(tokenId)
	bz := store.Get(key)

	if bz == nil {
		return AllowAddrs{
			Addrs: map[string]bool{},
		}, nil
	}

	var allowAddrs AllowAddrs
	err := tp.cdc.Unmarshal(bz, &allowAddrs)
	if err != nil {
		return AllowAddrs{
			Addrs: map[string]bool{},
		}, err
	}

	return allowAddrs, nil
}

func (tp TransferAuthPriviledge) AddAddr(ctx sdk.Context, addr, tokenId string) error {
	store := prefix.NewStore(ctx.KVStore(tp.storeKey), types.TokenKey)
	key := []byte(tokenId)
	bz := store.Get(key)
	var allowAddrs *AllowAddrs

	if bz == nil {
		allowAddrs = &AllowAddrs{
			Addrs: map[string]bool{},
		}
	}

	err := tp.cdc.Unmarshal(bz, allowAddrs)
	if err != nil {
		return err
	}

	allowAddrs.Addrs[addr] = true

	bz, err = tp.cdc.Marshal(allowAddrs)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

func (tp TransferAuthPriviledge) RemoveAddr(ctx sdk.Context, addr, tokenId string) error {
	store := prefix.NewStore(ctx.KVStore(tp.storeKey), types.TokenKey)
	key := []byte(tokenId)
	bz := store.Get(key)
	var allowAddrs *AllowAddrs

	if bz == nil {
		allowAddrs = &AllowAddrs{
			Addrs: map[string]bool{},
		}
	}

	err := tp.cdc.Unmarshal(bz, allowAddrs)
	if err != nil {
		return err
	}

	allowAddrs.Addrs[addr] = false

	bz, err = tp.cdc.Marshal(allowAddrs)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

func (tp TransferAuthPriviledge) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {}

func (tp TransferAuthPriviledge) IsAllow(ctx sdk.Context, tokenID string, sender string) (bool, error) {
	allowAddrs, err := tp.GetAddrList(ctx, tokenID)
	if err != nil {
		return false, err
	}

	var isAllow bool
	isAllow, has := allowAddrs.Addrs[sender]
	if !has {
		isAllow = false
	}

	return isAllow, nil

}
