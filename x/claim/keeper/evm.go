package keeper

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/evm/contracts"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	"github.com/cosmos/evm/x/vm/statedb"

	"github.com/realiotech/realio-network/x/claim/types"
)

// erc20ABI is the standard ERC-20 surface, reused from x/erc20 rather than
// redeclared. Only balanceOf and transfer are called.
var erc20ABI = contracts.ERC20MinterBurnerDecimalsContract.ABI

// MigrateERC20 moves the old address's entire balance of one ERC-20
// contract to the new address, and reports whether there was anything to
// move. A zero balance is not an error -- the admin lists candidate
// contracts without necessarily knowing which ones the address actually
// holds.
//
// The transfer runs inside the EVM with oldAddr as the caller, so the token
// contract sees exactly what it would have seen had the holder sent the
// transaction themselves. That the holder's key is leaked, and the address
// blacklisted, does not get in the way: x/blacklist's EVM guard is an
// EVM *hook* (PostTxProcessing), and hooks fire only on the msg-server
// transaction path, never on a module's own CallEVM.
//
// Afterwards the old balance is read back and must be zero. That single
// check is the whole verification, and it is deliberately not an equality
// check against what moved: a token that takes a fee on transfer delivers
// less than it debited, which is fine when the goal is emptying the old
// address. It also subsumes checking transfer's return value -- a token
// that reports failure without reverting moves nothing, so the balance
// still stands there.
func (k Keeper) MigrateERC20(ctx sdk.Context, contract, oldAddr, newAddr common.Address) (bool, error) {
	if err := k.assertSweepable(ctx, contract); err != nil {
		return false, err
	}
	k.ensureCallerAccount(ctx, oldAddr.Bytes())

	balance, err := k.erc20BalanceOf(ctx, contract, oldAddr)
	if err != nil {
		return false, err
	}
	if balance.Sign() == 0 {
		return false, nil
	}

	stateDB := statedb.New(ctx, k.evmKeeper, statedb.NewEmptyTxConfig())
	if _, err := k.evmKeeper.CallEVM(
		ctx, stateDB, erc20ABI, oldAddr, contract, true, false, nil,
		"transfer", newAddr, balance,
	); err != nil {
		return false, err
	}

	remaining, err := k.erc20BalanceOf(ctx, contract, oldAddr)
	if err != nil {
		return false, err
	}
	if remaining.Sign() != 0 {
		return false, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"transfer left %s behind on %s; the contract did not move the full balance", remaining, oldAddr)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeClaimedERC20,
		sdk.NewAttribute(types.AttributeKeyContract, contract.Hex()),
		sdk.NewAttribute(types.AttributeKeyOldAddress, oldAddr.Hex()),
		sdk.NewAttribute(types.AttributeKeyNewAddress, newAddr.Hex()),
		sdk.NewAttribute(types.AttributeKeyAmount, balance.String()),
	))
	return true, nil
}

// assertSweepable rejects the two kinds of address that look like a token
// contract in a message but cannot be swept by calling into the EVM.
//
// Neither case would fail loudly on its own. An ERC-20 that x/erc20
// registered as the representation of a native coin is a precompile, not
// deployed bytecode: "transferring" it runs bank's SendCoins, where
// x/blacklist's send restriction blocks a blacklisted sender paying a user
// account -- surfacing as a bare "address is blacklisted" with no hint
// that the balance was a bank balance all along. And a call to an address
// with no code at all does not revert in the EVM; it succeeds and returns
// nothing, so the mistake would only show up later as an unpack failure
// somewhere less obvious.
func (k Keeper) assertSweepable(ctx sdk.Context, contract common.Address) error {
	if id := k.erc20Keeper.GetTokenPairID(ctx, contract.Hex()); len(id) > 0 {
		if pair, found := k.erc20Keeper.GetTokenPair(ctx, id); found && pair.IsNativeCoin() {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
				"%s is the ERC-20 representation of native denom %q; claim that with MsgClaimBalances instead",
				contract, pair.Denom)
		}
	}

	acc := k.evmKeeper.GetAccountWithoutBalance(ctx, contract)
	if acc == nil || !acc.HasCodeHash() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s has no contract code", contract)
	}
	return nil
}

// ensureCallerAccount gives addr an auth account if it has none, so that
// the EVM call made on its behalf can read a nonce. An address that only
// ever received tokens -- airdropped an ERC-20, say -- never signed
// anything, so x/auth has never heard of it, and CallEVM's nonce lookup
// fails before the contract is even reached. Creating the account is
// exactly what would have happened the first time that address sent a
// transaction of its own.
func (k Keeper) ensureCallerAccount(ctx sdk.Context, addr sdk.AccAddress) {
	if k.accountKeeper.GetAccount(ctx, addr) == nil {
		k.accountKeeper.SetAccount(ctx, k.accountKeeper.NewAccountWithAddress(ctx, addr))
	}
}

// erc20BalanceOf reads a balance with a non-committing EVM call. x/erc20's
// own BalanceOf is not reused because it reports every failure as a nil
// balance, which here would be indistinguishable from "holds nothing" --
// exactly the distinction a sweep has to get right.
func (k Keeper) erc20BalanceOf(ctx sdk.Context, contract, account common.Address) (*big.Int, error) {
	stateDB := statedb.New(ctx, k.evmKeeper, statedb.NewEmptyTxConfig())
	res, err := k.evmKeeper.CallEVM(ctx, stateDB, erc20ABI, erc20types.ModuleAddress, contract, false, false, nil, "balanceOf", account)
	if err != nil {
		return nil, err
	}

	unpacked, err := erc20ABI.Unpack("balanceOf", res.Ret)
	if err != nil {
		return nil, errorsmod.Wrapf(erc20types.ErrABIUnpack, "unpacking balanceOf for %s: %s", contract, err)
	}
	if len(unpacked) == 0 {
		return nil, errorsmod.Wrapf(erc20types.ErrABIUnpack, "balanceOf on %s returned nothing", contract)
	}
	balance, ok := unpacked[0].(*big.Int)
	if !ok {
		return nil, errorsmod.Wrapf(erc20types.ErrABIUnpack, "balanceOf on %s returned a non-integer", contract)
	}
	return balance, nil
}
