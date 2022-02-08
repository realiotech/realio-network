package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateRstStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateRstStake
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateRstStake{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateRstStake{
				Creator: sample.AccAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgUpdateRstStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateRstStake
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUpdateRstStake{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgUpdateRstStake{
				Creator: sample.AccAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgDeleteRstStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeleteRstStake
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDeleteRstStake{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgDeleteRstStake{
				Creator: sample.AccAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
