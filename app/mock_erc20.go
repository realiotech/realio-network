package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

type MockErc20Keeper struct{}

func (k MockErc20Keeper) GetERC20PrecompileInstance(_ sdk.Context, _ common.Address) (contract vm.PrecompiledContract, found bool, err error) {
	return nil, false, nil
}
