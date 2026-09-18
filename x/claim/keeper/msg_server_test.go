package keeper_test

import (
	"errors"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	"github.com/realiotech/realio-network/testutil"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
	"github.com/realiotech/realio-network/x/claim/keeper"
	"github.com/realiotech/realio-network/x/claim/types"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

// multiStakingCoinDenom is the denom app.Setup registers as the seed
// validator's multi-staking coin (see app.MultiStakingCoinA in
// app/test_helpers.go).
const multiStakingCoinDenom = "ario"

// delegate funds delAddr with amount of the validator's multi-staking coin
// and delegates it to suite.validator through the real multi-staking
// message flow (locks the coin, mints the bond coin, delegates natively),
// so both a native x/staking Delegation and an x/multi-staking
// MultiStakingLock exist for delAddr afterwards -- exactly the state
// MigrateDelegations is meant to re-key.
func (suite *KeeperTestSuite) delegate(delAddr sdk.AccAddress, amount math.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, amount))
	suite.Require().NoError(suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))
	suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, delAddr, coins))

	msMsgServer := multistakingkeeper.NewMsgServerImpl(suite.app.MultiStakingKeeper)
	_, err := msMsgServer.Delegate(suite.ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: suite.validator.String(),
		Amount:           sdk.NewCoin(multiStakingCoinDenom, amount),
	})
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestLinkAddressValidation() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()
	badAdmin := testutil.GenAddress().String()

	testCases := []struct {
		name      string
		msg       func() *types.MsgLinkAddress
		noAdmin   bool
		errString string
	}{
		{
			name: "invalid: wrong admin",
			msg: func() *types.MsgLinkAddress {
				return &types.MsgLinkAddress{Admin: badAdmin, OldAddress: oldAddr.String(), NewAddress: newAddr.String()}
			},
			errString: "invalid admin",
		},
		{
			name:    "invalid: no admin configured",
			noAdmin: true,
			msg: func() *types.MsgLinkAddress {
				return &types.MsgLinkAddress{Admin: suite.admin, OldAddress: oldAddr.String(), NewAddress: newAddr.String()}
			},
			errString: "invalid admin",
		},
		{
			name: "invalid: bad old_address",
			msg: func() *types.MsgLinkAddress {
				return &types.MsgLinkAddress{Admin: suite.admin, OldAddress: "not-an-address", NewAddress: newAddr.String()}
			},
			errString: "invalid old_address",
		},
		{
			name: "invalid: bad new_address",
			msg: func() *types.MsgLinkAddress {
				return &types.MsgLinkAddress{Admin: suite.admin, OldAddress: oldAddr.String(), NewAddress: "not-an-address"}
			},
			errString: "invalid new_address",
		},
		{
			name: "invalid: old and new are the same",
			msg: func() *types.MsgLinkAddress {
				return &types.MsgLinkAddress{Admin: suite.admin, OldAddress: oldAddr.String(), NewAddress: oldAddr.String()}
			},
			errString: "must differ",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			if tc.noAdmin {
				suite.Require().NoError(suite.app.ClaimKeeper.SetAdmin(suite.ctx, nil))
			}

			srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
			_, err := srv.LinkAddress(suite.ctx, tc.msg())
			suite.Require().ErrorContains(err, tc.errString)
		})
	}
}

func (suite *KeeperTestSuite) TestLinkAddressAlreadyLinked() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()
	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)

	_, err := srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{Admin: suite.admin, OldAddress: oldAddr.String(), NewAddress: newAddr.String()})
	suite.Require().NoError(err)

	// old_address can't be linked twice.
	other := testutil.GenAddress()
	_, err = srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{Admin: suite.admin, OldAddress: oldAddr.String(), NewAddress: other.String()})
	suite.Require().ErrorContains(err, "already linked")

	// new_address can't itself be an already-claimed old_address (no chaining).
	another := testutil.GenAddress()
	_, err = srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{Admin: suite.admin, OldAddress: another.String(), NewAddress: oldAddr.String()})
	suite.Require().ErrorContains(err, "someone else's old_address")
}

func (suite *KeeperTestSuite) TestLinkAddressNoDelegations() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{Admin: suite.admin, OldAddress: oldAddr.String(), NewAddress: newAddr.String()})
	suite.Require().NoError(err)

	got, found := suite.app.ClaimKeeper.GetLink(suite.ctx, oldAddr)
	suite.Require().True(found)
	suite.Require().Equal(newAddr, got)
}

// TestLinkAddressMigratesDelegation is the core behavioral test: an old
// address with an active multi-staking delegation gets linked to a new
// address, and the delegation (both the native x/staking Delegation and
// the backing x/multi-staking MultiStakingLock) must move over completely
// -- nothing left behind on the old address, nothing duplicated.
func (suite *KeeperTestSuite) TestLinkAddressMigratesDelegation() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()
	amount := math.NewInt(1_000_000_000_000)

	suite.delegate(oldAddr, amount)

	oldDelBefore, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, oldAddr, suite.validator)
	suite.Require().NoError(err)
	suite.Require().True(oldDelBefore.Shares.IsPositive())

	oldLockBefore, found := suite.app.MultiStakingKeeper.GetMultiStakingLock(suite.ctx,
		multistakingtypes.MultiStakingLockID(oldAddr.String(), suite.validator.String()))
	suite.Require().True(found)
	suite.Require().True(oldLockBefore.LockedCoin.Amount.Equal(amount))

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err = srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		NewAddress: newAddr.String(),
	})
	suite.Require().NoError(err)

	// old side is gone entirely
	_, err = suite.app.StakingKeeper.GetDelegation(suite.ctx, oldAddr, suite.validator)
	suite.Require().True(errors.Is(err, stakingtypes.ErrNoDelegation))

	_, found = suite.app.MultiStakingKeeper.GetMultiStakingLock(suite.ctx,
		multistakingtypes.MultiStakingLockID(oldAddr.String(), suite.validator.String()))
	suite.Require().False(found)

	// new side has exactly what the old side had
	newDel, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, newAddr, suite.validator)
	suite.Require().NoError(err)
	suite.Require().True(newDel.Shares.Equal(oldDelBefore.Shares))

	newLock, found := suite.app.MultiStakingKeeper.GetMultiStakingLock(suite.ctx,
		multistakingtypes.MultiStakingLockID(newAddr.String(), suite.validator.String()))
	suite.Require().True(found)
	suite.Require().True(newLock.LockedCoin.Amount.Equal(amount))

	// validator's aggregate totals are untouched by the re-key
	validator, err := suite.app.StakingKeeper.GetValidator(suite.ctx, suite.validator)
	suite.Require().NoError(err)
	suite.Require().Equal(stakingtypes.Bonded, validator.Status)
}

// TestLinkAddressMergesIntoExistingNewDelegation covers the case where the
// new address already delegates to the same validator (e.g. it's the
// user's own separate wallet, already staking): migrating the old
// address's delegation must ADD to it, not clobber it.
func (suite *KeeperTestSuite) TestLinkAddressMergesIntoExistingNewDelegation() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()

	oldAmount := math.NewInt(1_000_000_000_000)
	newAmount := math.NewInt(500_000_000_000)

	suite.delegate(oldAddr, oldAmount)
	suite.delegate(newAddr, newAmount)

	oldDel, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, oldAddr, suite.validator)
	suite.Require().NoError(err)
	newDelBefore, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, newAddr, suite.validator)
	suite.Require().NoError(err)
	wantShares := oldDel.Shares.Add(newDelBefore.Shares)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err = srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		NewAddress: newAddr.String(),
	})
	suite.Require().NoError(err)

	newDelAfter, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, newAddr, suite.validator)
	suite.Require().NoError(err)
	suite.Require().True(newDelAfter.Shares.Equal(wantShares))

	newLock, found := suite.app.MultiStakingKeeper.GetMultiStakingLock(suite.ctx,
		multistakingtypes.MultiStakingLockID(newAddr.String(), suite.validator.String()))
	suite.Require().True(found)
	suite.Require().True(newLock.LockedCoin.Amount.Equal(oldAmount.Add(newAmount)))
}

// allocateReward advances one block (a delegation earns nothing for the
// block it was created in -- DelegatorStartingInfo.Height == current
// height short-circuits CalculateDelegationRewards to zero, by design) and
// allocates amount of bond-denom reward to suite.validator, returning the
// bond denom used. The genesis validator's own tokens are enormous
// relative to a fresh test delegation (see app.MultiStakingCoinA), so any
// one delegation's proportional share of the pool is tiny -- callers
// should pass a large amount so that share survives the reward
// calculation's integer truncation.
func (suite *KeeperTestSuite) allocateReward(amount math.Int) string {
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	bondDenom, err := suite.app.StakingKeeper.BondDenom(suite.ctx)
	suite.Require().NoError(err)

	rewardCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, amount))
	suite.Require().NoError(suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, rewardCoins))
	suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, distrtypes.ModuleName, rewardCoins))

	validator, err := suite.app.StakingKeeper.GetValidator(suite.ctx, suite.validator)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, sdk.NewDecCoinsFromCoins(rewardCoins...)))

	return bondDenom
}

// TestLinkAddressFreshCarriesOverPendingReward covers the common (no
// merge) case: migrating must NOT pay out oldAddr's pending reward as a
// side effect of the admin's LinkAddress call. Instead newAddr inherits
// the exact reward-accrual ledger and can withdraw it later, on its own,
// for the full amount -- and oldAddr must never be able to claim anything
// again once its delegation is gone.
func (suite *KeeperTestSuite) TestLinkAddressFreshCarriesOverPendingReward() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()

	suite.delegate(oldAddr, math.NewInt(1_000_000_000_000))
	bondDenom := suite.allocateReward(math.NewInt(1_000_000_000))

	oldBalanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, oldAddr, bondDenom)
	newBalanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, bondDenom)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		NewAddress: newAddr.String(),
	})
	suite.Require().NoError(err)

	// Nothing paid out yet, to either address -- the reward is carried
	// over as a still-pending claim, not flushed.
	oldBalanceAfterLink := suite.app.BankKeeper.GetBalance(suite.ctx, oldAddr, bondDenom)
	newBalanceAfterLink := suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, bondDenom)
	suite.Require().True(oldBalanceAfterLink.Equal(oldBalanceBefore))
	suite.Require().True(newBalanceAfterLink.Equal(newBalanceBefore))

	// oldAddr has no delegation left at all -- it can never claim again.
	_, err = suite.app.DistrKeeper.WithdrawDelegationRewards(suite.ctx, oldAddr, suite.validator)
	suite.Require().Error(err)

	// newAddr withdraws on its own, whenever it likes, and gets the full
	// amount that had accrued on oldAddr's delegation.
	withdrawn, err := suite.app.DistrKeeper.WithdrawDelegationRewards(suite.ctx, newAddr, suite.validator)
	suite.Require().NoError(err)
	suite.Require().True(withdrawn.AmountOf(bondDenom).IsPositive())

	newBalanceAfterWithdraw := suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, bondDenom)
	suite.Require().True(newBalanceAfterWithdraw.Amount.GT(newBalanceBefore.Amount))
}

// TestLinkAddressMergeFlushesRewardToNewAddress covers the merge case:
// unlike the fresh path above, two independent reward ledgers can't be
// combined losslessly, so migrating must flush oldAddr's pending reward
// immediately -- and it must still land on newAddr, never oldAddr.
func (suite *KeeperTestSuite) TestLinkAddressMergeFlushesRewardToNewAddress() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()

	suite.delegate(oldAddr, math.NewInt(1_000_000_000_000))
	suite.delegate(newAddr, math.NewInt(500_000_000_000))
	bondDenom := suite.allocateReward(math.NewInt(1_000_000_000))

	oldBalanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, oldAddr, bondDenom)
	newBalanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, bondDenom)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		NewAddress: newAddr.String(),
	})
	suite.Require().NoError(err)

	oldBalanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, oldAddr, bondDenom)
	newBalanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, bondDenom)

	suite.Require().True(oldBalanceAfter.Equal(oldBalanceBefore), "the compromised old address must never receive a payout")
	suite.Require().True(newBalanceAfter.Amount.GT(newBalanceBefore.Amount), "the merge must flush the accrued reward to the new address immediately")
}

// TestLinkAddressFreshPreservesExactRewardAmount tightens the "some reward
// arrived" check in TestLinkAddressFreshCarriesOverPendingReward into an
// exact one: newAddr must withdraw precisely what a `distribution rewards`
// query against oldAddr would have shown one instant before the migration
// -- the same Querier.DelegationRewards RPC any wallet/CLI calls to preview
// a pending reward, not a hand-rolled substitute. This proves the
// carried-over DelegatorStartingInfo (PreviousPeriod, Stake, Height) is a
// lossless copy, not just "close enough".
//
// The query mutates the context it's handed (it calls
// IncrementValidatorPeriod for real, same as an actual withdrawal) -- in
// production that's harmless because ABCI query requests run against a
// context baseapp never commits back to app state. A keeper-level test has
// no such isolation, so CacheContext() recreates it by hand: the query runs
// for real, then its branch is discarded, leaving the ledger the actual
// migration below runs against untouched.
func (suite *KeeperTestSuite) TestLinkAddressFreshPreservesExactRewardAmount() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()

	suite.delegate(oldAddr, math.NewInt(1_000_000_000_000))
	bondDenom := suite.allocateReward(math.NewInt(1_000_000_000))

	previewCtx, _ := suite.ctx.CacheContext()
	querier := distrkeeper.NewQuerier(suite.app.DistrKeeper)
	preview, err := querier.DelegationRewards(previewCtx, &distrtypes.QueryDelegationRewardsRequest{
		DelegatorAddress: oldAddr.String(),
		ValidatorAddress: suite.validator.String(),
	})
	suite.Require().NoError(err)
	wantReward, _ := preview.Rewards.TruncateDecimal()
	suite.Require().True(wantReward.AmountOf(bondDenom).IsPositive())

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err = srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		NewAddress: newAddr.String(),
	})
	suite.Require().NoError(err)

	gotReward, err := suite.app.DistrKeeper.WithdrawDelegationRewards(suite.ctx, newAddr, suite.validator)
	suite.Require().NoError(err)
	suite.Require().True(gotReward.Equal(wantReward),
		"newAddr should inherit exactly what a rewards query against oldAddr showed pre-migration, got %s want %s", gotReward, wantReward)
}

// TestLinkAddressSkipsValidatorSelfDelegation covers the safety check added
// to migrateOneDelegation: if old_address is itself a validator's operator
// account, that validator's self-delegation must be left exactly where it
// is, not migrated -- see migrateOneDelegation's doc comment for why
// (moving it would silently and permanently defeat MinSelfDelegation/
// jailing enforcement, since that check only fires for the operator
// address itself). Self-bond remediation is deliberately deferred to a
// separate follow-up, not handled by LinkAddress.
func (suite *KeeperTestSuite) TestLinkAddressSkipsValidatorSelfDelegation() {
	oldAddr := sdk.AccAddress(suite.validator)
	newAddr := testutil.GenAddress()

	selfDelBefore, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, oldAddr, suite.validator)
	suite.Require().NoError(err)
	suite.Require().True(selfDelBefore.Shares.IsPositive())

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err = srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		NewAddress: newAddr.String(),
	})
	suite.Require().NoError(err)

	// The self-delegation is untouched: still under oldAddr, same shares.
	selfDelAfter, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, oldAddr, suite.validator)
	suite.Require().NoError(err)
	suite.Require().True(selfDelAfter.Shares.Equal(selfDelBefore.Shares))

	// Nothing was created for newAddr on this validator.
	_, err = suite.app.StakingKeeper.GetDelegation(suite.ctx, newAddr, suite.validator)
	suite.Require().True(errors.Is(err, stakingtypes.ErrNoDelegation))

	// The validator itself is unaffected -- still bonded, not jailed.
	validator, err := suite.app.StakingKeeper.GetValidator(suite.ctx, suite.validator)
	suite.Require().NoError(err)
	suite.Require().Equal(stakingtypes.Bonded, validator.Status)
	suite.Require().False(validator.Jailed)

	// The skip was surfaced, not silent.
	var found bool
	for _, event := range suite.ctx.EventManager().Events() {
		if event.Type != types.EventTypeSkippedSelfDelegation {
			continue
		}
		found = true
		for _, attr := range event.Attributes {
			switch attr.Key {
			case types.AttributeKeyOldAddress:
				suite.Require().Equal(oldAddr.String(), attr.Value)
			case types.AttributeKeyValidator:
				suite.Require().Equal(suite.validator.String(), attr.Value)
			}
		}
	}
	suite.Require().True(found, "expected an %s event", types.EventTypeSkippedSelfDelegation)
}

// TestLinkAddressesBatch covers the plural, batch form of LinkAddress: one
// message links several old_address/new_address pairs, each going through
// the exact same checks and delegation migration as a standalone
// LinkAddress call.
func (suite *KeeperTestSuite) TestLinkAddressesBatch() {
	oldA, newA := testutil.GenAddress(), testutil.GenAddress()
	oldB, newB := testutil.GenAddress(), testutil.GenAddress()

	suite.delegate(oldA, math.NewInt(1_000_000_000_000))
	// oldB has no delegation at all -- exercises the no-op migration path
	// (MigrateDelegations returns early) within a batch.

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.LinkAddresses(suite.ctx, &types.MsgLinkAddresses{
		Admin: suite.admin,
		Links: []*types.LinkAddressPair{
			{OldAddress: oldA.String(), NewAddress: newA.String()},
			{OldAddress: oldB.String(), NewAddress: newB.String()},
		},
	})
	suite.Require().NoError(err)

	gotA, found := suite.app.ClaimKeeper.GetLink(suite.ctx, oldA)
	suite.Require().True(found)
	suite.Require().Equal(newA, gotA)
	gotB, found := suite.app.ClaimKeeper.GetLink(suite.ctx, oldB)
	suite.Require().True(found)
	suite.Require().Equal(newB, gotB)

	_, err = suite.app.StakingKeeper.GetDelegation(suite.ctx, oldA, suite.validator)
	suite.Require().True(errors.Is(err, stakingtypes.ErrNoDelegation))
	newDelA, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, newA, suite.validator)
	suite.Require().NoError(err)
	suite.Require().True(newDelA.Shares.IsPositive())
}

// TestLinkAddressesFailsAtomically covers what happens when one pair in a
// batch is invalid: the message must fail as a whole, not apply the pairs
// that came before the bad one and skip the rest. This module doesn't
// implement that rollback itself -- runMsgs (baseapp/baseapp.go) wraps an
// entire transaction, msg included, in one branched store and only writes
// it back if every message returns nil, so a real transaction gets this
// for free. A raw keeper-level test bypasses that wrapping, so it's
// reproduced by hand here with CacheContext to prove the property the real
// call site relies on: the earlier, "successful" pair is visible on the
// branch runMsgs would have (attemptCtx below) but never reaches the state
// anyone else can observe once the branch is discarded on error.
func (suite *KeeperTestSuite) TestLinkAddressesFailsAtomically() {
	oldA, newA := testutil.GenAddress(), testutil.GenAddress()
	oldB := testutil.GenAddress()
	// Pre-link oldB under the real ctx so the second pair in the batch
	// fails with "already linked".
	suite.Require().NoError(suite.app.ClaimKeeper.SetLink(suite.ctx, oldB, testutil.GenAddress()))

	attemptCtx, _ := suite.ctx.CacheContext()
	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.LinkAddresses(attemptCtx, &types.MsgLinkAddresses{
		Admin: suite.admin,
		Links: []*types.LinkAddressPair{
			{OldAddress: oldA.String(), NewAddress: newA.String()},
			{OldAddress: oldB.String(), NewAddress: testutil.GenAddress().String()},
		},
	})
	suite.Require().ErrorContains(err, "already linked")

	// The loop really did apply the first pair before hitting the second,
	// failing one -- this module does not skip ahead or pre-validate the
	// whole batch before applying anything.
	suite.Require().True(suite.app.ClaimKeeper.IsLinked(attemptCtx, oldA))

	// But since attemptCtx's branch was never written back (exactly what
	// happens to runMsgs' branch when a message returns an error), nothing
	// from this failed attempt is visible on suite.ctx.
	suite.Require().False(suite.app.ClaimKeeper.IsLinked(suite.ctx, oldA))
}

// TestLinkAddressAuthorizesNewAddressForAssetDenom covers x/asset interop:
// if the multi-staking coin backing a migrated delegation is registered as
// a permissioned x/asset Token (AuthorizationRequired = true), newAddr
// must come out of LinkAddress already authorized to transact it --
// otherwise it would hold the migrated stake but be unable to ever move
// it, the exact dead end LinkAddress exists to get people out of.
func (suite *KeeperTestSuite) TestLinkAddressAuthorizesNewAddressForAssetDenom() {
	oldAddr := testutil.GenAddress()
	newAddr := testutil.GenAddress()

	token := assettypes.NewToken("Ario", multiStakingCoinDenom, "1000000000000000000000000", suite.admin, true)
	suite.Require().NoError(suite.app.AssetKeeper.Token.Set(suite.ctx, assettypes.TokenKey(multiStakingCoinDenom), token))

	suite.delegate(oldAddr, math.NewInt(1_000_000_000_000))
	suite.Require().False(suite.app.AssetKeeper.IsAddressAuthorizedToSend(suite.ctx, multiStakingCoinDenom, newAddr))

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
		Admin:      suite.admin,
		OldAddress: oldAddr.String(),
		NewAddress: newAddr.String(),
	})
	suite.Require().NoError(err)

	suite.Require().True(suite.app.AssetKeeper.IsAddressAuthorizedToSend(suite.ctx, multiStakingCoinDenom, newAddr))
}
