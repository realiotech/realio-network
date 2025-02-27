// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	erc20types "github.com/evmos/os/x/erc20/types"
	"github.com/realiotech/realio-network/x/erc20/types"
)

type msgServer struct {
	Erc20Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Erc20Keeper) types.MsgServer {
	return &msgServer{Erc20Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// RegisterERC20 implements the gRPC MsgServer interface. After a successful governance vote
// it updates creates the token pair for an ERC20 contract if the requested authority
// is the Cosmos SDK governance module account
func (ms msgServer) RegisterERC20Owner(goCtx context.Context, req *types.MsgRegisterERC20Owner) (*types.MsgRegisterERC20OwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Check if the conversion is globally enabled
	if !ms.Keeper.IsERC20Enabled(ctx) {
		return nil, erc20types.ErrERC20Disabled.Wrap("registration is currently disabled by governance")
	}

	if ms.authority != req.Authority {
		return nil, fmt.Errorf("invalid authority, expected: %s, get: %s", ms.authority, req.Authority)
	}

	// We only regist ones
	err := ms.Erc20Keeper.SetContractOwner(ctx, req.Erc20Address, req.Owner)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterERC20OwnerResponse{}, nil
}
