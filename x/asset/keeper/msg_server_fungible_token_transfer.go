package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	"github.com/realiotech/network/x/asset/types"
)

func (k msgServer) SendFungibleTokenTransfer(goCtx context.Context, msg *types.MsgSendFungibleTokenTransfer) (*types.MsgSendFungibleTokenTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: logic before transmitting the packet

	// Construct the packet
	var packet types.FungibleTokenTransferPacketData

	packet.Denom = msg.Denom
	packet.Amount = msg.Amount
	packet.Receiver = msg.Receiver
	packet.Sender = msg.Creator

	// Transmit the packet
	err := k.TransmitFungibleTokenTransferPacket(
		ctx,
		packet,
		msg.Port,
		msg.ChannelID,
		clienttypes.ZeroHeight(),
		msg.TimeoutTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendFungibleTokenTransferResponse{}, nil
}
