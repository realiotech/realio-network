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

func (msg *MsgTransferToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.To); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}

	return nil
}
