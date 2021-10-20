package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgTransferToken{}

func NewMsgTransferToken(creator string, index string, symbol string, from string, to string, amount int64) *MsgTransferToken {
	return &MsgTransferToken{
		Creator: creator,
		Index:   index,
		Symbol:  symbol,
		From:    from,
		To:      to,
		Amount:  amount,
	}
}

func (msg *MsgTransferToken) Route() string {
	return RouterKey
}

func (msg *MsgTransferToken) Type() string {
	return "TransferToken"
}

func (msg *MsgTransferToken) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgTransferToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgTransferToken) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
