package types

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
)

type PrivilegeMsgI interface {
	NeedPrivilege() string
}

type PrivilegeI interface {
	Name() string
	RegisterInterfaces()
	MsgHandler() MsgHandler
	QueryHandler() QueryHandler
	CLI() *cobra.Command
}

type MsgHandler func(context context.Context, privMsg proto.Message) (proto.Message, error)

type QueryHandler func(context context.Context, privQuery proto.Message) (proto.Message, error)
