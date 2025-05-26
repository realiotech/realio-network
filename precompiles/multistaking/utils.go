// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package multistaking

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	erc20types "github.com/cosmos/evm/x/erc20/types"
)

// convertERC20ToSDKCoin converts ERC20 tokens to SDK coins for staking
func (p Precompile) convertERC20ToSDKCoin(
	ctx sdk.Context,
	caller common.Address,
	erc20Token common.Address,
	amount *big.Int,
) (sdk.Coin, error) {
	denom, err := p.getSDKDenomForERC20(ctx, erc20Token)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Create conversion message
	convertMsg := &erc20types.MsgConvertERC20{
		ContractAddress: erc20Token.Hex(),
		Amount:          math.NewIntFromBigInt(amount),
		Receiver:        sdk.AccAddress(caller.Bytes()).String(),
		Sender:          caller.Hex(),
	}

	// Execute conversion
	_, err = p.erc20Keeper.ConvertERC20(ctx, convertMsg)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("ERC20 conversion failed: %v", err)
	}

	return sdk.NewCoin(denom, math.NewIntFromBigInt(amount)), nil
}

// convertSDKCoinToERC20 converts SDK coins back to ERC20 tokens
func (p Precompile) convertSDKCoinToERC20(
	ctx sdk.Context,
	caller sdk.AccAddress,
	coin sdk.Coin,
) error {
	// Create conversion message
	convertMsg := &erc20types.MsgConvertCoin{
		Coin:     coin,
		Receiver: caller.String(),
		Sender:   caller.String(),
	}

	// Execute conversion
	_, err := p.erc20Keeper.ConvertCoin(ctx, convertMsg)
	if err != nil {
		return fmt.Errorf("SDK coin to ERC20 conversion failed: %v", err)
	}

	return nil
}

// getSDKDenomForERC20 gets the SDK denomination for an ERC20 token
func (p Precompile) getSDKDenomForERC20(ctx sdk.Context, erc20Token common.Address) (string, error) {
	tokenId := p.erc20Keeper.GetTokenPairID(ctx, erc20Token.Hex())
	// Get the token pair for this ERC20
	tokenPair, found := p.erc20Keeper.GetTokenPair(ctx, tokenId)
	if !found {
		return "", fmt.Errorf("token pair not found: %s", erc20Token.Hex())
	}
	return tokenPair.Denom, nil
}