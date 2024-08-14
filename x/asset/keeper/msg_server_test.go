package keeper_test

import (
	"fmt"
	"slices"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
)

var (
	managerAddr = sdk.AccAddress(rand.Bytes(address.Len))
	creatorAddr = sdk.AccAddress(rand.Bytes(address.Len))
	userAddr1   = sdk.AccAddress(rand.Bytes(address.Len))
	userAddr2   = sdk.AccAddress(rand.Bytes(address.Len))
	userAddr3   = sdk.AccAddress(rand.Bytes(address.Len))
	name        = "viet nam dong"
	symbol      = "vnd"
	amount      = uint64(1000)
)

func (s *KeeperTestSuite) TestCreateToken() {
	tests := []struct {
		name       string
		expectPass bool
		setup      func() *types.MsgCreateToken
	}{
		{
			name:       "success",
			expectPass: true,
			setup: func() *types.MsgCreateToken {
				return &types.MsgCreateToken{
					Creator:            creatorAddr.String(),
					Manager:            managerAddr.String(),
					Name:               name,
					Symbol:             symbol,
					Decimal:            2,
					Description:        "",
					ExcludedPrivileges: []string{},
					AddNewPrivilege:    true,
				}
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			msg := test.setup()

			_, err := s.msgServer.CreateToken(s.ctx, msg)
			if test.expectPass {
				s.Require().NoError(err)

				lowerCaseSymbol := strings.ToLower(msg.Symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, msg.Creator, lowerCaseSymbol)
				_, isFound := s.assetKeeper.GetTokenManagement(s.ctx, tokenId)
				s.Require().True(isFound)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateToken() {

	tests := []struct {
		name       string
		expectPass bool
		setup      func(keeper.Keeper, sdk.Context) *types.MsgUpdateToken
	}{
		{
			name:       "success",
			expectPass: true,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgUpdateToken {
				description := ""

				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)
				token := types.NewToken(tokenId, strings.ToLower(name), lowerCaseSymbol, 2, description)
				k.SetToken(ctx, tokenId, token)

				tokenManage := types.NewTokenManagement(managerAddr.String(), true, []string{})
				k.SetTokenManagement(ctx, tokenId, tokenManage)

				return &types.MsgUpdateToken{
					Manager:         managerAddr.String(),
					TokenId:         tokenId,
					Name:            name,
					Symbol:          "u" + symbol, // old token is symbol
					Description:     description,
					AddNewPrivilege: false,
				}
			},
		},
		{
			name:       "token not exists",
			expectPass: false,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgUpdateToken {
				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)

				return &types.MsgUpdateToken{
					Manager:         managerAddr.String(),
					TokenId:         tokenId,
					Name:            name,
					Symbol:          symbol,
					Description:     "",
					AddNewPrivilege: false,
				}
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			msg := test.setup(*s.assetKeeper, s.ctx)

			oldToken, isFound := s.assetKeeper.GetToken(s.ctx, msg.TokenId)
			if test.expectPass {
				s.Require().True(isFound)
			} else {
				s.Require().False(isFound)
			}

			_, err := s.msgServer.UpdateToken(s.ctx, msg)
			if test.expectPass {
				s.Require().NoError(err)

				newToken, isFound := s.assetKeeper.GetToken(s.ctx, msg.TokenId)
				s.Require().True(isFound)
				s.Require().NotEqual(oldToken.Symbol, newToken.Symbol)
			} else {
				s.Require().ErrorIs(sdkerrors.ErrNotFound, err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAllocateToken() {
	tests := []struct {
		name       string
		expectPass bool
		setup      func(k keeper.Keeper, ctx sdk.Context) *types.MsgAllocateToken
	}{
		{
			name:       "success",
			expectPass: true,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgAllocateToken {
				description := ""

				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)
				token := types.NewToken(tokenId, strings.ToLower(name), lowerCaseSymbol, 2, description)
				k.SetToken(ctx, tokenId, token)

				tokenManage := types.NewTokenManagement(managerAddr.String(), true, []string{})
				k.SetTokenManagement(ctx, tokenId, tokenManage)

				return &types.MsgAllocateToken{
					Manager: managerAddr.String(),
					TokenId: tokenId,
					Balances: []types.Balance{
						{
							Address: creatorAddr.String(),
							Amount:  amount,
						},
					},
				}
			},
		},
		{
			name:       "token not exists",
			expectPass: false,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgAllocateToken {
				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)

				return &types.MsgAllocateToken{
					Manager: managerAddr.String(),
					TokenId: tokenId,
					Balances: []types.Balance{
						{
							Address: creatorAddr.String(),
							Amount:  amount,
						},
					},
				}
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			msg := test.setup(*s.assetKeeper, s.ctx)

			_, err := s.msgServer.AllocateToken(s.ctx, msg)
			if test.expectPass {
				s.Require().NoError(err)

				coin := s.bankKeeper.GetBalance(s.ctx, creatorAddr, msg.TokenId)
				s.Require().Equal(amount, coin.Amount.Uint64())
			} else {
				s.Require().ErrorIs(sdkerrors.ErrNotFound, err)
			}

		})
	}
}

func (s *KeeperTestSuite) TestAssignPrivilege() {
	tests := []struct {
		name       string
		expectPass bool
		setup      func(k keeper.Keeper, ctx sdk.Context) *types.MsgAssignPrivilege
	}{
		{
			name:       "success",
			expectPass: true,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgAssignPrivilege {
				description := ""

				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)
				token := types.NewToken(tokenId, strings.ToLower(name), lowerCaseSymbol, 2, description)
				k.SetToken(ctx, tokenId, token)

				tokenManage := types.NewTokenManagement(managerAddr.String(), true, []string{})
				k.SetTokenManagement(ctx, tokenId, tokenManage)

				return &types.MsgAssignPrivilege{
					Manager: managerAddr.String(),
					TokenId: tokenId,
					AssignedTo: []string{
						userAddr1.String(),
						userAddr2.String(),
						userAddr3.String(),
					},
					Privilege: creatorAddr.String(),
				}
			},
		},
		{
			name:       "token not exists",
			expectPass: false,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgAssignPrivilege {
				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)

				return &types.MsgAssignPrivilege{
					Manager: managerAddr.String(),
					TokenId: tokenId,
					AssignedTo: []string{
						userAddr1.String(),
						userAddr2.String(),
						userAddr3.String(),
					},
					Privilege: creatorAddr.String(),
				}
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			msg := test.setup(*s.assetKeeper, s.ctx)

			_, err := s.msgServer.AssignPrivilege(s.ctx, msg)
			if test.expectPass {
				s.Require().NoError(err)

				privList := s.assetKeeper.GetTokenAccountPrivileges(s.ctx, msg.TokenId, userAddr1)
				s.Require().True(slices.Contains(privList, msg.Privilege))
			} else {
				s.Require().ErrorIs(sdkerrors.ErrNotFound, err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnassignPrivilege() {

	tests := []struct {
		name       string
		expectPass bool
		setup      func(k keeper.Keeper, ctx sdk.Context) *types.MsgUnassignPrivilege
	}{
		{
			name:       "success",
			expectPass: true,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgUnassignPrivilege {
				description := ""

				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)
				token := types.NewToken(tokenId, strings.ToLower(name), lowerCaseSymbol, 2, description)
				k.SetToken(ctx, tokenId, token)

				tokenManage := types.NewTokenManagement(managerAddr.String(), true, []string{})
				k.SetTokenManagement(ctx, tokenId, tokenManage)

				k.SetTokenPrivilegeAccount(ctx, tokenId, creatorAddr.String(), userAddr1)
				k.SetTokenPrivilegeAccount(ctx, tokenId, creatorAddr.String(), userAddr2)
				k.SetTokenPrivilegeAccount(ctx, tokenId, creatorAddr.String(), userAddr3)
				return &types.MsgUnassignPrivilege{
					Manager: managerAddr.String(),
					TokenId: tokenId,
					UnassignedFrom: []string{
						userAddr1.String(),
						userAddr2.String(),
					},
					Privilege: creatorAddr.String(),
				}
			},
		},
		{
			name:       "token not exists",
			expectPass: false,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgUnassignPrivilege {
				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)
				return &types.MsgUnassignPrivilege{
					Manager: managerAddr.String(),
					TokenId: tokenId,
					UnassignedFrom: []string{
						userAddr1.String(),
						userAddr2.String(),
					},
					Privilege: creatorAddr.String(),
				}
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			msg := test.setup(*s.assetKeeper, s.ctx)

			_, err := s.msgServer.UnassignPrivilege(s.ctx, msg)
			if test.expectPass {
				s.Require().NoError(err)

				privList := s.assetKeeper.GetTokenAccountPrivileges(s.ctx, msg.TokenId, userAddr1)
				s.Require().False(slices.Contains(privList, msg.Privilege))

				privList = s.assetKeeper.GetTokenAccountPrivileges(s.ctx, msg.TokenId, userAddr3) //user3 not in UnassignedFrom
				s.Require().True(slices.Contains(privList, msg.Privilege))
			} else {
				s.Require().ErrorIs(sdkerrors.ErrNotFound, err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestDisablePrivilege() {

	tests := []struct {
		name       string
		expectPass bool
		setup      func(k keeper.Keeper, ctx sdk.Context) *types.MsgDisablePrivilege
	}{
		{
			name:       "success",
			expectPass: true,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgDisablePrivilege {
				description := ""

				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)
				token := types.NewToken(tokenId, strings.ToLower(name), lowerCaseSymbol, 2, description)
				k.SetToken(ctx, tokenId, token)

				tokenManage := types.NewTokenManagement(managerAddr.String(), true, []string{})
				k.SetTokenManagement(ctx, tokenId, tokenManage)

				return &types.MsgDisablePrivilege{
					Manager:           managerAddr.String(),
					TokenId:           tokenId,
					DisabledPrivilege: userAddr1.String(),
				}
			},
		},
		{
			name:       "token not exists",
			expectPass: false,
			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgDisablePrivilege {

				lowerCaseSymbol := strings.ToLower(symbol)
				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)

				return &types.MsgDisablePrivilege{
					Manager:           managerAddr.String(),
					TokenId:           tokenId,
					DisabledPrivilege: userAddr1.String(),
				}
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			msg := test.setup(*s.assetKeeper, s.ctx)

			_, err := s.msgServer.DisablePrivilege(s.ctx, msg)
			if test.expectPass {
				s.Require().NoError(err)

				tm, found := s.assetKeeper.GetTokenManagement(s.ctx, msg.TokenId)
				s.Require().True(found)
				s.Require().True(slices.Contains(tm.ExcludedPrivileges, userAddr1.String()))
			} else {
				s.Require().ErrorIs(sdkerrors.ErrNotFound, err)
			}
		})
	}
}

// func (s *KeeperTestSuite) TestExecutePrivilege() {

// 	tests := []struct {
// 		name       string
// 		expectPass bool
// 		setup      func(k keeper.Keeper, ctx sdk.Context) *types.MsgExecutePrivilege
// 	}{
// 		{
// 			name:       "success",
// 			expectPass: true,
// 			setup: func(k keeper.Keeper, ctx sdk.Context) *types.MsgExecutePrivilege {
// 				description := ""

// 				lowerCaseSymbol := strings.ToLower(symbol)
// 				tokenId := fmt.Sprintf("%s/%s/%s", types.ModuleName, creatorAddr.String(), lowerCaseSymbol)
// 				token := types.NewToken(tokenId, strings.ToLower(name), lowerCaseSymbol, 2, description)
// 				k.SetToken(ctx, tokenId, token)

// 				tokenManage := types.NewTokenManagement(managerAddr.String(), true, []string{})
// 				k.SetTokenManagement(ctx, tokenId, tokenManage)

// 				k.SetTokenPrivilegeAccount(ctx, tokenId, creatorAddr.String(), userAddr1)

// 				var privilegeMsg any.Any
// 				// p, e := any.NewAnyWithValue(privilegeMsg)
// 				// privilege, err := anypb.New(privilegeMsg)
// 				// s.Require().NoError(err)
// 				return &types.MsgExecutePrivilege{
// 					Address:      userAddr1.String(),
// 					TokenId:      tokenId,
// 					PrivilegeMsg: &privilegeMsg,
// 				}
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		s.Run(test.name, func() {
// 			s.SetupTest()
// 			msg := test.setup(*s.assetKeeper, s.ctx)

// 			s.msgServer.ExecutePrivilege(s.ctx, msg)
// 		})
// 	}
// }
