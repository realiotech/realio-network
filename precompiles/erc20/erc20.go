// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"embed"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	auth "github.com/evmos/os/precompiles/authorization"
	cmn "github.com/evmos/os/precompiles/common"
	erc20types "github.com/evmos/os/x/erc20/types"
	"github.com/evmos/os/x/evm/core/vm"
	transferkeeper "github.com/evmos/os/x/ibc/transfer/keeper"
	erc20extendkeeper "github.com/realiotech/realio-network/x/erc20/keeper"
)

const (
	// abiPath defines the path to the ERC-20 precompile ABI JSON file.
	abiPath = "abi.json"

	GasTransfer          = 3_000_000
	GasMint              = 3_000_000
	GasBurn              = 3_000_000
	GasApprove           = 30_956
	GasIncreaseAllowance = 34_605
	GasDecreaseAllowance = 34_519
	GasName              = 3_421
	GasSymbol            = 3_464
	GasDecimals          = 427
	GasTotalSupply       = 2_477
	GasBalanceOf         = 2_851
	GasAllowance         = 3_246
	GasOwner = 3_422
	GasRenounceOwnership = 33_320
	GasTransferOwnership = 35_333
)

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

var _ vm.PrecompiledContract = &Precompile{}

// Precompile defines the precompiled contract for ERC-20.
type Precompile struct {
	cmn.Precompile
	tokenPair      erc20types.TokenPair
	transferKeeper transferkeeper.Keeper
	// BankKeeper is a public field so that the werc20 precompile can use it.
	BankKeeper   bankkeeper.Keeper
	ExtendKeeper erc20extendkeeper.Erc20Keeper
}

// NewPrecompile creates a new ERC-20 Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	tokenPair erc20types.TokenPair,
	bankKeeper bankkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	erc20ExtendKeeper erc20extendkeeper.Erc20Keeper,
) (*Precompile, error) {
	newABI, err := cmn.LoadABI(f, abiPath)
	if err != nil {
		return nil, err
	}

	p := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newABI,
			AuthzKeeper:          authzKeeper,
			ApprovalExpiration:   cmn.DefaultExpirationDuration,
			KvGasConfig:          storetypes.GasConfig{},
			TransientKVGasConfig: storetypes.GasConfig{},
		},
		tokenPair:      tokenPair,
		BankKeeper:     bankKeeper,
		transferKeeper: transferKeeper,
		ExtendKeeper:   erc20ExtendKeeper,
	}
	fmt.Println("ERC20 contract addr", p.tokenPair.GetERC20Contract())
	// Address defines the address of the ERC-20 precompile contract.
	p.SetAddress(p.tokenPair.GetERC20Contract())
	return p, nil
}

// RequiredGas calculates the contract gas used for the
func (p Precompile) RequiredGas(input []byte) uint64 {
	// NOTE: This check avoid panicking when trying to decode the method ID
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]
	method, err := p.MethodById(methodID)
	if err != nil {
		return 0
	}

	// TODO: these values were obtained from Remix using the ERC20.sol from OpenZeppelin.
	// We should execute the transactions using the ERC20MinterBurnerDecimals.sol from Evmos testnet
	// to ensure parity in the values.
	switch method.Name {
	// ERC-20 transactions
	case TransferMethod:
		return GasTransfer
	case TransferFromMethod:
		return GasTransfer
	case MintMethod:
		return GasMint
	case BurnMethod:
		return GasBurn
	case BurnFromMethod:
		return GasBurn
	case auth.ApproveMethod:
		return GasApprove
	case auth.IncreaseAllowanceMethod:
		return GasIncreaseAllowance
	case auth.DecreaseAllowanceMethod:
		return GasDecreaseAllowance
	// ERC-20 queries
	case NameMethod:
		return GasName
	case SymbolMethod:
		return GasSymbol
	case DecimalsMethod:
		return GasDecimals
	case TotalSupplyMethod:
		return GasTotalSupply
	case BalanceOfMethod:
		return GasBalanceOf
	case auth.AllowanceMethod:
		return GasAllowance

	// Ownable transactions:
	case RenounceOwnership:
		return GasRenounceOwnership
	case TransferOwnership:
		return GasTransferOwnership
	// Ownable query
	case OwnerMethod:
		return GasOwner
	default:
		return 0
	}
}

// Run executes the precompiled contract ERC-20 methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	fmt.Println("Go to Precompile.Run")
	// ERC20 precompiles cannot receive funds because they are not managed by an
	// EOA and will not be possible to recover funds sent to an instance of
	// them.This check is a safety measure because currently funds cannot be
	// received due to the lack of a fallback handler.
	if value := contract.Value(); value.Sign() == 1 {
		return nil, fmt.Errorf(ErrCannotReceiveFunds, contract.Value().String())
	}

	ctx, stateDB, snapshot, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	fmt.Println("RunSetup", err, method, args)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	fmt.Println("Method name: ", method.Name)
	bz, err = p.HandleMethod(ctx, contract, stateDB, method, args)
	if err != nil {
		return nil, err
	}

	cost := ctx.GasMeter().GasConsumed() - initialGas

	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}
	if err := p.AddJournalEntries(stateDB, snapshot); err != nil {
		return nil, err
	}
	return bz, nil
}

// IsTransaction checks if the given method name corresponds to a transaction or query.
func (Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case TransferMethod,
		TransferFromMethod,
		MintMethod,
		BurnMethod,
		BurnFromMethod,
		auth.ApproveMethod,
		auth.IncreaseAllowanceMethod,
		auth.DecreaseAllowanceMethod:
		return true
	default:
		return false
	}
}

// HandleMethod handles the execution of each of the ERC-20 methods.
func (p *Precompile) HandleMethod(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) (bz []byte, err error) {
	fmt.Println("HandleMethod", method.Name)
	switch method.Name {
	// ERC-20 transactions
	case TransferMethod:
		bz, err = p.Transfer(ctx, contract, stateDB, method, args)
	case TransferFromMethod:
		bz, err = p.TransferFrom(ctx, contract, stateDB, method, args)
	case auth.ApproveMethod:
		bz, err = p.Approve(ctx, contract, stateDB, method, args)
	case auth.IncreaseAllowanceMethod:
		bz, err = p.IncreaseAllowance(ctx, contract, stateDB, method, args)
	case auth.DecreaseAllowanceMethod:
		bz, err = p.DecreaseAllowance(ctx, contract, stateDB, method, args)
	case MintMethod:
		bz, err = p.Mint(ctx, contract, stateDB, method, args)
	case BurnMethod:
		bz, err = p.Burn(ctx, contract, stateDB, method, args)
	case BurnFromMethod:
		bz, err = p.BurnFrom(ctx, contract, stateDB, method, args)
	// ERC-20 queries
	case NameMethod:
		bz, err = p.Name(ctx, contract, stateDB, method, args)
	case SymbolMethod:
		bz, err = p.Symbol(ctx, contract, stateDB, method, args)
	case DecimalsMethod:
		bz, err = p.Decimals(ctx, contract, stateDB, method, args)
	case TotalSupplyMethod:
		bz, err = p.TotalSupply(ctx, contract, stateDB, method, args)
	case BalanceOfMethod:
		bz, err = p.BalanceOf(ctx, contract, stateDB, method, args)
	case auth.AllowanceMethod:
		bz, err = p.Allowance(ctx, contract, stateDB, method, args)

	// Ownable transactions
	case RenounceOwnership:
		bz, err = p.RenounceOwnership(ctx, contract, stateDB, method, args)
	case TransferOwnership:
		bz, err = p.TransferOwnership(ctx, contract, stateDB, method, args)
	// Ownable transactions
	case OwnerMethod:
		bz, err = p.Owner(ctx, contract, stateDB, method, args)
	default:
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	return bz, err
}
