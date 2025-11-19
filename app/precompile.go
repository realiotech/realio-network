package app

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"
	erc20Keeper "github.com/cosmos/evm/x/erc20/keeper"
	transferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v10/modules/core/04-channel/keeper"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/codec"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	precompiletypes "github.com/cosmos/evm/precompiles/types"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	precompileMultiStaking "github.com/realiotech/realio-network/precompile/multistaking"
)

// NewAvailableStaticPrecompiles returns the list of all available static precompiled contracts from Cosmos EVM.
//
// NOTE: this should only be used during initialization of the Keeper.
func NewAvailableStaticPrecompiles(
	cdc codec.Codec,
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	bankKeeper cmn.BankKeeper,
	erc20Keeper erc20Keeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	channelKeeper *channelkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	slashingKeeper slashingkeeper.Keeper,
	multiStakingKeeper multistakingkeeper.Keeper,
	appCodec codec.Codec,
	addrCodec address.Codec,
	valAddrCodec address.Codec,
) map[common.Address]vm.PrecompiledContract {
	precompiles := precompiletypes.DefaultStaticPrecompiles(
		stakingKeeper,
		distributionKeeper,
		bankKeeper,
		&erc20Keeper,
		&transferKeeper,
		channelKeeper,
		govKeeper,
		slashingKeeper,
		appCodec,
	)

	mulStakingPrecompile, err := precompileMultiStaking.NewPrecompile(cdc, stakingKeeper, multiStakingKeeper, erc20Keeper, addrCodec, valAddrCodec)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bank precompile: %w", err))
	}

	precompiles[mulStakingPrecompile.Address()] = mulStakingPrecompile

	return precompiles
}
