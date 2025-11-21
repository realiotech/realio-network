package bank

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	cmn "github.com/cosmos/evm/precompiles/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Balance contains the amount for a corresponding ERC-20 contract address.
type Balance struct {
	Denom  string
	Amount *big.Int
}

type Output struct {
	Addr   common.Address `abi:"addr"`
	Amount string     `abi:"amount"`
}

// ParseBalancesArgs parses the call arguments for the bank Balances query.
func ParseBalancesArgs(args []interface{}) (sdk.AccAddress, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	account, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "account", common.Address{}, args[0])
	}

	return account.Bytes(), nil
}

// ParseSupplyOfArgs parses the call arguments for the bank SupplyOf query.
func ParseSupplyOfArgs(args []interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	denom, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf(cmn.ErrInvalidType, "erc20Address", common.Address{}, args[0])
	}

	return denom, nil
}

func ParseSendArgs(args []interface{}) (common.Address, sdk.Coins, error) {
	if len(args) != 2 {
		return common.Address{}, sdk.Coins{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	receiver, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, sdk.Coins{}, fmt.Errorf(cmn.ErrInvalidType, "receriver", common.Address{}, args[0])
	}

	coinStr, ok := args[1].(string)
	if !ok {
		return common.Address{}, sdk.Coins{}, fmt.Errorf(cmn.ErrInvalidType, "denom", common.Address{}, args[1])
	}

	coins, err := sdk.ParseCoinsNormalized(coinStr)
	if err != nil {
		return common.Address{}, sdk.Coins{}, err
	}

	return receiver, coins, nil
}

func ParseMultiSendArgs(args []interface{}) (sdk.Coins, []Output, error) {
	if len(args) != 2 {
		return sdk.Coins{}, nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	coinStr, ok := args[0].(string)
	if !ok {
		return sdk.Coins{}, nil, fmt.Errorf(cmn.ErrInvalidType, "receiver", common.Address{}, args[0])
	}
	coins, err := sdk.ParseCoinsNormalized(coinStr)
	if err != nil {
		return sdk.Coins{}, nil, err
	}

	outputs, ok := args[1].([]Output)
	if !ok {
		return sdk.Coins{}, nil, fmt.Errorf(cmn.ErrInvalidType, "output", []Output{}, args[1])
	}

	return coins, outputs, nil
}
