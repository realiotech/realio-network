package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/realiotech/realio-network/testutil"
)

// fundAccount mints coins directly to addr, bypassing SendCoins (and thus
// the send restriction under test) — the setup step, not the thing being
// tested.
func (suite *KeeperTestSuite) fundAccount(addr sdk.AccAddress, coins sdk.Coins) {
	suite.Require().NoError(suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))
	suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, coins))
}

// TestBlacklistSendRestrictionBlocksBlacklistedSender proves the actual gap
// this closes: a blacklisted address can't move its own coins out via
// SendCoins, regardless of what's driving the call (a plain MsgSend, an
// authz MsgExec's inner message, a feegrant fee deduction, an ERC-20
// precompile's transferFrom — all of them eventually call SendCoins).
func (suite *KeeperTestSuite) TestBlacklistSendRestrictionBlocksBlacklistedSender() {
	blacklisted := testutil.GenAddress()
	recipient := testutil.GenAddress()
	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))

	suite.fundAccount(blacklisted, coins)
	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, blacklisted))

	err := suite.app.BankKeeper.SendCoins(suite.ctx, blacklisted, recipient, coins)
	suite.Require().Error(err, "a blacklisted address must not be able to move its own coins out via SendCoins")

	// and the funds must genuinely still be there — not partially moved
	balance := suite.app.BankKeeper.GetBalance(suite.ctx, blacklisted, "stake")
	suite.Require().Equal(coins.AmountOf("stake"), balance.Amount)
}

// TestBlacklistSendRestrictionAllowsBlacklistedRecipient proves the
// restriction deliberately only checks the source: funds can still land ON
// a blacklisted address (a matured unbonding delegation, a reward payout,
// anyone sending it money). Blocking that direction too would risk a panic
// inside the chain's own automatic payout paths (BeginBlock/EndBlock code
// that doesn't expect SendCoins to fail), not just reject a transaction.
func (suite *KeeperTestSuite) TestBlacklistSendRestrictionAllowsBlacklistedRecipient() {
	blacklisted := testutil.GenAddress()
	sender := testutil.GenAddress()
	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))

	suite.fundAccount(sender, coins)
	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, blacklisted))

	err := suite.app.BankKeeper.SendCoins(suite.ctx, sender, blacklisted, coins)
	suite.Require().NoError(err, "incoming transfers to a blacklisted address must still succeed")

	balance := suite.app.BankKeeper.GetBalance(suite.ctx, blacklisted, "stake")
	suite.Require().Equal(coins.AmountOf("stake"), balance.Amount)
}

// TestBlacklistSendRestrictionAllowsModuleDirectedSendsFromBlacklisted
// proves the deliberate carve-out: a blacklisted address's own balance
// moving TO a module account (e.g. multistaking burning its locked
// "representation" coin as part of paying out an already-matured
// unbonding) must still succeed. Discovered the hard way: an earlier
// version of this restriction checked fromAddr unconditionally and panicked
// x/multi-staking's EndBlocker the moment a blacklisted delegator's
// unbonding matured — since that burn-then-payout bookkeeping debits the
// delegator's own account before crediting it back, and multistaking's
// EndBlocker isn't written to handle SendCoins failing there. See the
// comment on BlacklistSendRestriction for the full reasoning, including why
// this doesn't reopen the feegrant gap (closed separately, at the ante
// layer, precisely because it collides with this carve-out).
func (suite *KeeperTestSuite) TestBlacklistSendRestrictionAllowsModuleDirectedSendsFromBlacklisted() {
	blacklisted := testutil.GenAddress()
	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))

	suite.fundAccount(blacklisted, coins)
	suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, blacklisted))

	err := suite.app.BankKeeper.SendCoinsFromAccountToModule(suite.ctx, blacklisted, authtypes.FeeCollectorName, coins)
	suite.Require().NoError(err, "a blacklisted address's own balance must still be movable to a module account — "+
		"automatic protocol settlement (unbonding/reward bookkeeping) depends on this not failing")
}
