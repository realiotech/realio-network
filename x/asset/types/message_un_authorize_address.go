package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUnAuthorizeAddress = "un_authorize_address"

var _ sdk.Msg = &MsgUnAuthorizeAddress{}

func NewMsgUnAuthorizeAddress(manager string, symbol string, address string) *MsgUnAuthorizeAddress {
	return &MsgUnAuthorizeAddress{
		Manager: manager,
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
	manager, err := sdk.AccAddressFromBech32(msg.Manager)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{manager}
}

func (msg *MsgUnAuthorizeAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnAuthorizeAddress) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Manager); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid manager address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	return nil
}
