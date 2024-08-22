package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// implements sdk.Msg
func (m MsgMock) GetSigners() []sdk.AccAddress {
	return nil
}

// implements sdk.Msg
func (m MsgMock) ValidateBasic() error {
	return nil
}

// implements PrivilegeMsgI interface
func (m MsgMock) NeedPrivilege() string {
	return "test"
}
