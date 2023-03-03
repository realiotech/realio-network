package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAuthorizeAddress = "authorize_address"

var _ sdk.Msg = &MsgAuthorizeAddress{}

func NewMsgAuthorizeAddress(manager string, symbol string, address string) *MsgAuthorizeAddress {
	return &MsgAuthorizeAddress{
		Manager: manager,
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
	manager, err := sdk.AccAddressFromBech32(msg.Manager)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{manager}
}

func (msg *MsgAuthorizeAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAuthorizeAddress) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Manager)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid manager address (%s)", err)
	}
	return nil
}
