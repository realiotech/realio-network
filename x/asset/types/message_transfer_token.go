package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgTransferToken = "transfer_token"

var _ sdk.Msg = &MsgTransferToken{}

func NewMsgTransferToken(symbol string, from string, to string, amount string) *MsgTransferToken {
	return &MsgTransferToken{
		Symbol: symbol,
		From:   from,
		To:     to,
		Amount: amount,
	}
}

func (msg *MsgTransferToken) Route() string {
	return RouterKey
}

func (msg *MsgTransferToken) Type() string {
	return TypeMsgTransferToken
}

func (msg *MsgTransferToken) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgTransferToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgTransferToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.To); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}

	return nil
}
