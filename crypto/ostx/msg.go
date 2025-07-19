package ostx

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmapi "github.com/evmos/os/api/os/evm/v1"
	protov2 "google.golang.org/protobuf/proto"
)

// message type and route constants
const (
	// TypeMsgEthereumTx defines the type string of an Ethereum transaction
	TypeMsgEthereumTx = "ethereum_tx"
)

var MsgEthereumTxCustomGetSigner = txsigning.CustomGetSigner{
	MsgType: protov2.MessageName(&evmapi.MsgEthereumTx{}),
	Fn:      evmapi.GetSigners,
}

func (msg MsgEthereumTx) GetData() *codectypes.Any {
	return msg.Data
}

func (msg MsgEthereumTx) GetTxData() (TxData, error) {
	txData, err := UnpackTxData(msg.Data)
	if err != nil {
		return nil, err
	}
	return txData, nil
}

// AsTransaction creates an Ethereum Transaction type from the msg fields
func (msg MsgEthereumTx) AsTransaction() *ethtypes.Transaction {
	txData, err := UnpackTxData(msg.Data)
	if err != nil {
		return nil
	}

	return ethtypes.NewTx(txData.AsEthereumData())
}

// GetSender extracts the sender address from the signature values using the latest signer for the given chainID.
func (msg MsgEthereumTx) GetSender(chainID *big.Int) (common.Address, error) {
	signer := ethtypes.LatestSignerForChainID(chainID)
	from, err := signer.Sender(msg.AsTransaction())
	if err != nil {
		return common.Address{}, err
	}

	msg.From = from.Hex()
	return from, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgEthereumTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if msg.Data == nil {
		return nil
	}

	var txData TxData
	return unpacker.UnpackAny(msg.Data, &txData)
}

// UnpackTxData unpacks an Any into a TxData. It returns an error if the
// client state can't be unpacked into a TxData.
func UnpackTxData(anyTxData *codectypes.Any) (TxData, error) {
	if anyTxData == nil {
		return nil, errorsmod.Wrap(errortypes.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	txData, ok := anyTxData.GetCachedValue().(TxData)
	if !ok {
		return nil, errorsmod.Wrapf(errortypes.ErrUnpackAny, "cannot unpack Any into TxData %T", anyTxData)
	}

	return txData, nil
}
