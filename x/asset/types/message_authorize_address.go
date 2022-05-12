package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAuthorizeAddress = "authorize_address"

var _ sdk.Msg = &MsgAuthorizeAddress{}

func NewMsgAuthorizeAddress(creator string, symbol string, address string) *MsgAuthorizeAddress {
	return &MsgAuthorizeAddress{
		Creator: creator,
		Symbol:  symbol,
		Address: address,
	}
}

func (msg *MsgAuthorizeAddress) Route() string {
	return RouterKey
}

func (msg *MsgAuthorizeAddress) Type() string {
	return TypeMsgAuthorizeAddress
}

func (msg *MsgAuthorizeAddress) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAuthorizeAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAuthorizeAddress) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
