package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateToken = "create_token"

var _ sdk.Msg = &MsgCreateToken{}

func NewMsgCreateToken(manager string, name string, symbol string, total string, authorizationRequired bool) *MsgCreateToken {
	return &MsgCreateToken{
		Manager:               manager,
		Name:                  name,
		Symbol:                symbol,
		Total:                 total,
		AuthorizationRequired: authorizationRequired,
	}
}

func (msg *MsgCreateToken) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Manager)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid manager address (%s)", err)
	}
	return nil
}
