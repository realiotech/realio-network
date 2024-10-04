package keeper

import (
	"context"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

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

	isFound := k.bankKeeper.HasSupply(ctx, baseDenom)
	if isFound {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "token with denom %s already exists", baseDenom)
	}

	managerAccAddress, err := sdk.AccAddressFromBech32(msg.Manager)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid manager address")
	}

	token := types.NewToken(lowerCaseName, lowerCaseSymbol, msg.Total, msg.Manager, msg.AuthorizationRequired)

	if msg.AuthorizationRequired {
		// create authorization for module account and manager
		assetModuleAddress := k.ak.GetModuleAddress(types.ModuleName)
		moduleAuthorization := types.NewAuthorization(assetModuleAddress)
		newAuthorizationManager := types.NewAuthorization(managerAccAddress)
		token.Authorized = append(token.Authorized, moduleAuthorization, newAuthorizationManager)
	}

	k.Token.Set(ctx, lowerCaseSymbol, token)
	err = k.Token.Set(goCtx, lowerCaseSymbol, token)
	if err != nil {
		return nil, types.ErrSetTokenUnable
	}

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
