package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*PrivilegeI)(nil), nil)
	cdc.RegisterConcrete(&MsgCreateToken{}, "asset/CreateToken", nil)
	cdc.RegisterConcrete(&MsgUpdateToken{}, "asset/UpdateToken", nil)
	cdc.RegisterConcrete(&MsgAllocateToken{}, "asset/AllocateToken", nil)
	cdc.RegisterConcrete(&MsgAssignPrivilege{}, "asset/AssignPrivilege", nil)
	cdc.RegisterConcrete(&MsgUnassignPrivilege{}, "asset/UnassignPrivilege", nil)
	cdc.RegisterConcrete(&MsgDisablePrivilege{}, "asset/DisablePrivilege", nil)
	cdc.RegisterConcrete(&MsgExecutePrivilege{}, "asset/ExecutePrivilege", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	var privilege *PrivilegeI
	registry.RegisterInterface(
		"realionetwork.asset.v1.PrivilegeI",
		privilege,
	)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateToken{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateToken{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAllocateToken{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAssignPrivilege{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUnassignPrivilege{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDisablePrivilege{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgExecutePrivilege{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
