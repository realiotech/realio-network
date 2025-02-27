// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/os/precompiles/common"
	"github.com/evmos/os/x/evm/core/vm"
	evmtypes "github.com/evmos/os/x/evm/types"
	bridgetypes "github.com/realiotech/realio-network/x/bridge/types"
)

const (
	// TransferMethod defines the ABI method name for the ERC-20 transfer
	// transaction.
	TransferMethod = "transfer"
	// TransferFromMethod defines the ABI method name for the ERC-20 transferFrom
	// transaction.
	TransferFromMethod = "transferFrom"

	BurnMethod = "burn"

	BurnFromMethod = "burnFrom"

	MintMethod = "mint"

	RenounceOwnership = "renounceOwnership"

	TransferOwnership = "transferOwnership"
)

// SendMsgURL defines the authorization type for MsgSend
var SendMsgURL = sdk.MsgTypeURL(&banktypes.MsgSend{})

// Transfer executes a direct transfer from the caller address to the
// destination address.
func (p *Precompile) Transfer(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	from := contract.CallerAddress
	to, amount, err := ParseTransferArgs(args)
	if err != nil {
		return nil, err
	}

	return p.transfer(ctx, contract, stateDB, method, from, to, amount)
}

// TransferFrom executes a transfer on behalf of the specified from address in
// the call data to the destination address.
func (p *Precompile) TransferFrom(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	from, to, amount, err := ParseTransferFromArgs(args)
	if err != nil {
		return nil, err
	}

	return p.transfer(ctx, contract, stateDB, method, from, to, amount)
}

// transfer is a common function that handles transfers for the ERC-20 Transfer
// and TransferFrom methods. It executes a bank Send message if the spender is
// the sender of the transfer, otherwise it executes an authorization.
func (p *Precompile) transfer(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	from, to common.Address,
	amount *big.Int,
) (data []byte, err error) {
	coins := sdk.Coins{{Denom: p.tokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}}

	msg := banktypes.NewMsgSend(from.Bytes(), to.Bytes(), coins)

	if err = msg.Amount.Validate(); err != nil {
		return nil, err
	}

	isTransferFrom := method.Name == TransferFromMethod
	owner := sdk.AccAddress(from.Bytes())
	spenderAddr := contract.CallerAddress
	spender := sdk.AccAddress(spenderAddr.Bytes()) // aka. grantee
	ownerIsSpender := spender.Equals(owner)

	var prevAllowance *big.Int
	if ownerIsSpender {
		msgSrv := bankkeeper.NewMsgServerImpl(p.BankKeeper)
		_, err = msgSrv.Send(ctx, msg)
	} else {
		_, _, prevAllowance, err = GetAuthzExpirationAndAllowance(p.AuthzKeeper, ctx, spenderAddr, from, p.tokenPair.Denom)
		if err != nil {
			return nil, ConvertErrToERC20Error(errorsmod.Wrap(err, authz.ErrNoAuthorizationFound.Error()))
		}

		_, err = p.AuthzKeeper.DispatchActions(ctx, spender, []sdk.Msg{msg})
	}

	if err != nil {
		err = ConvertErrToERC20Error(err)
		// This should return an error to avoid the contract from being executed and an event being emitted
		return nil, err
	}

	evmDenom := evmtypes.GetEVMCoinDenom()
	if p.tokenPair.Denom == evmDenom {
		convertedAmount := evmtypes.ConvertAmountTo18DecimalsBigInt(amount)
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(from, convertedAmount, cmn.Sub),
			cmn.NewBalanceChangeEntry(to, convertedAmount, cmn.Add))
	}

	if err = p.EmitTransferEvent(ctx, stateDB, from, to, amount); err != nil {
		return nil, err
	}

	// NOTE: if it's a direct transfer, we return here but if used through transferFrom,
	// we need to emit the approval event with the new allowance.
	if !isTransferFrom {
		return method.Outputs.Pack(true)
	}

	var newAllowance *big.Int
	if ownerIsSpender {
		// NOTE: in case the spender is the owner we emit an approval event with
		// the maxUint256 value.
		newAllowance = abi.MaxUint256
	} else {
		newAllowance = new(big.Int).Sub(prevAllowance, amount)
	}

	if err = p.EmitApprovalEvent(ctx, stateDB, from, spenderAddr, newAllowance); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

func (p *Precompile) Mint(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	to, amount, err := ParseMintArgs(args)
	if err != nil {
		return nil, err
	}

	return p.mint(ctx, contract, stateDB, method, to, amount)
}

func (p *Precompile) mint(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	to common.Address,
	amount *big.Int,
) (data []byte, err error) {

	minter := contract.CallerAddress
	denom := p.tokenPair.Denom

	// Check if minter is token owner
	isOwner := p.ExtendKeeper.IsContractOwner(ctx, p.tokenPair.GetERC20Contract(), minter)
	if !isOwner {
		return nil, ConvertErrToERC20Error(fmt.Errorf("minter is not token owner"))
	}
	
	mintToAddr := sdk.AccAddress(to.Bytes())

	coins := sdk.Coins{{Denom: denom, Amount: math.NewIntFromBigInt(amount)}}

	// Mint coins to bridge module then transfer to minter addr
	err = p.BankKeeper.MintCoins(ctx, bridgetypes.ModuleName, coins)
	if err != nil {
		return nil, ConvertErrToERC20Error(err)
	}

	fmt.Println("MintCoins", err)

	err = p.BankKeeper.SendCoinsFromModuleToAccount(ctx, bridgetypes.ModuleName, mintToAddr, coins)
	if err != nil {
		return nil, ConvertErrToERC20Error(err)
	}

	evmDenom := evmtypes.GetEVMCoinDenom()
	if denom == evmDenom {
		convertedAmount := evmtypes.ConvertAmountTo18DecimalsBigInt(amount)
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(to, convertedAmount, cmn.Add))
	}

	// if err = p.EmitMintEvent(ctx, stateDB, to, amount); err != nil {
	// 	return nil, err
	// }

	return method.Outputs.Pack()
}

func (p *Precompile) Burn(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	from := contract.CallerAddress
	amount, err := ParseBurnArgs(args)
	if err != nil {
		return nil, err
	}

	return p.burn(ctx, contract, stateDB, method, from, amount)
}

func (p *Precompile) BurnFrom(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	from, amount, err := ParseBurnFromArgs(args)
	if err != nil {
		return nil, err
	}

	return p.burn(ctx, contract, stateDB, method, from, amount)
}

func (p *Precompile) burn(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	from common.Address,
	amount *big.Int,
) (data []byte, err error) {

	coins := sdk.Coins{{Denom: p.tokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}}

	if err = coins.Validate(); err != nil {
		return nil, err
	}

	isBurnFrom := method.Name == BurnFromMethod
	owner := sdk.AccAddress(from.Bytes())
	spenderAddr := contract.CallerAddress
	spender := sdk.AccAddress(spenderAddr.Bytes()) // aka. grantee
	ownerIsSpender := spender.Equals(owner)

	var prevAllowance *big.Int
	if ownerIsSpender {
		err := p.BankKeeper.SendCoinsFromAccountToModule(ctx, owner, bridgetypes.ModuleName, coins)
		if err != nil {
			return nil, err
		}

		err = p.BankKeeper.BurnCoins(ctx, bridgetypes.ModuleName, coins)
		if err != nil {
			return nil, err
		}
	} else {
		_, _, prevAllowance, err = GetAuthzExpirationAndAllowance(p.AuthzKeeper, ctx, spenderAddr, from, p.tokenPair.Denom)
		if err != nil {
			return nil, ConvertErrToERC20Error(errorsmod.Wrap(err, authz.ErrNoAuthorizationFound.Error()))
		}

		// Send to module addr then burn
		msg := banktypes.NewMsgSend(from.Bytes(), sdkaddress.Module(bridgetypes.ModuleName), coins)

		_, err = p.AuthzKeeper.DispatchActions(ctx, spender, []sdk.Msg{msg})
		if err != nil {
			return nil, ConvertErrToERC20Error(err)
		}

		err = p.BankKeeper.BurnCoins(ctx, bridgetypes.ModuleName, coins)
	}

	if err != nil {
		err = ConvertErrToERC20Error(err)
		// This should return an error to avoid the contract from being executed and an event being emitted
		return nil, err
	}

	evmDenom := evmtypes.GetEVMCoinDenom()
	if p.tokenPair.Denom == evmDenom {
		convertedAmount := evmtypes.ConvertAmountTo18DecimalsBigInt(amount)
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(from, convertedAmount, cmn.Sub),
			cmn.NewBalanceChangeEntry(from, convertedAmount, cmn.Add))
	}

	if err = p.EmitBurnEvent(ctx, stateDB, from, amount); err != nil {
		return nil, err
	}

	// NOTE: if it's a direct transfer, we return here but if used through transferFrom,
	// we need to emit the approval event with the new allowance.
	if !isBurnFrom {
		return method.Outputs.Pack()
	}

	var newAllowance *big.Int
	if ownerIsSpender {
		// NOTE: in case the spender is the owner we emit an approval event with
		// the maxUint256 value.
		newAllowance = abi.MaxUint256
	} else {
		newAllowance = new(big.Int).Sub(prevAllowance, amount)
	}

	if err = p.EmitApprovalEvent(ctx, stateDB, from, spenderAddr, newAllowance); err != nil {
		return nil, err
	}

	return method.Outputs.Pack()
	
}

func (p *Precompile) RenounceOwnership(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Check if caller is contract owner
	caller := contract.CallerAddress
	isOwner := p.ExtendKeeper.IsContractOwner(ctx, p.tokenPair.GetERC20Contract(), caller)
	if !isOwner {
		return nil, ConvertErrToERC20Error(fmt.Errorf("caller is not contract owner"))
	}

	// Set owner to zero
	err := p.ExtendKeeper.SetContractOwner(ctx, p.tokenPair.GetERC20Contract().String(), "")
	if err != nil {
		return nil, ConvertErrToERC20Error(err)
	}

	return method.Outputs.Pack()
}

func (p *Precompile) TransferOwnership(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {

	newOwner, err := ParseTransferOwnershipArgs(args)
	if err != nil {
		return nil, err
	}

	// Check if caller is contract owner
	caller := contract.CallerAddress
	isOwner := p.ExtendKeeper.IsContractOwner(ctx, p.tokenPair.GetERC20Contract(), caller)
	if !isOwner {
		return nil, ConvertErrToERC20Error(fmt.Errorf("caller is not contract owner"))
	}

	// Set owner to newOwner
	err = p.ExtendKeeper.SetContractOwner(ctx, p.tokenPair.GetERC20Contract().String(), newOwner.String())
	if err != nil {
		return nil, ConvertErrToERC20Error(err)
	}

	return method.Outputs.Pack()
}
