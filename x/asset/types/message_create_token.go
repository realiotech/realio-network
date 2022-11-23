package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateToken = "create_token"

var _ sdk.Msg = &MsgCreateToken{}

func NewMsgCreateToken(creator string, name string, symbol string, total string, decimals string, authorizationRequired bool) *MsgCreateToken {
	return &MsgCreateToken{
		Creator:               creator,
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
	return TypeMsgCreateToken
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
