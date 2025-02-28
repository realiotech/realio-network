package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global erc20 module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to modules/erc20 and
	// defined at the application level.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	// AminoCdc is a amino codec created to support amino JSON compatible msgs.
	AminoCdc = codec.NewAminoCodec(amino) //nolint:staticcheck
)

const (
	registerERC20Owner = "realionetwork/erc20/MsgRegisterERC20Owner"
)

// // NOTE: This is required for the GetSignBytes function
func init() {
	RegisterCodec(amino)
	amino.Seal()
}

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterERC20Owner{},
	)
	registry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&RegisterERC20OwnerProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/erc20 interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization and EIP-712 compatibility.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterERC20Owner{}, registerERC20Owner, nil)
}
