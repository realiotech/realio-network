package keeper

import (
	"context"
	"fmt"
	"slices"
	"strings"

	errorsmod "cosmossdk.io/errors"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/realiotech/realio-network/x/asset/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) CreateToken(goCtx context.Context, msg *types.MsgCreateToken) (*types.MsgCreateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Check if the value already exists
	lowerCaseSymbol := strings.ToLower(msg.Symbol)
	lowerCaseName := strings.ToLower(msg.Name)
	tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, msg.Creator, lowerCaseSymbol)
	isFound := k.bankKeeper.HasSupply(ctx, tokenId)
	if isFound {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "token with denom %s already exists", tokenId)
	}

	token := types.NewToken(tokenId, lowerCaseName, lowerCaseSymbol, msg.Decimal, msg.Description)
	k.SetToken(ctx, tokenId, token)
	k.bankKeeper.SetDenomMetaData(ctx, bank.Metadata{
		Base: tokenId, Symbol: lowerCaseSymbol, Name: lowerCaseName,
		DenomUnits: []*bank.DenomUnit{{Denom: tokenId, Exponent: msg.Decimal}},
	})

	tokenManage := types.NewTokenManagement(msg.Manager, msg.AddNewPrivilege, msg.ExcludedPrivileges, msg.EnabledPrivileges)
	k.SetTokenManagement(ctx, tokenId, tokenManage)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenCreated,
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
		),
	)

	return &types.MsgCreateTokenResponse{}, nil
}

func (k msgServer) UpdateToken(goCtx context.Context, msg *types.MsgUpdateToken) (*types.MsgUpdateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	oldToken, found := k.GetToken(ctx, msg.TokenId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "token with denom %s is not exists", msg.TokenId)
	}
	tm, found := k.GetTokenManagement(ctx, msg.TokenId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "token with denom %s is not exists", msg.TokenId)
	}
	if tm.Manager != msg.Manager {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not token manager")
	}

	newToken := types.NewToken(msg.TokenId, msg.Name, msg.Symbol, oldToken.Decimal, msg.Description)
	k.SetToken(ctx, msg.TokenId, newToken)

	return &types.MsgUpdateTokenResponse{}, nil
}

func (k msgServer) AllocateToken(goCtx context.Context, msg *types.MsgAllocateToken) (*types.MsgAllocateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	isFound := k.bankKeeper.HasSupply(ctx, msg.TokenId)
	if isFound {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "token with denom %s already allocate", msg.TokenId)
	}

	tm, found := k.GetTokenManagement(ctx, msg.TokenId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "token with denom %s is not exists", msg.TokenId)
	}
	if tm.Manager != msg.Manager {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not token manager")
	}

	for _, balance := range msg.Balances {
		userAcc, err := sdk.AccAddressFromBech32(balance.Address)
		if err != nil {
			return nil, err
		}

		mintCoins := sdk.NewCoins(sdk.NewCoin(msg.TokenId, sdk.NewIntFromUint64(balance.Amount)))
		err = k.bankKeeper.MintCoins(ctx, types.ModuleName, mintCoins)
		if err != nil {
			return nil, err
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userAcc, mintCoins)
		if err != nil {
			return nil, err
		}
	}
	return &types.MsgAllocateTokenResponse{}, nil
}

func (k msgServer) AssignPrivilege(goCtx context.Context, msg *types.MsgAssignPrivilege) (*types.MsgAssignPrivilegeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	tm, found := k.GetTokenManagement(ctx, msg.TokenId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "token with denom %s is not exists", msg.TokenId)
	}
	if tm.Manager != msg.Manager {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not token manager")
	}
	if !slices.Contains(tm.EnabledPrivileges, msg.GetPrivilege()) {
		if !tm.AddNewPrivilege {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "can't add new privilege")
		} else if slices.Contains(tm.ExcludedPrivileges, msg.Privilege) {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "privilege %s is excluded", msg.Privilege)
		}
		tm.EnabledPrivileges = append(tm.EnabledPrivileges, msg.Privilege)
	}

	for _, user := range msg.AssignedTo {
		userAcc, err := sdk.AccAddressFromBech32(user)
		if err != nil {
			return nil, err
		}
		k.SetTokenPrivilegeAccount(ctx, msg.TokenId, msg.Privilege, userAcc)
		k.SetTokenManagement(ctx, msg.TokenId, tm)
	}

	return &types.MsgAssignPrivilegeResponse{}, nil
}

func (k msgServer) UnassignPrivilege(goCtx context.Context, msg *types.MsgUnassignPrivilege) (*types.MsgUnassignPrivilegeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	tm, found := k.GetTokenManagement(ctx, msg.TokenId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "token with denom %s doesn't exists", msg.TokenId)
	}
	if tm.Manager != msg.Manager {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not token manager")
	}

	for _, user := range msg.UnassignedFrom {
		userAcc, err := sdk.AccAddressFromBech32(user)
		if err != nil {
			return nil, err
		}
		k.DeleteTokenPrivilegeAccount(ctx, msg.TokenId, msg.Privilege, userAcc)
	}

	return &types.MsgUnassignPrivilegeResponse{}, nil
}

func (k msgServer) DisablePrivilege(goCtx context.Context, msg *types.MsgDisablePrivilege) (*types.MsgDisablePrivilegeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	tm, found := k.GetTokenManagement(ctx, msg.TokenId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "token with denom %s is not exists", msg.TokenId)
	}
	if tm.Manager != msg.Manager {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not token manager")
	}
	if slices.Contains(tm.ExcludedPrivileges, msg.DisabledPrivilege) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "privilege %s is already excluded", msg.DisabledPrivilege)
	}

	tm.ExcludedPrivileges = append(tm.ExcludedPrivileges, msg.DisabledPrivilege)
	k.SetTokenManagement(ctx, msg.TokenId, tm)
	return &types.MsgDisablePrivilegeResponse{}, nil
}

func (k msgServer) ExecutePrivilege(goCtx context.Context, msg *types.MsgExecutePrivilege) (*types.MsgExecutePrivilegeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	userAcc, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	tm, found := k.GetTokenManagement(ctx, msg.TokenId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "token with denom %s is not exists", msg.TokenId)
	}
	msgPriv, isPrivilegeMsg := msg.PrivilegeMsg.GetCachedValue().(types.PrivilegeMsgI)
	if !isPrivilegeMsg {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "message is not privilege msg")
	}
	privName := msgPriv.NeedPrivilege()
	if slices.Contains(tm.ExcludedPrivileges, privName) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "privilege %s is excluded", privName)
	}
	if !slices.Contains(k.GetTokenAccountPrivileges(ctx, msg.TokenId, userAcc), privName) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "user does not not have %s privilege", privName)
	}

	privImplementation, ok := k.PrivilegeManager[privName]
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "privilege name %s is not registered yet", privName)
	}

	protoMsg, err := UnpackAnyMsg(msg.PrivilegeMsg)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid privilege message")
	}

	sdkMsg, ok := protoMsg.(sdk.Msg)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "privilege sdk message")
	}

	msgHandler := privImplementation.MsgHandler()
	_, err = msgHandler(ctx, sdkMsg, msg.TokenId, userAcc)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "fail to execute privilege message")
	}

	return &types.MsgExecutePrivilegeResponse{}, nil
}
