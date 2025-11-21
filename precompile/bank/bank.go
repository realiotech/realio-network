//
// The bank package contains the implementation of the x/bank module precompile.
// The precompiles returns all bank's information in the original decimals
// representation stored in the module.

package bank

import (
	"embed"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

var _ vm.PrecompiledContract = &Precompile{}

var (
	// Embed abi json file to the executable binary. Needed when importing as dependency.
	//
	//go:embed abi.json
	f   embed.FS
	ABI abi.ABI
)

func init() {
	var err error
	ABI, err = cmn.LoadABI(f, "abi.json")
	if err != nil {
		panic(err)
	}
}

// Precompile defines the bank precompile
type Precompile struct {
	cmn.Precompile

	abi.ABI
	bankKeeper  bankkeeper.Keeper
	erc20Keeper cmn.ERC20Keeper
}

// NewPrecompile creates a new bank Precompile instance implementing the
// PrecompiledContract interface.
func NewPrecompile(
	bankKeeper bankkeeper.Keeper,
	erc20Keeper cmn.ERC20Keeper,
) *Precompile {
	// NOTE: we set an empty gas configuration to avoid extra gas costs
	// during the run execution
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ContractAddress:      common.HexToAddress(evmtypes.BankPrecompileAddress),
		},
		ABI:         ABI,
		bankKeeper:  bankKeeper,
		erc20Keeper: erc20Keeper,
	}
}

// RequiredGas calculates the precompiled contract's base gas rate.
func (p Precompile) RequiredGas(input []byte) uint64 {
	// NOTE: This check avoid panicking when trying to decode the method ID
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]

	method, err := p.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method))
}

func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.Execute(ctx, evm.StateDB, contract, readonly)
	})
}
 
// Execute executes the precompiled contract bank query methods defined in the ABI.
func (p Precompile) Execute(ctx sdk.Context, stateDB vm.StateDB, contract *vm.Contract, readOnly bool) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.ABI, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	var bz []byte
	switch method.Name {
	// Bank queries
	case BalancesMethod:
		bz, err = p.Balances(ctx, method, args)
	case TotalSupplyMethod:
		bz, err = p.TotalSupply(ctx, method, args)
	case SupplyOfMethod:
		bz, err = p.SupplyOf(ctx, method, args)
	case SendMethod:
		bz, err = p.Send(ctx, contract.Caller(), method, args)
	case MultiSendMethod:
		bz, err = p.MultiSend(ctx, contract.Caller(), method, args)
	default:
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	return bz, err
}

// IsTransaction checks if the given method name corresponds to a transaction or query.
// It returns false since all bank methods are queries.
func (Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case SendMethod, MultiSendMethod:
		return true
	default:
		return false
	}
}

