package keeper

import (
	"context"
	"fmt"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"strings"

	realionetworktypes "github.com/realiotech/realio-network/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"cosmossdk.io/math"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (k msgServer) CreateToken(goCtx context.Context, msg *types.MsgCreateToken) (*types.MsgCreateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value already exists
	lowerCaseSymbol := strings.ToLower(msg.Symbol)
	lowerCaseName := strings.ToLower(msg.Name)
	baseDenom := fmt.Sprintf("a%s", lowerCaseSymbol)

	_, isFound := k.GetToken(
		ctx,
		msg.Symbol,
	)
	if isFound {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "symbol %s already set", msg.Symbol)
	}

	managerAccAddress, err := sdk.AccAddressFromBech32(msg.Manager)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid manager address")
	}

	token := types.Token{
		Name:                  lowerCaseName,
		Symbol:                lowerCaseSymbol,
		Total:                 msg.Total,
		Manager:               msg.Manager,
		AuthorizationRequired: msg.AuthorizationRequired,
	}

	if msg.AuthorizationRequired {
		if token.Authorized == nil {
			// initialize map on first write
			m := make(map[string]*types.TokenAuthorization)
			token.Authorized = m
		}

		newAuthorizationManager := types.TokenAuthorization{Address: msg.Manager, TokenSymbol: lowerCaseSymbol, Authorized: true}
		assetModuleAddress := k.ak.GetModuleAddress(types.ModuleName)
		newAuthorizationModule := types.TokenAuthorization{Address: assetModuleAddress.String(), TokenSymbol: lowerCaseSymbol, Authorized: true}

		token.Authorized[msg.Manager] = &newAuthorizationManager
		token.Authorized[assetModuleAddress.String()] = &newAuthorizationModule
	}

	k.SetToken(ctx, token)

	k.bankKeeper.SetDenomMetaData(ctx, bank.Metadata{
		Base: baseDenom, Symbol: lowerCaseSymbol, Name: lowerCaseName,
		DenomUnits: []*bank.DenomUnit{{Denom: lowerCaseSymbol, Exponent: 18}, {Denom: baseDenom, Exponent: 0}},
	})

	// mint coins for the current module
	// normalize into chains 10^18 denomination
	totalInt, _ := math.NewIntFromString(msg.Total)
	canonicalAmount := totalInt.Mul(realionetworktypes.PowerReduction)
	coin := sdk.Coins{{Denom: baseDenom, Amount: canonicalAmount}}

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, coin)

	if err != nil {
		panic(err)
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, managerAccAddress, coin)
	if err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenCreated,
			sdk.NewAttribute(sdk.AttributeKeyAmount, fmt.Sprint(msg.Total)),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
		),
	)

	return &types.MsgCreateTokenResponse{}, nil
}
