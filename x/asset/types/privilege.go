package types

import (
	"context"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	// "github.com/cosmos/cosmos-sdk/store/types"
)

type PrivilegeMsgI interface {
	NeedPrivilege() string
}

type PrivilegeI interface {
	Name() string
	RegisterInterfaces(registry cdctypes.InterfaceRegistry)
	MsgHandler() MsgHandler
	QueryHandler() QueryHandler
}

type MsgHandler func(context context.Context, msg proto.Message, tokenID string, privAcc sdk.AccAddress) (proto.Message, error)

type QueryHandler func(context context.Context, privQuery proto.Message, tokenID string) (proto.Message, error)
