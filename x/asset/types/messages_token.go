package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateToken{}

func NewMsgCreateToken(
	creator string,
	index string,
	name string,
	symbol string,
	total int64,
	decimals string,
	authorizationRequired bool,

) *MsgCreateToken {
	return &MsgCreateToken{
		Creator:               creator,
		Index:                 index,
		Name:                  name,
		Symbol:                symbol,
		Total:                 total,
		Decimals:              decimals,
		AuthorizationRequired: authorizationRequired,
	}
}

func (msg *MsgCreateToken) Route() string {
	return RouterKey
}

func (msg *MsgCreateToken) Type() string {
	return "CreateToken"
}

func (msg *MsgCreateToken) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateToken) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateToken{}

func NewMsgUpdateToken(
	creator string,
	index string,
	authorizationRequired bool,

) *MsgUpdateToken {
	return &MsgUpdateToken{
		Creator:               creator,
		Index:                 index,
		AuthorizationRequired: authorizationRequired,
	}
}

func (msg *MsgUpdateToken) Route() string {
	return RouterKey
}

func (msg *MsgUpdateToken) Type() string {
	return "UpdateToken"
}

func (msg *MsgUpdateToken) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateToken) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgAuthorizeAddress{}

func NewMsgAuthorizeAddress(creator string, index string, address string) *MsgAuthorizeAddress {
	return &MsgAuthorizeAddress{
		Creator: creator,
		Index:   index,
		Address: address,
	}
}

func (msg *MsgAuthorizeAddress) Route() string {
	return RouterKey
}

func (msg *MsgAuthorizeAddress) Type() string {
	return "AuthorizeAddress"
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

var _ sdk.Msg = &MsgUnAuthorizeAddress{}

func NewMsgUnAuthorizeAddress(creator string, index string, address string) *MsgUnAuthorizeAddress {
	return &MsgUnAuthorizeAddress{
		Creator: creator,
		Index:   index,
		Address: address,
	}
}

func (msg *MsgUnAuthorizeAddress) Route() string {
	return RouterKey
}

func (msg *MsgUnAuthorizeAddress) Type() string {
	return "UnAuthorizeAddress"
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