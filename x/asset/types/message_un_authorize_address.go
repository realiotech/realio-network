package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUnAuthorizeAddress = "un_authorize_address"

var _ sdk.Msg = &MsgUnAuthorizeAddress{}

func NewMsgUnAuthorizeAddress(creator string, symbol string, address string) *MsgUnAuthorizeAddress {
	return &MsgUnAuthorizeAddress{
		Creator: creator,
		Symbol:  symbol,
		Address: address,
	}
}

func (msg *MsgUnAuthorizeAddress) Route() string {
	return RouterKey
}

func (msg *MsgUnAuthorizeAddress) Type() string {
	return TypeMsgUnAuthorizeAddress
}

func (msg *MsgUnAuthorizeAddress) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUnAuthorizeAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnAuthorizeAddress) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
