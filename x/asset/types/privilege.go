package types

import (
	"context"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
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
	CLI() *cobra.Command
}

type MsgHandler func(context context.Context, privStore storetypes.KVStore, msg proto.Message, tokenID string, privAcc sdk.AccAddress) (proto.Message, error)

type QueryHandler func(context context.Context, privQuery proto.Message, tokenID string) (proto.Message, error)
