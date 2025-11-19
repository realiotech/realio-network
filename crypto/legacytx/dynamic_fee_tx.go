// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package legacytx

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/evm/utils"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethutils "github.com/evmos/os/utils/eth"
)

func NewDynamicFeeTx(tx *ethtypes.Transaction) (*DynamicFeeTx, error) {
	txData := &DynamicFeeTx{
		Nonce:    tx.Nonce(),
		Data:     tx.Data(),
		GasLimit: tx.Gas(),
	}

	v, r, s := tx.RawSignatureValues()
	if to := tx.To(); to != nil {
		txData.To = to.Hex()
	}

	if tx.Value() != nil {
		amountInt, err := utils.SafeNewIntFromBigInt(tx.Value())
		if err != nil {
			return nil, err
		}
		txData.Amount = &amountInt
	}

	if tx.GasFeeCap() != nil {
		gasFeeCapInt, err := utils.SafeNewIntFromBigInt(tx.GasFeeCap())
		if err != nil {
			return nil, err
		}
		txData.GasFeeCap = &gasFeeCapInt
	}

	if tx.GasTipCap() != nil {
		gasTipCapInt, err := utils.SafeNewIntFromBigInt(tx.GasTipCap())
		if err != nil {
			return nil, err
		}
		txData.GasTipCap = &gasTipCapInt
	}

	if tx.AccessList() != nil {
		al := tx.AccessList()
		txData.Accesses = NewAccessList(&al)
	}

	txData.SetSignatureValues(tx.ChainId(), v, r, s)
	return txData, nil
}

// TxType returns the tx type
func (tx *DynamicFeeTx) TxType() uint8 {
	return ethtypes.DynamicFeeTxType
}

// Copy returns an instance with the same field values
func (tx *DynamicFeeTx) Copy() TxData {
	return &DynamicFeeTx{
		ChainID:   tx.ChainID,
		Nonce:     tx.Nonce,
		GasTipCap: tx.GasTipCap,
		GasFeeCap: tx.GasFeeCap,
		GasLimit:  tx.GasLimit,
		To:        tx.To,
		Amount:    tx.Amount,
		Data:      common.CopyBytes(tx.Data),
		Accesses:  tx.Accesses,
		V:         common.CopyBytes(tx.V),
		R:         common.CopyBytes(tx.R),
		S:         common.CopyBytes(tx.S),
	}
}

// GetChainID returns the chain id field from the DynamicFeeTx
func (tx *DynamicFeeTx) GetChainID() *big.Int {
	if tx.ChainID == nil {
		return nil
	}

	return tx.ChainID.BigInt()
}

// GetAccessList returns the AccessList field.
func (tx *DynamicFeeTx) GetAccessList() ethtypes.AccessList {
	if tx.Accesses == nil {
		return nil
	}
	return *tx.Accesses.ToEthAccessList()
}

// GetData returns the a copy of the input data bytes.
func (tx *DynamicFeeTx) GetData() []byte {
	return common.CopyBytes(tx.Data)
}

// GetGas returns the gas limit.
func (tx *DynamicFeeTx) GetGas() uint64 {
	return tx.GasLimit
}

// GetGasPrice returns the gas fee cap field.
func (tx *DynamicFeeTx) GetGasPrice() *big.Int {
	return tx.GetGasFeeCap()
}

// GetGasTipCap returns the gas tip cap field.
func (tx *DynamicFeeTx) GetGasTipCap() *big.Int {
	if tx.GasTipCap == nil {
		return nil
	}
	return tx.GasTipCap.BigInt()
}

// GetGasFeeCap returns the gas fee cap field.
func (tx *DynamicFeeTx) GetGasFeeCap() *big.Int {
	if tx.GasFeeCap == nil {
		return nil
	}
	return tx.GasFeeCap.BigInt()
}

// GetValue returns the tx amount.
func (tx *DynamicFeeTx) GetValue() *big.Int {
	if tx.Amount == nil {
		return nil
	}

	return tx.Amount.BigInt()
}

// GetNonce returns the account sequence for the transaction.
func (tx *DynamicFeeTx) GetNonce() uint64 { return tx.Nonce }

// GetTo returns the pointer to the recipient address.
func (tx *DynamicFeeTx) GetTo() *common.Address {
	if tx.To == "" {
		return nil
	}
	to := common.HexToAddress(tx.To)
	return &to
}

// AsEthereumData returns an DynamicFeeTx transaction tx from the proto-formatted
// TxData defined on the Cosmos EVM.
func (tx *DynamicFeeTx) AsEthereumData() ethtypes.TxData {
	v, r, s := tx.GetRawSignatureValues()
	return &ethtypes.DynamicFeeTx{
		ChainID:    tx.GetChainID(),
		Nonce:      tx.GetNonce(),
		GasTipCap:  tx.GetGasTipCap(),
		GasFeeCap:  tx.GetGasFeeCap(),
		Gas:        tx.GetGas(),
		To:         tx.GetTo(),
		Value:      tx.GetValue(),
		Data:       tx.GetData(),
		AccessList: tx.GetAccessList(),
		V:          v,
		R:          r,
		S:          s,
	}
}

// GetRawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (tx *DynamicFeeTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return ethutils.RawSignatureValues(tx.V, tx.R, tx.S)
}

// SetSignatureValues sets the signature values to the transaction.
func (tx *DynamicFeeTx) SetSignatureValues(chainID, v, r, s *big.Int) {
	if v != nil {
		tx.V = v.Bytes()
	}
	if r != nil {
		tx.R = r.Bytes()
	}
	if s != nil {
		tx.S = s.Bytes()
	}
	if chainID != nil {
		chainIDInt := sdkmath.NewIntFromBigInt(chainID)
		tx.ChainID = &chainIDInt
	}
}

// Validate performs a stateless validation of the tx fields.
func (tx DynamicFeeTx) Validate() error {
	return nil
}

// Fee returns gasprice * gaslimit.
func (tx DynamicFeeTx) Fee() *big.Int {
	return fee(tx.GetGasFeeCap(), tx.GasLimit)
}

// Cost returns amount + gasprice * gaslimit.
func (tx DynamicFeeTx) Cost() *big.Int {
	return cost(tx.Fee(), tx.GetValue())
}

// EffectiveGasPrice returns the effective gas price
func (tx *DynamicFeeTx) EffectiveGasPrice(baseFee *big.Int) *big.Int {
	return baseFee
}

// EffectiveFee returns effective_gasprice * gaslimit.
func (tx DynamicFeeTx) EffectiveFee(baseFee *big.Int) *big.Int {
	return fee(tx.EffectiveGasPrice(baseFee), tx.GasLimit)
}

// EffectiveCost returns amount + effective_gasprice * gaslimit.
func (tx DynamicFeeTx) EffectiveCost(baseFee *big.Int) *big.Int {
	return cost(tx.EffectiveFee(baseFee), tx.GetValue())
}
