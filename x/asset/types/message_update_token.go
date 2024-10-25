package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateToken = "update_token"

var _ sdk.Msg = &MsgUpdateToken{}

func NewMsgUpdateToken(manager string, symbol string, authorizationRequired bool) *MsgUpdateToken {
	return &MsgUpdateToken{
		Manager:               manager,
		Symbol:                symbol,
		AuthorizationRequired: authorizationRequired,
	}
}

func (msg *MsgUpdateToken) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Manager)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid manager address (%s)", err)
	}
	return nil
}
