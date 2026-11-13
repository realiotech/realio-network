package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/realiotech/realio-network/x/claim/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// LinkAddress records that OldAddress (a leaked/compromised address whose
// holder has finished re-doing KYC) was recovered to NewAddress. It moves
// nothing: the link is what every Claim message afterwards reads to find
// where assets should go.
func (ms msgServer) LinkAddress(ctx context.Context, msg *types.MsgLinkAddress) (*types.MsgLinkAddressResponse, error) {
	if err := ms.checkAdmin(ctx, msg.Admin); err != nil {
		return nil, err
	}

	if err := ms.recordLink(ctx, msg.OldAddress, msg.NewAddress); err != nil {
		return nil, err
	}

	return &types.MsgLinkAddressResponse{}, nil
}

// LinkAddresses is the batch form of LinkAddress: it applies every
// old_address -> new_address pair in msg.Links, in order, under a single
// admin check. Like any other message, a failure partway through fails the
// whole transaction -- either every pair in the batch is applied, or none
// are; there is no partial-batch state to reconcile.
func (ms msgServer) LinkAddresses(ctx context.Context, msg *types.MsgLinkAddresses) (*types.MsgLinkAddressesResponse, error) {
	if err := ms.checkAdmin(ctx, msg.Admin); err != nil {
		return nil, err
	}
	if len(msg.Links) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "links must not be empty")
	}

	for i, link := range msg.Links {
		if err := ms.recordLink(ctx, link.OldAddress, link.NewAddress); err != nil {
			return nil, errorsmod.Wrapf(err, "link %d (old_address %q)", i, link.OldAddress)
		}
	}

	return &types.MsgLinkAddressesResponse{}, nil
}

// ClaimDelegations re-keys every active delegation the linked old address
// holds over to its new address. See Keeper.MigrateDelegations for what
// moves and what is deliberately left behind.
func (ms msgServer) ClaimDelegations(ctx context.Context, msg *types.MsgClaimDelegations) (*types.MsgClaimDelegationsResponse, error) {
	oldAddr, newAddr, err := ms.authorizeClaim(ctx, msg.Admin, msg.OldAddress)
	if err != nil {
		return nil, err
	}

	if err := ms.MigrateDelegations(ctx, oldAddr, newAddr); err != nil {
		return nil, errorsmod.Wrap(err, "migrating delegations")
	}

	return &types.MsgClaimDelegationsResponse{NewAddress: newAddr.String()}, nil
}

// MigrateBankBalance recovers bank coins for the linked old address: the
// whole spendable balance, plus whatever the admin pays out of the
// attacker on top.
//
// One message because it is one operation per holder, but two keeper calls
// and two events, so the record of what came back from where survives.
// See Keeper.MigrateBankBalance and Keeper.ClaimFromAttacker.
func (ms msgServer) MigrateBankBalance(ctx context.Context, msg *types.MsgMigrateBankBalance) (*types.MsgMigrateBankBalanceResponse, error) {
	oldAddr, newAddr, err := ms.authorizeClaim(ctx, msg.Admin, msg.OldAddress)
	if err != nil {
		return nil, err
	}
	if err := msg.FromAttacker.Validate(); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "invalid from_attacker: %s", err)
	}

	// An address with nothing spendable is swept successfully, for zero --
	// not an error, the same way ClaimTokens skips a contract the holder
	// holds none of. The admin does not always know which addresses still
	// hold something, and erroring here would take down the payout for an
	// address the attacker drained to empty, which is precisely the
	// address whose payout matters most.
	swept, err := ms.Keeper.MigrateBankBalance(ctx, oldAddr, newAddr)
	if err != nil {
		return nil, err
	}

	if !msg.FromAttacker.IsZero() {
		if err := ms.Keeper.ClaimFromAttacker(ctx, oldAddr, newAddr, msg.FromAttacker); err != nil {
			return nil, err
		}
	}

	// A message that finds nothing to move is not an error either. A
	// re-sweep pass over every recovered address is a normal thing to run
	// -- most addresses will have nothing new, and erroring on those would
	// make the pass unusable. The response reports what moved, so a no-op
	// is visible without being fatal, matching how ClaimTokens reports a
	// sweep that found every listed contract empty.
	var remaining sdk.Coins
	if attacker, ok := ms.GetAttacker(ctx); ok {
		remaining = ms.SpendableCoinsOf(ctx, attacker)
	}
	return &types.MsgMigrateBankBalanceResponse{
		NewAddress:          newAddr.String(),
		Swept:               swept,
		ClaimedFromAttacker: msg.FromAttacker,
		AttackerRemaining:   remaining,
	}, nil
}

// ClaimTokens sweeps ERC-20 balances held by the linked old address over
// to its new address, by running the transfers inside the EVM with the old
// address as the caller.
//
// All or nothing within this message: one bad contract fails the whole
// sweep, so there is never a half-swept address to reconcile against the
// list that was submitted.
func (ms msgServer) ClaimTokens(ctx context.Context, msg *types.MsgClaimTokens) (*types.MsgClaimTokensResponse, error) {
	oldAddr, newAddr, err := ms.authorizeClaim(ctx, msg.Admin, msg.OldAddress)
	if err != nil {
		return nil, err
	}
	if len(msg.Erc20Contracts) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "erc20_contracts must not be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	oldEVM := common.BytesToAddress(oldAddr.Bytes())
	newEVM := common.BytesToAddress(newAddr.Bytes())

	var erc20Swept uint64
	for i, contract := range msg.Erc20Contracts {
		if !common.IsHexAddress(contract) {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress,
				"erc20_contracts[%d]: %q is not a hex address", i, contract)
		}
		moved, err := ms.MigrateERC20(sdkCtx, common.HexToAddress(contract), oldEVM, newEVM)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "erc20_contracts[%d] (%s)", i, contract)
		}
		if moved {
			erc20Swept++
		}
	}

	return &types.MsgClaimTokensResponse{
		NewAddress:          newAddr.String(),
		Erc20ContractsSwept: erc20Swept,
	}, nil
}

// authorizeClaim is the gate every Claim message passes through: the
// sender must be the configured admin, and old_address must already be
// linked. It returns the pair the claim moves assets between, so that no
// Claim message has a field for the destination -- the only address an
// admin names is the one being emptied, never where it empties to.
func (ms msgServer) authorizeClaim(ctx context.Context, msgAdmin, oldAddress string) (oldAddr, newAddr sdk.AccAddress, err error) {
	if err := ms.checkAdmin(ctx, msgAdmin); err != nil {
		return nil, nil, err
	}

	oldAddr, err = sdk.AccAddressFromBech32(oldAddress)
	if err != nil {
		return nil, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid old_address %q: %s", oldAddress, err)
	}

	newAddr, ok := ms.GetLink(ctx, oldAddr)
	if !ok {
		return nil, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"old_address %s is not linked; submit MsgLinkAddress first", oldAddress)
	}
	return oldAddr, newAddr, nil
}

// checkAdmin verifies msgAdmin (the bech32 admin field off an incoming
// message) matches the module's configured admin.
func (ms msgServer) checkAdmin(ctx context.Context, msgAdmin string) error {
	admin, ok := ms.GetAdmin(ctx)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid admin; no admin configured")
	}

	msgAdminAddr, err := sdk.AccAddressFromBech32(msgAdmin)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid admin address %q: %s", msgAdmin, err)
	}
	if !admin.Equals(msgAdminAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "invalid admin; expected %s, got %s", admin, msgAdmin)
	}
	return nil
}

// recordLink validates one old_address/new_address pair and records it.
// Shared by LinkAddress and LinkAddresses so both go through identical
// checks.
func (ms msgServer) recordLink(ctx context.Context, oldAddrStr, newAddrStr string) error {
	oldAddr, err := sdk.AccAddressFromBech32(oldAddrStr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid old_address %q: %s", oldAddrStr, err)
	}
	newAddr, err := sdk.AccAddressFromBech32(newAddrStr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid new_address %q: %s", newAddrStr, err)
	}
	if oldAddr.Equals(newAddr) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "old_address and new_address must differ")
	}
	if ms.IsLinked(ctx, oldAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "old_address %s is already linked", oldAddrStr)
	}
	if ms.IsLinked(ctx, newAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "new_address %s was itself claimed as someone else's old_address; link to its new address instead", newAddrStr)
	}

	return ms.SetLink(ctx, oldAddr, newAddr)
}
