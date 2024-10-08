package app

import (
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/simapp/params"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	evmcryptocodec "github.com/evmos/os/crypto/codec"
	"github.com/evmos/os/ethereum/eip712"
	evmostypes "github.com/evmos/os/types"
	erc20types "github.com/evmos/os/x/erc20/types"
	evmtypes "github.com/evmos/os/x/evm/types"
)

// MakeEncodingConfig creates the EncodingConfig for realio network
func MakeEncodingConfig() params.EncodingConfig {
	legacyAmino := codec.NewLegacyAmino()
	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
		CustomGetSigners: map[protoreflect.FullName]signing.GetSignersFunc{
			evmtypes.MsgEthereumTxCustomGetSigner.MsgType:     evmtypes.MsgEthereumTxCustomGetSigner.Fn,
			erc20types.MsgConvertERC20CustomGetSigner.MsgType: erc20types.MsgConvertERC20CustomGetSigner.Fn,
		},
	}

	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	codec := codec.NewProtoCodec(interfaceRegistry)

	txConfig := tx.NewTxConfig(codec, tx.DefaultSignModes)

	std.RegisterInterfaces(interfaceRegistry)
	evmostypes.RegisterInterfaces(interfaceRegistry)
	evmcryptocodec.RegisterCrypto(legacyAmino)
	evmcryptocodec.RegisterInterfaces(interfaceRegistry)

	// This is needed for the EIP712 txs because currently is using
	// the deprecated method legacytx.StdSignBytes
	legacytx.RegressionTestingAminoCodec = legacyAmino
	eip712.SetEncodingConfig(legacyAmino, interfaceRegistry)

	ModuleBasics.RegisterLegacyAminoCodec(legacyAmino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	return params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          txConfig,
		Amino:             legacyAmino,
	}
}
