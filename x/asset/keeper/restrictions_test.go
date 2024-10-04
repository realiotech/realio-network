package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"
)

func (suite *KeeperTestSuite) TestRestrictions() {
	srv := keeper.NewMsgServerImpl(suite.app.AssetKeeper)
	wctx := sdk.WrapSDKContext(suite.ctx)
	manager := suite.testUser1Address
	testUser := suite.testUser2Address
	t1 := &types.MsgCreateToken{
		Manager: manager,
		Symbol:  "RST", Total: "1000", AuthorizationRequired: true,
	}
	_, err := srv.CreateToken(wctx, t1)
	suite.Require().NoError(err)

	authUserMsg := &types.MsgAuthorizeAddress{
		Manager: manager,
		Symbol:  "RST", Address: testUser,
	}

	_, _ = srv.AuthorizeAddress(wctx, authUserMsg)

	cases := []struct {
		name    string
		from    sdk.AccAddress
		to      sdk.AccAddress
		amount  sdk.Coins
		expPass bool
	}{
		{
			"module accounts can send to any account",
			suite.app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName),
			suite.testUser1Acc,
			sdk.NewCoins(sdk.NewCoin("arst", math.NewInt(100))),
			true,
		},
		{
			"module accounts can send to any account",
			suite.app.AccountKeeper.GetModuleAddress(stakingtypes.NotBondedPoolName),
			suite.testUser1Acc,
			sdk.NewCoins(sdk.NewCoin("arst", math.NewInt(100))),
			true,
		},
		{
			"module accounts can send to any account",
			suite.app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName),
			suite.testUser1Acc,
			sdk.NewCoins(sdk.NewCoin("arst", math.NewInt(100))),
			true,
		},
		{
			"unauthorized accounts cannot send",
			suite.testUser3Acc,
			suite.testUser1Acc,
			sdk.NewCoins(sdk.NewCoin("arst", math.NewInt(100))),
			false,
		},
	}

	for _, tc := range cases {
		toAddr, err := suite.app.AssetKeeper.AssetSendRestriction(suite.ctx, tc.from, tc.to, tc.amount)
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
			suite.Require().Equal(toAddr, tc.to)
		} else {
			suite.Require().Error(err)
		}
	}
}
