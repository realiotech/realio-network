package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/suite"

	"github.com/realiotech/realio-network/testutil"
)

type MessageTestSuite struct {
	suite.Suite
}

func TestMessageAuthorizeTestSuite(t *testing.T) {
	suite.Run(t, new(MessageTestSuite))
}

func (suite *MessageTestSuite) TestMsgAuthorizeAddress_ValidateBasic() {
	tests := []struct {
		name string
		msg  MsgAuthorizeAddress
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgAuthorizeAddress{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgAuthorizeAddress{
				Creator: testutil.GenAddress().String(),
			},
		},
	}
	for _, tt := range tests {
		err := tt.msg.ValidateBasic()
		if tt.err != nil {
			suite.Require().ErrorIs(err, tt.err)
			return
		}
		suite.Require().NoError(err)
	}
}

func (suite *MessageTestSuite) TestMsgCreateToken_ValidateBasic() {
	tests := []struct {
		name string
		msg  MsgCreateToken
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateToken{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateToken{
				Creator: testutil.GenAddress().String(),
			},
		},
	}
	for _, tt := range tests {
		err := tt.msg.ValidateBasic()
		if tt.err != nil {
			suite.Require().ErrorIs(err, tt.err)
			return
		}
		suite.Require().NoError(err)
	}
}

func (suite *MessageTestSuite) TestMsgTransferToken_ValidateBasic() {
	tests := []struct {
		name string
		msg  MsgTransferToken
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgTransferToken{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgTransferToken{
				Creator: testutil.GenAddress().String(),
			},
		},
	}
	for _, tt := range tests {
		err := tt.msg.ValidateBasic()
		if tt.err != nil {
			suite.Require().ErrorIs(err, tt.err)
			return
		}
		suite.Require().NoError(err)
	}
}

func (suite *MessageTestSuite) TestMsgUnAuthorizeAddress_ValidateBasic() {
	tests := []struct {
		name string
		msg  MsgUnAuthorizeAddress
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUnAuthorizeAddress{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgUnAuthorizeAddress{
				Creator: testutil.GenAddress().String(),
			},
		},
	}
	for _, tt := range tests {
		err := tt.msg.ValidateBasic()
		if tt.err != nil {
			suite.Require().ErrorIs(err, tt.err)
			return
		}
		suite.Require().NoError(err)
	}
}

func (suite *MessageTestSuite) TestMsgUpdateToken_ValidateBasic() {
	tests := []struct {
		name string
		msg  MsgUpdateToken
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUpdateToken{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgUpdateToken{
				Creator: testutil.GenAddress().String(),
			},
		},
	}
	for _, tt := range tests {
		err := tt.msg.ValidateBasic()
		if tt.err != nil {
			suite.Require().ErrorIs(err, tt.err)
			return
		}
		suite.Require().NoError(err)
	}
}
