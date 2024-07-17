package keeper

import (
	"context"
	"fmt"
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
		DenomUnits: []*bank.DenomUnit{{Denom: lowerCaseSymbol, Exponent: msg.Decimal}, {Denom: tokenId, Exponent: 0}},
	})

	tokenManage := types.NewTokenManagement(msg.Manager, msg.AddNewPrivilege, msg.ExcludedPrivileges)
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
	return nil, nil
}

func (k msgServer) AssignPrivilege(goCtx context.Context, msg *types.MsgAssignPrivilege) (*types.MsgAssignPrivilegeResponse, error) {
	return nil, nil
}

func (k msgServer) UnassignPrivilege(goCtx context.Context, msg *types.MsgUnassignPrivilege) (*types.MsgUnassignPrivilegeResponse, error) {
	return nil, nil
}

func (k msgServer) DisablePrivilege(goCtx context.Context, msg *types.MsgDisablePrivilege) (*types.MsgDisablePrivilegeResponse, error) {
	return nil, nil
}

func (k msgServer) ExecutePrivilege(goCtx context.Context, msg *types.MsgExecutePrivilege) (*types.MsgExecutePrivilegeResponse, error) {
	return nil, nil
}