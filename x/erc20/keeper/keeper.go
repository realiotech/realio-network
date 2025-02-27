package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	erc20keeper "github.com/evmos/os/x/erc20/keeper"
	"github.com/realiotech/realio-network/x/erc20/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

type Erc20Keeper struct {
	erc20keeper.Keeper
	ContractOwner collections.Map[string, string]
	authority     string
}

func NewErc20Keeper(
	erc20k erc20keeper.Keeper,
	cdc codec.Codec,
	storeService store.KVStoreService,
	authority string,
) Erc20Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return Erc20Keeper{
		Keeper:        erc20k,
		ContractOwner: collections.NewMap(sb, types.ContractOwnerKey, "contract_owner", collections.StringKey, collections.StringValue),
		authority:     authority,
	}
}

func (k Erc20Keeper) GetContractOwner(ctx sdk.Context, addr string) (common.Address, error) {
	owner, err := k.ContractOwner.Get(ctx, addr)
	if err != nil {
		return common.Address{}, err
	}
	return common.HexToAddress(owner), nil
}

func (k Erc20Keeper) IsContractOwner(ctx sdk.Context, contract common.Address, addr common.Address) (isOwner bool) {
	owner, err := k.ContractOwner.Get(ctx, contract.String())
	if err != nil {
		return false
	}
	if owner == addr.String() {
		return true
	}
	return false
}

func (k Erc20Keeper) GetContractOwners(ctx sdk.Context) []types.TokenOwner {
	owners := []types.TokenOwner{}

	k.ContractOwner.Walk(ctx, nil, func(key, value string) (stop bool, err error) {
		owners = append(owners, types.TokenOwner{
			ContractAddress: key,
			OwnerAddress:    value,
		})
		return false, nil
	})
	return owners
}

func (k Erc20Keeper) SetContractOwner(ctx sdk.Context, addr string, owner string) error {
	return k.ContractOwner.Set(ctx, addr, owner)
}
