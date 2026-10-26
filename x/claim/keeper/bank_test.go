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

// setupAttacker configures an attacker address holding amount -- the pool
// ClaimFromAttacker draws on.
func (suite *KeeperTestSuite) setupAttacker(amount math.Int) sdk.AccAddress {
	attacker := testutil.GenAddress()
	suite.Require().NoError(suite.app.ClaimKeeper.SetAttacker(suite.ctx, attacker))
	suite.fund(attacker, multiStakingCoinDenom, amount)
	return attacker
}

func (suite *KeeperTestSuite) balance(addr sdk.AccAddress) math.Int {
	return suite.app.BankKeeper.GetBalance(suite.ctx, addr, multiStakingCoinDenom).Amount
}

func coins(amount int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, math.NewInt(amount)))
}

// ---------------------------------------------------------------------------
// MigrateBankBalance -- the holder's own coins
// ---------------------------------------------------------------------------

func (suite *KeeperTestSuite) TestMigrateBankBalance() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	suite.fund(oldAddr, multiStakingCoinDenom, math.NewInt(400))
	suite.link(oldAddr, newAddr)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	res, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(newAddr.String(), res.NewAddress)

	// No amount is given: the whole balance comes across.
	suite.Require().Equal(math.NewInt(400), res.Swept.AmountOf(multiStakingCoinDenom))
	suite.Require().True(suite.balance(oldAddr).IsZero())
	suite.Require().Equal(math.NewInt(400), suite.balance(newAddr))

	// Nothing is left parked in the module account it travelled through.
	moduleAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	suite.Require().True(suite.balance(moduleAddr).IsZero())
}

// TestMigrateBankBalanceIsSafeToRepeat is half the reason the two
// operations are separate messages: a sweep moves whatever is there now,
// so running it again after more coins arrive is correct, not a double
// payment. Running it over an empty address is a successful sweep of
// nothing, which is what makes it safe to bundle with ClaimFromAttacker
// for an address the attacker drained.
func (suite *KeeperTestSuite) TestMigrateBankBalanceIsSafeToRepeat() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	suite.fund(oldAddr, multiStakingCoinDenom, math.NewInt(100))
	suite.link(oldAddr, newAddr)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	msg := &types.MsgMigrateBankBalance{Admin: suite.admin, OldAddress: oldAddr.String()}

	_, err := srv.MigrateBankBalance(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(100), suite.balance(newAddr))

	// Nothing more to take: a second run succeeds and reports zero, so a
	// re-sweep pass over every recovered address is safe to run blind.
	res, err := srv.MigrateBankBalance(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().True(res.Swept.IsZero())
	suite.Require().True(res.ClaimedFromAttacker.IsZero())
	suite.Require().Equal(math.NewInt(100), suite.balance(newAddr))

	// More arrives later -- a late IBC transfer, a matured unbonding -- and
	// the same message picks it up.
	suite.fund(oldAddr, multiStakingCoinDenom, math.NewInt(70))
	_, err = srv.MigrateBankBalance(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(170), suite.balance(newAddr))
}

func (suite *KeeperTestSuite) TestMigrateBankBalanceValidation() {
	linkedOld, linkedNew := testutil.GenAddress(), testutil.GenAddress()
	unlinked := testutil.GenAddress()
	suite.fund(linkedOld, multiStakingCoinDenom, math.NewInt(100))
	suite.link(linkedOld, linkedNew)

	testCases := []struct {
		name      string
		msg       *types.MsgMigrateBankBalance
		errString string
	}{
		{
			name:      "wrong admin",
			msg:       &types.MsgMigrateBankBalance{Admin: testutil.GenAddress().String(), OldAddress: linkedOld.String()},
			errString: "invalid admin",
		},
		{
			name:      "old_address not linked",
			msg:       &types.MsgMigrateBankBalance{Admin: suite.admin, OldAddress: unlinked.String()},
			errString: "is not linked",
		},
	}

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			attemptCtx, _ := suite.ctx.CacheContext()
			_, err := srv.MigrateBankBalance(attemptCtx, tc.msg)
			suite.Require().ErrorContains(err, tc.errString)
		})
	}
}

// ---------------------------------------------------------------------------
// ClaimFromAttacker -- what was taken before the freeze
// ---------------------------------------------------------------------------

// TestMigrateBankBalanceFromAttacker is the case the sweep cannot reach: an address the
// attacker drained has nothing left to move, so the only thing that
// recovers anything is a payment out of the attacker's own account.
func (suite *KeeperTestSuite) TestMigrateBankBalanceFromAttacker() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	attacker := suite.setupAttacker(math.NewInt(10_000))
	suite.link(oldAddr, newAddr)
	suite.Require().True(suite.balance(oldAddr).IsZero(), "fixture: the old address starts drained")

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	res, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
		Admin:        suite.admin,
		OldAddress:   oldAddr.String(),
		FromAttacker: coins(2_500),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(newAddr.String(), res.NewAddress)

	// Paid to the new address, out of the attacker, not out of nowhere.
	suite.Require().Equal(math.NewInt(2_500), suite.balance(newAddr))
	suite.Require().Equal(math.NewInt(7_500), suite.balance(attacker))

	// The response reports what is left in the pool to distribute.
	suite.Require().Equal(math.NewInt(7_500), res.AttackerRemaining.AmountOf(multiStakingCoinDenom))
}

// TestMigrateBankBalanceDrawsOneSharedPool: the attacker's balance is a
// single finite pool across every victim, which is why this is sized by
// the admin and reconciled rather than derived per address.
func (suite *KeeperTestSuite) TestMigrateBankBalanceDrawsOneSharedPool() {
	attacker := suite.setupAttacker(math.NewInt(1_000))
	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)

	var newAddrs []sdk.AccAddress
	for range 3 {
		oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
		suite.link(oldAddr, newAddr)
		newAddrs = append(newAddrs, newAddr)
		_, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
			Admin: suite.admin, OldAddress: oldAddr.String(), FromAttacker: coins(300),
		})
		suite.Require().NoError(err)
	}
	for _, a := range newAddrs {
		suite.Require().Equal(math.NewInt(300), suite.balance(a))
	}
	suite.Require().Equal(math.NewInt(100), suite.balance(attacker))

	// The pool runs dry, and the fourth claim is refused rather than
	// over-drawing it.
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	suite.link(oldAddr, newAddr)
	_, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
		Admin: suite.admin, OldAddress: oldAddr.String(), FromAttacker: coins(300),
	})
	suite.Require().ErrorContains(err, "insufficient funds")
	suite.Require().Equal(math.NewInt(100), suite.balance(attacker))
}

// TestMigrateBankBalanceFromAttackerIsNotIdempotent pins the sharp edge
// of the merged message: the sweep half is safe to repeat, the payout half
// is not. Re-sending with from_attacker set pays again. To re-sweep an
// address later, send it with from_attacker empty.
func (suite *KeeperTestSuite) TestMigrateBankBalanceFromAttackerIsNotIdempotent() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	suite.setupAttacker(math.NewInt(10_000))
	suite.link(oldAddr, newAddr)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	msg := &types.MsgMigrateBankBalance{Admin: suite.admin, OldAddress: oldAddr.String(), FromAttacker: coins(500)}

	_, err := srv.MigrateBankBalance(suite.ctx, msg)
	suite.Require().NoError(err)
	_, err = srv.MigrateBankBalance(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(1_000), suite.balance(newAddr),
		"the chain cannot tell a second award from a repeat; that is the admin's to get right")
}

func (suite *KeeperTestSuite) TestMigrateBankBalanceFromAttackerValidation() {
	linkedOld, linkedNew := testutil.GenAddress(), testutil.GenAddress()
	unlinked := testutil.GenAddress()
	attacker := suite.setupAttacker(math.NewInt(100))
	suite.link(linkedOld, linkedNew)

	// The attacker linked to somewhere, to prove it is refused as a subject.
	suite.link(attacker, testutil.GenAddress())

	testCases := []struct {
		name      string
		setup     func(sdk.Context)
		msg       *types.MsgMigrateBankBalance
		errString string
	}{
		{
			name:      "wrong admin",
			msg:       &types.MsgMigrateBankBalance{Admin: testutil.GenAddress().String(), OldAddress: linkedOld.String(), FromAttacker: coins(10)},
			errString: "invalid admin",
		},
		{
			name:      "old_address not linked",
			msg:       &types.MsgMigrateBankBalance{Admin: suite.admin, OldAddress: unlinked.String(), FromAttacker: coins(10)},
			errString: "is not linked",
		},
		{
			name:      "more than the pool holds",
			msg:       &types.MsgMigrateBankBalance{Admin: suite.admin, OldAddress: linkedOld.String(), FromAttacker: coins(101)},
			errString: "insufficient funds",
		},
		{
			name:      "the attacker is owed nothing",
			msg:       &types.MsgMigrateBankBalance{Admin: suite.admin, OldAddress: attacker.String(), FromAttacker: coins(10)},
			errString: "is the attacker",
		},
		{
			name:      "no attacker configured",
			setup:     func(ctx sdk.Context) { suite.Require().NoError(suite.app.ClaimKeeper.SetAttacker(ctx, nil)) },
			msg:       &types.MsgMigrateBankBalance{Admin: suite.admin, OldAddress: linkedOld.String(), FromAttacker: coins(10)},
			errString: "no attacker address is configured",
		},
	}

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			attemptCtx, _ := suite.ctx.CacheContext()
			if tc.setup != nil {
				tc.setup(attemptCtx)
			}
			_, err := srv.MigrateBankBalance(attemptCtx, tc.msg)
			suite.Require().ErrorContains(err, tc.errString)
		})
	}
}

// ---------------------------------------------------------------------------
// Shared properties
// ---------------------------------------------------------------------------

// TestBankClaimsSurviveBlacklisting is why both go through the module
// account. The old address and the attacker are both frozen, so a direct
// send out of either is refused, while the claims go through.
func (suite *KeeperTestSuite) TestBankClaimsSurviveBlacklisting() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	attacker := suite.setupAttacker(math.NewInt(10_000))
	suite.fund(oldAddr, multiStakingCoinDenom, math.NewInt(1_000))
	suite.link(oldAddr, newAddr)
	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, oldAddr))
	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, attacker))

	for name, from := range map[string]sdk.AccAddress{"victim": oldAddr, "attacker": attacker} {
		directCtx, _ := suite.ctx.CacheContext()
		suite.Require().ErrorContains(
			suite.app.BankKeeper.SendCoins(directCtx, from, newAddr, coins(100)), "blacklisted",
			"a direct send out of the frozen %s should be refused", name)
	}

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
		Admin: suite.admin, OldAddress: oldAddr.String(),
	})
	suite.Require().NoError(err)
	_, err = srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
		Admin: suite.admin, OldAddress: oldAddr.String(), FromAttacker: coins(3_000),
	})
	suite.Require().NoError(err)

	suite.Require().True(suite.balance(oldAddr).IsZero())
	suite.Require().Equal(math.NewInt(7_000), suite.balance(attacker))
	suite.Require().Equal(math.NewInt(4_000), suite.balance(newAddr))
}

// TestBankClaimsLeaveAssetAuthorizationAlone: the module-account hop skips
// x/asset's restriction, so a permissioned token lands on the new address
// whether or not that address may hold it -- and this module does not put
// it on the authorized list. Only the Token's manager can, with
// MsgAuthorizeAddress.
func (suite *KeeperTestSuite) TestBankClaimsLeaveAssetAuthorizationAlone() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	suite.fund(oldAddr, multiStakingCoinDenom, math.NewInt(1_000))
	suite.link(oldAddr, newAddr)

	token := assettypes.NewToken("Ario", multiStakingCoinDenom, "1000000000000000000000000", suite.admin, true)
	suite.Require().NoError(suite.app.AssetKeeper.Token.Set(suite.ctx, assettypes.TokenKey(multiStakingCoinDenom), token))

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
		Admin: suite.admin, OldAddress: oldAddr.String(),
	})
	suite.Require().NoError(err)

	// The coins arrived...
	suite.Require().Equal(math.NewInt(1_000), suite.balance(newAddr))
	// ...and the new address still cannot send them on.
	suite.Require().False(suite.app.AssetKeeper.IsAddressAuthorizedToSend(suite.ctx, multiStakingCoinDenom, newAddr))
	suite.Require().Error(suite.app.BankKeeper.SendCoins(
		suite.ctx, newAddr, testutil.GenAddress(), coins(1),
	))
}

// TestBundledBankClaimsForADrainedAddress is the case that decides whether
// the two bank messages can share a transaction. An address the attacker
// emptied has nothing to sweep -- on mainnet exactly one victim is in that
// position, and it is owed 58,041.338 RIO. If a zero sweep were an error,
// bundling would roll that payout back with it.
func (suite *KeeperTestSuite) TestBundledBankClaimsForADrainedAddress() {
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()
	suite.setupAttacker(math.NewInt(10_000))
	suite.link(oldAddr, newAddr)
	suite.Require().True(suite.balance(oldAddr).IsZero(), "fixture: drained to empty")

	// Both messages in one transaction is one cache branch that either
	// commits whole or not at all -- which is what runMsgs gives them.
	txCtx, commit := suite.ctx.CacheContext()
	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)

	sweep, err := srv.MigrateBankBalance(txCtx, &types.MsgMigrateBankBalance{
		Admin: suite.admin, OldAddress: oldAddr.String(),
	})
	suite.Require().NoError(err, "a sweep of an empty address must not fail the transaction")
	suite.Require().True(sweep.Swept.IsZero())

	_, err = srv.MigrateBankBalance(txCtx, &types.MsgMigrateBankBalance{
		Admin: suite.admin, OldAddress: oldAddr.String(), FromAttacker: coins(5_000),
	})
	suite.Require().NoError(err)
	commit()

	suite.Require().Equal(math.NewInt(5_000), suite.balance(newAddr))
}
