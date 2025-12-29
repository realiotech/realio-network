package app

import (
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/simapp/params"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	evmcryptocodec "github.com/cosmos/evm/crypto/codec"
	evmaddress "github.com/cosmos/evm/encoding/address"
	enccodec "github.com/cosmos/evm/encoding/codec"
	"github.com/cosmos/evm/ethereum/eip712"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/cosmos/evm/x/vm/types/legacy"
	ethcryptocodec "github.com/realiotech/realio-network/crypto/codec"
)

// legacyFallbackTxConfig wraps a TxConfig to provide legacy transaction decoding fallback
type legacyFallbackTxConfig struct {
	client.TxConfig
}

// TxDecoder returns a decoder that falls back to legacy decoding for old EVM transactions
func (c legacyFallbackTxConfig) TxDecoder() sdk.TxDecoder {
	primaryDecoder := c.TxConfig.TxDecoder()
	return func(txBytes []byte) (sdk.Tx, error) {
		// Try primary decoder first
		tx, err := primaryDecoder(txBytes)
		if err == nil {
			return tx, nil
		}

		// Check if this looks like a legacy EVM tx error
		if !isLegacyTxError(err) {
			return nil, err
		}

		// Try legacy decoder
		legacyTx, legacyErr := legacy.DecodeTx(txBytes)
		if legacyErr != nil {
			// Return original error if legacy also fails
			return nil, err
		}

		return legacyTx, nil
	}
}

// isLegacyTxError checks if an error indicates a legacy transaction format issue
func isLegacyTxError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "errUnknownField") &&
		strings.Contains(errStr, "MsgEthereumTx")
}

// MakeEncodingConfig creates the EncodingConfig for realio network
func MakeEncodingConfig(evmChainID uint64) params.EncodingConfig {
	legacyAmino := codec.NewLegacyAmino()
	signingOptions := signing.Options{
		AddressCodec:          evmaddress.NewEvmCodec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: evmaddress.NewEvmCodec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
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
	enccodec.RegisterInterfaces(interfaceRegistry)
	evmcryptocodec.RegisterCrypto(legacyAmino)
	evmcryptocodec.RegisterInterfaces(interfaceRegistry)
	ethcryptocodec.RegisterCrypto(legacyAmino)
	ethcryptocodec.RegisterInterfaces(interfaceRegistry)

	// This is needed for the EIP712 txs because currently is using
	// the deprecated method legacytx.StdSignBytes
	legacytx.RegressionTestingAminoCodec = legacyAmino
	eip712.SetEncodingConfig(legacyAmino, interfaceRegistry, evmChainID)

	ModuleBasics.RegisterLegacyAminoCodec(legacyAmino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	// Wrap TxConfig with legacy fallback for decoding old EVM transactions
	wrappedTxConfig := legacyFallbackTxConfig{TxConfig: txConfig}

	return params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          wrappedTxConfig,
		Amino:             legacyAmino,
	}
}
