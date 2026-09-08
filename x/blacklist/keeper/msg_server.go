package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/realiotech/realio-network/x/blacklist/types"
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

// UpdateBlacklist adds and/or removes addresses from the blacklist. Only the
// module's authority (x/gov by default) may call this — there is no other
// way to mutate the blacklist after genesis.
func (ms msgServer) UpdateBlacklist(ctx context.Context, msg *types.MsgUpdateBlacklist) (*types.MsgUpdateBlacklistResponse, error) {
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	for _, addr := range msg.AddAddresses {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid add_addresses entry %q: %s", addr, err)
		}
		if err := ms.SetBlacklisted(ctx, accAddr); err != nil {
			return nil, err
		}
	}

	for _, addr := range msg.RemoveAddresses {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid remove_addresses entry %q: %s", addr, err)
		}
		if err := ms.RemoveBlacklisted(ctx, accAddr); err != nil {
			return nil, err
		}
	}

	return &types.MsgUpdateBlacklistResponse{}, nil
}
