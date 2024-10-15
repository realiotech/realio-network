package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgBridgeIn{}, "bridge/BridgeIn", nil)
	cdc.RegisterConcrete(&MsgBridgeOut{}, "bridge/BridgeOut", nil)
	cdc.RegisterConcrete(&MsgRegisterNewCoins{}, "bridge/RegisterNewCoins", nil)
	cdc.RegisterConcrete(&MsgDeregisterCoins{}, "bridge/DeregisterCoins", nil)
	cdc.RegisterConcrete(&MsgUpdateEpochDuration{}, "bridge/UpdateEpochDuration", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "bridge/UpdateParams", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgBridgeIn{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgBridgeOut{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterNewCoins{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeregisterCoins{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateEpochDuration{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
