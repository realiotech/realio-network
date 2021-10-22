package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSendToAlgorand{}

func NewMsgSendToAlgorand(creator string, index string, denom string, algorandReceiver string, amount int64) *MsgSendToAlgorand {
	return &MsgSendToAlgorand{
		Creator:          creator,
		Index:            index,
		Denom:            denom,
		AlgorandReceiver: algorandReceiver,
		Amount:           amount,
	}
}

func (msg *MsgSendToAlgorand) Route() string {
	return RouterKey
}

func (msg *MsgSendToAlgorand) Type() string {
	return "SendToAlgorand"
}

func (msg *MsgSendToAlgorand) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSendToAlgorand) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSendToAlgorand) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
