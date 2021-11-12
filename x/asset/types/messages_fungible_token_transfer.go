package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSendFungibleTokenTransfer{}

func NewMsgSendFungibleTokenTransfer(
	creator string,
	port string,
	channelID string,
	timeoutTimestamp uint64,
	denom string,
	amount uint64,
	receiver string,
) *MsgSendFungibleTokenTransfer {
	return &MsgSendFungibleTokenTransfer{
		Creator:          creator,
		Port:             port,
		ChannelID:        channelID,
		TimeoutTimestamp: timeoutTimestamp,
		Denom:            denom,
		Amount:           amount,
		Receiver:         receiver,
	}
}

func (msg *MsgSendFungibleTokenTransfer) Route() string {
	return RouterKey
}

func (msg *MsgSendFungibleTokenTransfer) Type() string {
	return "SendFungibleTokenTransfer"
}

func (msg *MsgSendFungibleTokenTransfer) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSendFungibleTokenTransfer) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSendFungibleTokenTransfer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if msg.Port == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid packet port")
	}
	if msg.ChannelID == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid packet channel")
	}
	if msg.TimeoutTimestamp == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid packet timeout")
	}
	return nil
}
