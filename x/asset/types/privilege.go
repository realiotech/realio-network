package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
)

type PrivilegeMsgI interface {
	NeedPrivilege() string
}

type PrivilegeI interface {
	Name() string
	RegisterInterfaces() error
	MsgHandler() MsgHandler
	QueryHandler() QueryHandler
	CLI() *cobra.Command
}

type MsgHandler func(context context.Context, privMsg sdk.Msg) (proto.Message, error)

type QueryHandler func(context context.Context, privQuery sdk.Msg) (proto.Message, error)
