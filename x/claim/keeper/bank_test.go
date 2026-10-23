package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/testutil"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
	"github.com/realiotech/realio-network/x/claim/keeper"
	"github.com/realiotech/realio-network/x/claim/types"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

// fund mints denom to addr.
func (suite *KeeperTestSuite) fund(addr sdk.AccAddress, denom string, amount math.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
	suite.Require().NoError(suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))
	suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, coins))
}

func (suite *KeeperTestSuite) TestClaimBalancesMovesStatedAmount() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	suite.fund(oldAddr, multiStakingCoinDenom, math.NewInt(1_000))
	suite.link(oldAddr, newAddr)

	// The admin states the amount, so a recovery can be scoped to part of
	// what the old address holds rather than always draining it.
	amount := sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, math.NewInt(600)))

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	res, err := srv.ClaimBalances(suite.ctx, &types.MsgClaimBalances{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		Amount:     amount,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(newAddr.String(), res.NewAddress)

	suite.Require().Equal(math.NewInt(400), suite.app.BankKeeper.GetBalance(suite.ctx, oldAddr, multiStakingCoinDenom).Amount)
	suite.Require().Equal(math.NewInt(600), suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, multiStakingCoinDenom).Amount)

	// Nothing is left parked in the module account it travelled through.
	moduleAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	suite.Require().True(suite.app.BankKeeper.GetBalance(suite.ctx, moduleAddr, multiStakingCoinDenom).IsZero())
}

// TestClaimBalancesFromBlacklistedAddress is the reason the coins hop
// through the module account at all. A blacklisted address cannot pay a
// user account -- which is exactly what a recovery is -- so a direct send
// is refused, while the claim goes through.
func (suite *KeeperTestSuite) TestClaimBalancesFromBlacklistedAddress() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	suite.fund(oldAddr, multiStakingCoinDenom, math.NewInt(1_000))
	suite.link(oldAddr, newAddr)
	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, oldAddr))

	amount := sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, math.NewInt(1_000)))

	// A plain transfer out of the blacklisted address is refused...
	directCtx, _ := suite.ctx.CacheContext()
	err := suite.app.BankKeeper.SendCoins(directCtx, oldAddr, newAddr, amount)
	suite.Require().ErrorContains(err, "blacklisted")

	// ...while the claim, which escrows via the module account first,
	// completes.
	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err = srv.ClaimBalances(suite.ctx, &types.MsgClaimBalances{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		Amount:     amount,
	})
	suite.Require().NoError(err)

	suite.Require().True(suite.app.BankKeeper.GetBalance(suite.ctx, oldAddr, multiStakingCoinDenom).IsZero())
	suite.Require().Equal(math.NewInt(1_000), suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, multiStakingCoinDenom).Amount)
}

// TestClaimBalancesLeavesAssetAuthorizationAlone: the module-account hop
// skips x/asset's restriction, so a permissioned token lands on the new
// address whether or not that address may hold it -- and this module does
// not put it on the authorized list. The transfer therefore succeeds and
// leaves the new address holding something only the Token's manager can
// unlock for it, with MsgAuthorizeAddress.
func (suite *KeeperTestSuite) TestClaimBalancesLeavesAssetAuthorizationAlone() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	suite.fund(oldAddr, multiStakingCoinDenom, math.NewInt(1_000))
	suite.link(oldAddr, newAddr)

	token := assettypes.NewToken("Ario", multiStakingCoinDenom, "1000000000000000000000000", suite.admin, true)
	suite.Require().NoError(suite.app.AssetKeeper.Token.Set(suite.ctx, assettypes.TokenKey(multiStakingCoinDenom), token))

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.ClaimBalances(suite.ctx, &types.MsgClaimBalances{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		Amount:     sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, math.NewInt(1_000))),
	})
	suite.Require().NoError(err)

	// The coins arrived...
	suite.Require().Equal(math.NewInt(1_000), suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, multiStakingCoinDenom).Amount)
	// ...and the new address still cannot send them on.
	suite.Require().False(suite.app.AssetKeeper.IsAddressAuthorizedToSend(suite.ctx, multiStakingCoinDenom, newAddr))
	suite.Require().Error(suite.app.BankKeeper.SendCoins(
		suite.ctx, newAddr, testutil.GenAddress(),
		sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, math.NewInt(1))),
	))
}

func (suite *KeeperTestSuite) TestClaimBalancesValidation() {
	linkedOld, linkedNew := testutil.GenAddress(), testutil.GenAddress()
	unlinked := testutil.GenAddress()
	suite.fund(linkedOld, multiStakingCoinDenom, math.NewInt(100))
	suite.link(linkedOld, linkedNew)

	some := sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, math.NewInt(10)))

	testCases := []struct {
		name      string
		msg       *types.MsgClaimBalances
		errString string
	}{
		{
			name:      "wrong admin",
			msg:       &types.MsgClaimBalances{Admin: testutil.GenAddress().String(), OldAddress: linkedOld.String(), Amount: some},
			errString: "invalid admin",
		},
		{
			name:      "old_address not linked",
			msg:       &types.MsgClaimBalances{Admin: suite.admin, OldAddress: unlinked.String(), Amount: some},
			errString: "is not linked",
		},
		{
			name:      "no amount",
			msg:       &types.MsgClaimBalances{Admin: suite.admin, OldAddress: linkedOld.String()},
			errString: "amount must not be empty",
		},
		{
			name: "more than the address holds",
			msg: &types.MsgClaimBalances{
				Admin: suite.admin, OldAddress: linkedOld.String(),
				Amount: sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, math.NewInt(101))),
			},
			errString: "insufficient funds",
		},
	}

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			attemptCtx, _ := suite.ctx.CacheContext()
			_, err := srv.ClaimBalances(attemptCtx, tc.msg)
			suite.Require().ErrorContains(err, tc.errString)
		})
	}
}
