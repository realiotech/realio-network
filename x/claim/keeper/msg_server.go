package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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
// holder has finished re-doing KYC) now maps to NewAddress, and migrates
// every active staking delegation OldAddress holds over to NewAddress in
// the same step -- see Keeper.MigrateDelegations. Only the module's
// configured admin may call this.
func (ms msgServer) LinkAddress(ctx context.Context, msg *types.MsgLinkAddress) (*types.MsgLinkAddressResponse, error) {
	if err := ms.checkAdmin(ctx, msg.Admin); err != nil {
		return nil, err
	}

	if err := ms.linkOneAddress(ctx, msg.OldAddress, msg.NewAddress); err != nil {
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
		if err := ms.linkOneAddress(ctx, link.OldAddress, link.NewAddress); err != nil {
			return nil, errorsmod.Wrapf(err, "link %d (old_address %q)", i, link.OldAddress)
		}
	}

	return &types.MsgLinkAddressesResponse{}, nil
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

// linkOneAddress validates one old_address/new_address pair, records the
// link, and migrates old_address's delegations to new_address. Shared by
// LinkAddress and LinkAddresses so both go through identical checks.
func (ms msgServer) linkOneAddress(ctx context.Context, oldAddrStr, newAddrStr string) error {
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

	if err := ms.SetLink(ctx, oldAddr, newAddr); err != nil {
		return err
	}

	if err := ms.MigrateDelegations(ctx, oldAddr, newAddr); err != nil {
		return errorsmod.Wrap(err, "migrating delegations")
	}

	return nil
}
