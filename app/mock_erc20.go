package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	erc20keeper "github.com/evmos/os/x/erc20/keeper"
	"github.com/evmos/os/x/evm/core/vm"
	transferkeeper "github.com/evmos/os/x/ibc/transfer/keeper"
	"github.com/realiotech/realio-network/precompiles/erc20"
	erc20extendkeeper "github.com/realiotech/realio-network/x/erc20/keeper"
)

type MockErc20Keeper struct {
	erc20keeper.Keeper
	erc20ExtendKeeper erc20extendkeeper.Erc20Keeper

	bankKeeper     bankkeeper.Keeper
	authzKeeper    authzkeeper.Keeper
	transferKeeper *transferkeeper.Keeper
}

func NewMockErc20Keeper(
	erc20k erc20keeper.Keeper,
	bk bankkeeper.Keeper,
	ak authzkeeper.Keeper,
	tk *transferkeeper.Keeper,
	erc20ExtendKeeper erc20extendkeeper.Erc20Keeper,
) MockErc20Keeper {
	return MockErc20Keeper{
		Keeper:            erc20k,
		bankKeeper:        bk,
		authzKeeper:       ak,
		transferKeeper:    tk,
		erc20ExtendKeeper: erc20ExtendKeeper,
	}
}

func (k MockErc20Keeper) GetERC20PrecompileInstance(ctx sdk.Context, addr common.Address) (contract vm.PrecompiledContract, found bool, err error) {
	params := k.GetParams(ctx)
	if !k.IsAvailableERC20Precompile(&params, addr) {
		return nil, false, nil
	}

	address := addr.String()

	// check if the precompile is an ERC20 contract
	id := k.GetTokenPairID(ctx, address)
	if len(id) == 0 {
		return nil, false, fmt.Errorf("precompile id not found: %s", address)
	}
	pair, ok := k.GetTokenPair(ctx, id)
	if !ok {
		return nil, false, fmt.Errorf("token pair not found: %s", address)
	}

	precompile, err := erc20.NewPrecompile(pair, k.bankKeeper, k.authzKeeper, *k.transferKeeper, k.erc20ExtendKeeper)
	if err != nil {
		return nil, false, err
	}

	return precompile, true, nil
}
