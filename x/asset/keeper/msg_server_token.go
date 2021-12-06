package keeper

import (
	"context"
	"fmt"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (k msgServer) CreateToken(goCtx context.Context, msg *types.MsgCreateToken) (*types.MsgCreateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value already exists
	_, isFound := k.GetToken(
		ctx,
		msg.Index,
	)
	if isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "index already set")
	}

	var creatorAccAddress, err = sdk.AccAddressFromBech32(msg.Creator)

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid creator address")
	}

	var token = types.Token{
		Creator:               msg.Creator,
		Index:                 msg.Index,
		Name:                  msg.Name,
		Symbol:                msg.Symbol,
		Total:                 msg.Total,
		Decimals:              msg.Decimals,
		AuthorizationRequired: msg.AuthorizationRequired,
	}

	// mint coins for the current module
	var coin = sdk.Coins{{Denom: msg.Symbol, Amount: sdk.NewInt(msg.Total)}}

	k.SetToken(
		ctx,
		token,
	)

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, coin)
	if err != nil {
		panic(err)
	}

	k.bankKeeper.SetDenomMetaData(ctx, bank.Metadata{Base: msg.Symbol, Display: msg.Symbol, DenomUnits: []*bank.DenomUnit{{Denom: msg.Symbol, Exponent: 0}}})

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAccAddress, coin)
	if err != nil {
		panic(err)
	}

	return &types.MsgCreateTokenResponse{}, nil
}

func (k msgServer) UpdateToken(goCtx context.Context, msg *types.MsgUpdateToken) (*types.MsgUpdateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	existing, isFound := k.GetToken(
		ctx,
		msg.Index,
	)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	var token = types.Token{
		Creator:               existing.Creator,
		Index:                 existing.Index,
		Name:                  existing.Name,
		Symbol:                existing.Symbol,
		Total:                 existing.Total,
		Decimals:              existing.Decimals,
		AuthorizationRequired: msg.AuthorizationRequired,
	}

	k.SetToken(ctx, token)

	return &types.MsgUpdateTokenResponse{}, nil
}

func (k msgServer) AuthorizeAddress(goCtx context.Context, msg *types.MsgAuthorizeAddress) (*types.MsgAuthorizeAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	token, isFound := k.GetToken(ctx, msg.Index)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("index %v not set", msg.Index))
	}

	// Checks if the the msg sender is the same as the current owner
	if msg.Creator != token.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "caller not authorized")
	}

	if token.Authorized == nil {
		// initialize map on first write
		m := make(map[string]*types.TokenAuthorization)
		token.Authorized = m
	}
	var newAuthorization = types.TokenAuthorization{Address: msg.Address, TokenIndex: msg.Index, Authorized: true}

	token.Authorized[msg.Address] = &newAuthorization

	k.SetToken(ctx, token)

	return &types.MsgAuthorizeAddressResponse{}, nil
}

func (k msgServer) UnAuthorizeAddress(goCtx context.Context, msg *types.MsgUnAuthorizeAddress) (*types.MsgUnAuthorizeAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	token, isFound := k.GetToken(ctx, msg.Index)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("index %v not set", msg.Index))
	}

	// Checks if the the msg sender is the same as the current owner
	if msg.Creator != token.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "caller not authorized")
	}

	delete(token.Authorized, msg.Address)

	k.SetToken(ctx, token)

	return &types.MsgUnAuthorizeAddressResponse{}, nil
}
