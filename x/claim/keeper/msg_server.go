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
	admin := ms.GetAdmin(ctx)
	if admin == "" || admin != msg.Admin {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "invalid admin; expected %s, got %s", admin, msg.Admin)
	}

	oldAddr, err := sdk.AccAddressFromBech32(msg.OldAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid old_address %q: %s", msg.OldAddress, err)
	}
	newAddr, err := sdk.AccAddressFromBech32(msg.NewAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid new_address %q: %s", msg.NewAddress, err)
	}
	if oldAddr.Equals(newAddr) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "old_address and new_address must differ")
	}
	if ms.IsLinked(ctx, oldAddr) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "old_address %s is already linked", msg.OldAddress)
	}
	if ms.IsLinked(ctx, newAddr) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "new_address %s was itself claimed as someone else's old_address; link to its new address instead", msg.NewAddress)
	}

	if err := ms.SetLink(ctx, oldAddr, msg.NewAddress); err != nil {
		return nil, err
	}

	if err := ms.MigrateDelegations(ctx, oldAddr, newAddr); err != nil {
		return nil, errorsmod.Wrap(err, "migrating delegations")
	}

	return &types.MsgLinkAddressResponse{}, nil
}
