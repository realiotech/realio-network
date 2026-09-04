package migrations_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	osecp256k1 "github.com/cosmos/evm/crypto/ethsecp256k1"
	"github.com/cosmos/evm/encoding"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"

	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/app/ante"
)

// leakedRealValidator is one of the two real leaked validators from the
// actual incident export (recover_genesis.json). Both are still present in
// x/staking, bonded, self-consistent — only their operator/self-delegator
// ACCOUNT address (not their consensus key) is believed to have leaked, so
// the remediation chosen (see conversation) is to blacklist that account,
// not rotate the validator identity: it keeps signing and earning rewards
// exactly as before, it just can never sign a transaction again.
type leakedRealValidator struct {
	name       string
	oldOperVal string // valoper form
	oldOperAcc string // acc form of the same bytes
}

var leakedRealValidators = []leakedRealValidator{
	{
		name:       "Teshy 103k Bonded RIO",
		oldOperVal: "realiovaloper18a32el4maw3pqr8xh3yrl9ja4lejs265a5nxtm",
		oldOperAcc: "realio18a32el4maw3pqr8xh3yrl9ja4lejs265fqsu6a",
	},
	{
		name:       "Teshy 136k Self Bonded RST",
		oldOperVal: "realiovaloper13jrrtkfuuvzdak6zxmr95hek9c228ug50sdsvs",
		oldOperAcc: "realio13jrrtkfuuvzdak6zxmr95hek9c228ug5myw2ak",
	},
}

// TestBlacklistForkAgainstRealGenesis runs the real blacklist fork (see
// forks.go's ScheduleForkUpgrade, at its real, unmodified BlacklistForkHeight)
// against the real pre-incident genesis export, then checks the three
// properties agreed for the "blacklist only, no rotation" remediation:
//
//  1. both real leaked validators keep signing/earning rewards after the
//     fork — their Validator record (tokens, shares, consensus pubkey,
//     bonded status) is untouched by the fork, and the reward-crediting
//     keeper path has no blacklist check in it;
//  2. the leaked operator account can never get a transaction through —
//     rejected by the blacklist ante decorator specifically, before
//     signature verification is ever reached, so it doesn't matter that we
//     don't hold this address's real private key;
//  3. an ordinary (non-blacklisted) user can still delegate to, and
//     undelegate from, a validator whose operator happens to be blacklisted
//     — blacklist status is a property of the operator ACCOUNT, not of the
//     validator entity delegators interact with.
func TestBlacklistForkAgainstRealGenesis(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, setupTime := app.SetupWithRealGenesis(t)
	ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, setupTime)

	// Both real leaked validators' actual stake/commission/rewards are
	// denominated in "ario" — the multistaking coin these two validators
	// were genuinely bonded with. NOT the same as
	// StakingKeeper.BondDenom(ctx), which returns the vestigial global
	// x/staking param "stake": multistaking layers multi-denom bonding on
	// top of x/staking, so that single legacy param isn't what any real
	// validator's stake is actually measured in on this chain.
	const bondDenom = "ario"
	var err error

	type valAddrs struct {
		acc sdk.AccAddress
		val sdk.ValAddress
	}
	addrs := make(map[string]valAddrs, len(leakedRealValidators))
	for _, lv := range leakedRealValidators {
		accAddr, err := sdk.AccAddressFromBech32(lv.oldOperAcc)
		require.NoError(t, err)
		valAddr, err := sdk.ValAddressFromBech32(lv.oldOperVal)
		require.NoError(t, err)
		addrs[lv.name] = valAddrs{acc: accAddr, val: valAddr}

		// --- 0. sanity, before the fork runs: not blacklisted yet, and
		// genuinely a bonded validator in this genesis (otherwise the test
		// below wouldn't be exercising anything meaningful). ---
		require.False(t, realioApp.BlacklistKeeper.IsBlacklisted(ctx, accAddr),
			"%s must not be blacklisted before the fork runs", lv.name)

		val, err := realioApp.StakingKeeper.GetValidator(ctx, valAddr)
		require.NoErrorf(t, err, "%s must exist in this genesis", lv.name)
		require.Truef(t, val.IsBonded(), "%s must be bonded in this genesis for the test to be meaningful", lv.name)
	}

	// --- 0b. create two genuinely in-flight unbonding delegations on Teshy
	// RIO, right before the fork runs, via the real multistaking Undelegate
	// path — one from a real delegator who is ABOUT to be blacklisted by
	// this same fork run (a separate leak from the validator's), one from a
	// real delegator who stays clean. Both are real genesis delegators with
	// real shares, not test-created ones. This is what "an undelegate
	// that's already in progress on a leaked validator" means concretely:
	// queued before the fork, maturing after it. ---
	rioVal := addrs[leakedRealValidators[0].name].val
	unbondingDelegatorLeaked, err := sdk.AccAddressFromBech32("realio1d2rjp2kxc7md7q9xmjmslmludexv5lvk338k6j")
	require.NoError(t, err)
	unbondingDelegatorClean, err := sdk.AccAddressFromBech32("realio1zkdk43nk7hjtdjsvj2kqvxu7k92vs9uk498wfp")
	require.NoError(t, err)

	require.False(t, realioApp.BlacklistKeeper.IsBlacklisted(ctx, unbondingDelegatorLeaked),
		"fixture assumption broken: expected this delegator to still be clean before the fork runs")
	require.False(t, realioApp.BlacklistKeeper.IsBlacklisted(ctx, unbondingDelegatorClean),
		"fixture assumption broken: expected this delegator to never be on the leaked list")

	unbondAmount := sdk.NewCoin("ario", sdkmath.NewIntFromUint64(100).MulRaw(1_000_000_000_000_000_000)) // 100 ario, well under either delegator's real balance
	for _, delAddr := range []sdk.AccAddress{unbondingDelegatorLeaked, unbondingDelegatorClean} {
		_, err := realioApp.MultiStakingKeeper.Undelegate(ctx, stakingtypes.NewMsgUndelegate(
			delAddr.String(), rioVal.String(), unbondAmount))
		require.NoErrorf(t, err, "creating a fresh in-flight unbonding delegation for %s must succeed", delAddr)
	}

	pendingUBDsBefore, err := realioApp.StakingKeeper.GetUnbondingDelegationsFromValidator(ctx, rioVal)
	require.NoError(t, err)
	require.GreaterOrEqualf(t, len(pendingUBDsBefore), 2, "expected at least the two freshly created in-flight unbondings")

	// --- 1. run the real blacklist fork, at its real, unmodified height —
	// no test-only override needed, since this genesis's initial_height is
	// itself the real BlacklistForkHeight (see SetupWithRealGenesis). ---
	realioApp.ScheduleForkUpgrade(ctx)

	commissionsBefore := make(map[string]sdk.DecCoins, len(leakedRealValidators))
	voteInfos := make([]abci.VoteInfo, 0, len(leakedRealValidators))
	powerReduction := realioApp.StakingKeeper.PowerReduction(ctx)

	for _, lv := range leakedRealValidators {
		a := addrs[lv.name]

		require.Truef(t, realioApp.BlacklistKeeper.IsBlacklisted(ctx, a.acc),
			"%s's operator account must be blacklisted after the fork", lv.name)

		// --- 2. the fork must not touch staking state at all: same
		// tokens/shares/consensus key, still bonded — proving it keeps
		// signing and remains eligible for rewards exactly as before. ---
		val, err := realioApp.StakingKeeper.GetValidator(ctx, a.val)
		require.NoErrorf(t, err, "%s must still exist after the fork", lv.name)
		require.Truef(t, val.IsBonded(), "%s must still be bonded after being blacklisted", lv.name)
		require.Falsef(t, val.Jailed, "%s must not be jailed by the fork", lv.name)

		commission, err := realioApp.DistrKeeper.GetValidatorAccumulatedCommission(ctx, a.val)
		require.NoError(t, err)
		commissionsBefore[lv.name] = commission.Commission

		consAddr, err := val.GetConsAddr()
		require.NoError(t, err)
		voteInfos = append(voteInfos, abci.VoteInfo{
			Validator:   abci.Validator{Address: consAddr, Power: val.ConsensusPower(powerReduction)},
			BlockIdFlag: cmtproto.BlockIDFlagCommit,
		})
	}

	// --- 3. reward/commission crediting through a REAL BeginBlocker/
	// EndBlocker cycle — not a direct keeper call. Both leaked validators
	// are credited as having signed the previous block (100% of
	// "previousTotalPower" between the two, so the resulting reward is
	// cleanly attributable to them) and BeginBlocker is driven exactly the
	// way a live node runs it every block: x/mint mints inflation into the
	// fee collector, x/distribution's BeginBlocker then allocates it by
	// vote power, all inside app.mm.BeginBlock. Nothing in that path checks
	// blacklist status anywhere. ---
	nextCtx := app.NewHeaderCtx(realioApp, initialHeight+1, proposerAddr, setupTime.Add(time.Second)).WithVoteInfos(voteInfos)
	_, err = realioApp.BeginBlocker(nextCtx)
	require.NoError(t, err)
	_, err = realioApp.EndBlocker(nextCtx)
	require.NoError(t, err)

	for _, lv := range leakedRealValidators {
		a := addrs[lv.name]
		commissionAfter, err := realioApp.DistrKeeper.GetValidatorAccumulatedCommission(nextCtx, a.val)
		require.NoError(t, err)
		require.Truef(t,
			commissionAfter.Commission.AmountOf(bondDenom).GT(commissionsBefore[lv.name].AmountOf(bondDenom)),
			"%s's accumulated commission must grow through a real BeginBlocker/EndBlocker cycle even though its operator is blacklisted (before=%s after=%s)",
			lv.name, commissionsBefore[lv.name], commissionAfter.Commission)
	}

	// --- 4. the leaked operator account can never get a transaction
	// through: CosmosBlacklistDecorator itself, called directly, rejects
	// it. Deliberately not routed through the full ante chain / CheckTx: it
	// reads GetSigners() straight off the message body and never touches
	// the signature at all, so an unsigned tx is enough to prove the
	// property — and doing it this way keeps the assertion pinned to the
	// blacklist decorator specifically, instead of depending on whichever
	// unrelated decorator (fee, pubkey-consistency, ...) happens to run
	// earliest in the full chain for a tx we can't produce a real signature
	// for anyway (we don't hold, and never will, this address's real
	// private key). ---
	leakedRIO := addrs[leakedRealValidators[0].name]
	leakedRST := addrs[leakedRealValidators[1].name]

	encodingConfig := encoding.MakeConfig(app.MainnetEVMChainID)
	unsignedTxBuilder := encodingConfig.TxConfig.NewTxBuilder()
	sendMsg := banktypes.NewMsgSend(leakedRIO.acc, leakedRST.acc, sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1)))
	require.NoError(t, unsignedTxBuilder.SetMsgs(sendMsg))

	blacklistDecorator := ante.NewCosmosBlacklistDecorator(realioApp.BlacklistKeeper, realioApp.AppCodec())
	noopNext := func(c sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return c, nil }
	_, err = blacklistDecorator.AnteHandle(nextCtx, unsignedTxBuilder.GetTx(), false, noopNext)
	require.Errorf(t, err, "a tx from the blacklisted %s must be rejected", leakedRealValidators[0].name)
	require.Containsf(t, err.Error(), "is blacklisted",
		"rejection must come from the blacklist decorator specifically (got: %s)", err.Error())

	// --- 5. an ordinary, freshly funded (non-blacklisted) user can still
	// delegate to, and fully undelegate from, a validator whose operator
	// account is blacklisted — through the real multistaking message
	// handler (the one actually wired up for MsgDelegate/MsgUndelegate on
	// this chain: it locks the multistaking coin and mints the bond coin
	// before forwarding to x/staking), not bare x/staking directly. ---
	delPriv, err := osecp256k1.GenerateKey()
	require.NoError(t, err)
	delAddr := sdk.AccAddress(delPriv.PubKey().Address())

	acc := realioApp.AccountKeeper.NewAccountWithAddress(nextCtx, delAddr)
	realioApp.AccountKeeper.SetAccount(nextCtx, acc)

	fundAmount := sdkmath.NewInt(1_000_000_000_000_000_000)
	fundCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, fundAmount))
	require.NoError(t, realioApp.BankKeeper.MintCoins(nextCtx, minttypes.ModuleName, fundCoins))
	require.NoError(t, realioApp.BankKeeper.SendCoinsFromModuleToAccount(nextCtx, minttypes.ModuleName, delAddr, fundCoins))

	multiStakingMsgServer := multistakingkeeper.NewMsgServerImpl(realioApp.MultiStakingKeeper)

	delegateAmount := sdkmath.NewInt(500_000_000_000_000_000)
	_, err = multiStakingMsgServer.Delegate(nextCtx, stakingtypes.NewMsgDelegate(
		delAddr.String(), leakedRIO.val.String(), sdk.NewCoin(bondDenom, delegateAmount)))
	require.NoErrorf(t, err, "delegating to a validator whose operator is blacklisted must succeed")

	del, err := realioApp.StakingKeeper.GetDelegation(nextCtx, delAddr, leakedRIO.val)
	require.NoError(t, err)
	require.True(t, del.Shares.IsPositive())

	balanceBeforeUnbond := realioApp.BankKeeper.GetBalance(nextCtx, delAddr, bondDenom)
	_, err = realioApp.MultiStakingKeeper.Undelegate(nextCtx, stakingtypes.NewMsgUndelegate(
		delAddr.String(), leakedRIO.val.String(), sdk.NewCoin(bondDenom, delegateAmount)))
	require.NoErrorf(t, err, "undelegating from a validator whose operator is blacklisted must succeed")

	ubd, err := realioApp.StakingKeeper.GetUnbondingDelegation(nextCtx, delAddr, leakedRIO.val)
	require.NoError(t, err)
	require.NotEmpty(t, ubd.Entries, "undelegate must have queued a real unbonding entry")

	balanceAfterUnbond := realioApp.BankKeeper.GetBalance(nextCtx, delAddr, bondDenom)
	require.Truef(t, balanceAfterUnbond.Amount.LTE(balanceBeforeUnbond.Amount),
		"undelegate itself doesn't pay out immediately (funds move to the unbonding queue, not straight to the wallet)")

	// --- 6. real, PRE-EXISTING delegators of Teshy RIO (from this genesis,
	// not created by this test) — same split as reality: most of them
	// happen to be on leaked_addresses.json themselves (a separate,
	// unrelated leak), and for those undelegating is correctly still
	// blocked (their OWN account is blacklisted, nothing to do with the
	// validator's operator); the rest are clean and must be able to
	// undelegate their real, pre-fork delegation normally. ---
	cleanExistingDelegator, err := sdk.AccAddressFromBech32("realio1qzglcm3jcllgw2py4rd9e7l2c9txpsx79vys44")
	require.NoError(t, err)
	require.Falsef(t, realioApp.BlacklistKeeper.IsBlacklisted(nextCtx, cleanExistingDelegator),
		"fixture assumption broken: this delegator was expected to not be on the leaked list itself")

	preExistingDel, err := realioApp.StakingKeeper.GetDelegation(nextCtx, cleanExistingDelegator, leakedRIO.val)
	require.NoError(t, err)
	require.True(t, preExistingDel.Shares.IsPositive(), "fixture assumption broken: expected a real pre-existing delegation")

	partialUnbond := sdk.NewCoin(bondDenom, sdkmath.NewInt(1_000_000_000_000_000_000)) // 1 token worth, well under this delegator's real ~15,470 balance
	_, err = realioApp.MultiStakingKeeper.Undelegate(nextCtx, stakingtypes.NewMsgUndelegate(
		cleanExistingDelegator.String(), leakedRIO.val.String(), partialUnbond))
	require.NoErrorf(t, err, "a clean pre-existing delegator of the blacklisted-operator validator must still be able to undelegate")

	leakedExistingDelegator, err := sdk.AccAddressFromBech32("realio1qh3x0ety53eeq4slpmvajye5jf9zq5jqtre2t2")
	require.NoError(t, err)
	require.Truef(t, realioApp.BlacklistKeeper.IsBlacklisted(nextCtx, leakedExistingDelegator),
		"fixture assumption broken: this delegator was expected to be on the leaked list itself (a separate leak from the validator's)")

	undelegateMsg := stakingtypes.NewMsgUndelegate(leakedExistingDelegator.String(), leakedRIO.val.String(), partialUnbond)
	unsignedUndelegateBuilder := encodingConfig.TxConfig.NewTxBuilder()
	require.NoError(t, unsignedUndelegateBuilder.SetMsgs(undelegateMsg))
	_, err = blacklistDecorator.AnteHandle(nextCtx, unsignedUndelegateBuilder.GetTx(), false, noopNext)
	require.Errorf(t, err, "a pre-existing delegator whose OWN account is separately blacklisted must still be blocked from undelegating")
	require.Containsf(t, err.Error(), "is blacklisted", "got: %s", err.Error())

	// --- 7. the two unbonding delegations queued back in step 0b — one for
	// a delegator who is now blacklisted, one for a delegator who stayed
	// clean — must both mature and pay out normally through the real
	// EndBlocker, exactly as if nothing had happened. Maturity payout is an
	// automatic per-block process (x/staking's own EndBlocker dequeuing the
	// UBD queue), not a signed transaction, so it is never routed through
	// the ante chain at all; and BlacklistSendRestriction only ever checks
	// the SENDER of a transfer, never the recipient — deliberately, so the
	// chain's own automatic payouts to a blacklisted address are never
	// blocked (see x/blacklist/keeper/restrictions.go). Both properties
	// combined mean an unbonding that was already in flight before the
	// fork is completely unaffected by it, no matter which side of it
	// ends up blacklisted. ---
	balancesBeforeMaturity := make(map[string]sdk.Coin, 2)
	for _, delAddr := range []sdk.AccAddress{unbondingDelegatorLeaked, unbondingDelegatorClean} {
		balancesBeforeMaturity[delAddr.String()] = realioApp.BankKeeper.GetBalance(nextCtx, delAddr, bondDenom)
	}

	farFutureCtx := app.NewHeaderCtx(realioApp, initialHeight+2, proposerAddr, setupTime.Add(30*24*time.Hour))
	require.NotPanics(t, func() {
		_, err := realioApp.EndBlocker(farFutureCtx)
		require.NoError(t, err)
	})

	require.Truef(t, realioApp.BlacklistKeeper.IsBlacklisted(farFutureCtx, unbondingDelegatorLeaked),
		"sanity: this delegator must still be blacklisted at maturity time")

	for _, delAddr := range []sdk.AccAddress{unbondingDelegatorLeaked, unbondingDelegatorClean} {
		balAfter := realioApp.BankKeeper.GetBalance(farFutureCtx, delAddr, bondDenom)
		before := balancesBeforeMaturity[delAddr.String()]
		require.Truef(t, balAfter.Amount.GT(before.Amount),
			"unbonding delegation for %s must mature and pay out normally regardless of blacklist status (before=%s after=%s)",
			delAddr, before, balAfter)
	}

	remaining, err := realioApp.StakingKeeper.GetUnbondingDelegationsFromValidator(farFutureCtx, rioVal)
	require.NoError(t, err)
	require.Empty(t, remaining, "both unbonding delegations must be fully settled, none stuck")
}
