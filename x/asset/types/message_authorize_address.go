package types

import (
	errorsmod "cosmossdk.io/errors"
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

func (msg *MsgAuthorizeAddress) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Manager)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid manager address (%s)", err)
	}
	return nil
}
