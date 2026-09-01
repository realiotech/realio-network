package app

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	"github.com/realiotech/realio-network/testutil"
)

// realGenesisPath is the actual pre-incident mainnet genesis export sitting
// at the repo root — the same file the leaked-validator rotation is meant
// to run against for real. Committed nowhere; this test is skipped if it's
// not present (e.g. on CI, or a fresh checkout).
const realGenesisPath = "../recover_genesis.json"

// consensusValidatorEntry mirrors the top-level consensus.validators[] shape
// in the exported genesis (hex address + base64 ed25519 pubkey + power),
// just enough to pick a ProposerAddress for FinalizeBlock.
type consensusValidatorEntry struct {
	Address string `json:"address"`
}

// SetupWithRealGenesis boots a full app against the real genesis export,
// finalizing the first block at time.Now() — the same way a real node
// resuming this chain today would. Any unbonding delegation whose
// completion_time has already elapsed by "now" matures and pays out right
// here, as part of this very first block's EndBlock, before the caller gets
// a chance to inspect it as "pending". That's expected real-world behavior,
// not a bug: overdue unbondings are settled at genesis load, and rotation
// only needs to deal with whatever is still in flight afterwards.
// Skips the test (rather than failing) if the file isn't present.
func SetupWithRealGenesis(t *testing.T) (realioApp *RealioNetwork, chainID string, initialHeight int64, proposerAddr []byte, blockTime time.Time) {
	t.Helper()

	raw, err := os.ReadFile(realGenesisPath)
	if err != nil {
		t.Skipf("real genesis fixture not present at %s, skipping: %v", realGenesisPath, err)
		return nil, "", 0, nil, time.Time{}
	}

	// The EVM module keeps its coin-info config in a process-global
	// sync.Once (github.com/cosmos/evm/x/vm.SetGlobalConfigVariables), same
	// as Setup() in test_helpers.go deals with — must reset it before this
	// InitChain, or InitGenesis panics if any other test in this binary
	// already initialized it first.
	evmtypes.NewEVMConfigurator().ResetTestConfig()

	var doc struct {
		ChainID       string                     `json:"chain_id"`
		InitialHeight int64                      `json:"initial_height"`
		AppState      map[string]json.RawMessage `json:"app_state"`
		Consensus     struct {
			Validators []consensusValidatorEntry `json:"validators"`
		} `json:"consensus"`
	}
	require.NoError(t, json.Unmarshal(raw, &doc))
	require.NotEmpty(t, doc.Consensus.Validators)
	require.NotZero(t, doc.InitialHeight)

	proposerAddr, err = hex.DecodeString(doc.Consensus.Validators[0].Address)
	require.NoError(t, err)

	appStateBytes, err := json.Marshal(doc.AppState)
	require.NoError(t, err)

	db := dbm.NewMemDB()
	realioApp = New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, simtestutil.EmptyAppOptions{},
		baseapp.SetChainID(doc.ChainID))

	_, err = realioApp.InitChain(&abci.RequestInitChain{
		ChainId:         doc.ChainID,
		InitialHeight:   doc.InitialHeight,
		ConsensusParams: DefaultConsensusParams,
		AppStateBytes:   appStateBytes,
	})
	require.NoError(t, err)

	blockTime = time.Now()

	// This genesis's initial_height is the real chain's actual halt height —
	// which may legitimately equal BlacklistForkHeight's current production
	// value (it does today: both are 19573266, since this genesis IS the
	// real halt-block export). Neutralize every height-triggered fork for
	// the duration of this one FinalizeBlock call, so none of them fire
	// here by coincidence before a test gets a chance to control it
	// explicitly. Callers that want to exercise a fork opt in afterwards by
	// setting the relevant height themselves and driving
	// BeginBlocker/EndBlocker directly, same as the rest of this file does.
	origBlacklistHeight, origRotationHeight := BlacklistForkHeight, ValidatorRotationHeight
	BlacklistForkHeight, ValidatorRotationHeight = -1, -1
	defer func() { BlacklistForkHeight, ValidatorRotationHeight = origBlacklistHeight, origRotationHeight }()

	// Deliberately not calling Commit() here: doing so tears down
	// finalizeBlockState, and BaseApp.NewContextLegacy(false, ...) — used
	// throughout this test to build ad-hoc contexts for direct
	// BeginBlocker/EndBlocker calls, the same pattern the rest of this
	// package's tests use against Setup() — reads from exactly that state.
	// Setup() in test_helpers.go follows the same convention.
	_, err = realioApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:          doc.InitialHeight,
		ProposerAddress: proposerAddr,
		Time:            blockTime,
	})
	require.NoError(t, err)

	return realioApp, doc.ChainID, doc.InitialHeight, proposerAddr, blockTime
}

func newHeaderCtx(realioApp *RealioNetwork, height int64, proposerAddr []byte, blockTime time.Time) sdk.Context {
	return realioApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Height:          height,
		ProposerAddress: proposerAddr,
		Time:            blockTime,
	}).WithBlockGasMeter(storetypes.NewInfiniteGasMeter())
}

// rotationTarget is one of the two real leaked validators, plus a real
// genesis delegator (with a real MultiStakingLock on that validator) used to
// create a genuinely in-flight unbonding delegation right before rotation —
// so the test actually exercises the migration+time-fly path, rather than
// relying on this genesis's one pre-existing UBD, which is already overdue
// relative to "now" and settles during genesis load itself (see
// SetupWithRealGenesis).
type rotationTarget struct {
	oldOperator      string
	multiStakerAddr  string
	coinDenom        string
	undelegateAmount math.Int
}

// TestRotateValidatorsAgainstRealGenesis is the end-to-end test against the
// real pre-incident genesis export: rotates both leaked validators
// (realiovaloper18a32el4maw3pqr8xh3yrl9ja4lejs265a5nxtm "Teshy 103k Bonded
// RIO" and realiovaloper13jrrtkfuuvzdak6zxmr95hek9c228ug50sdsvs "Teshy 136k
// Self Bonded RST") to freshly generated identities, then:
//  0. confirms the one real pre-existing unbonding delegation in this
//     genesis (already overdue relative to "now") settled correctly at
//     genesis load, before rotation ever runs;
//  1. creates a fresh, genuinely in-flight unbonding delegation on each old
//     validator (via the real multistaking message path, from a real
//     genesis delegator's real lock) immediately before rotation, then
//     time-flies past its maturity through the REAL app.EndBlocker (not a
//     direct keeper call) — proving the multistaking module's own EndBlock
//     (which has its own UBD-queue interaction, layered on top of
//     x/staking's) doesn't choke on funds mid-migration;
//  2. has users of the NEW validators fully undelegate afterwards, proving
//     the migrated identity is completely usable going forward, not just
//     "doesn't crash".
func TestRotateValidatorsAgainstRealGenesis(t *testing.T) {
	realioApp, _, initialHeight, proposerAddr, setupTime := SetupWithRealGenesis(t)

	targets := []rotationTarget{
		{
			oldOperator:      "realiovaloper18a32el4maw3pqr8xh3yrl9ja4lejs265a5nxtm", // Teshy 103k Bonded RIO
			multiStakerAddr:  "realio1d2rjp2kxc7md7q9xmjmslmludexv5lvk338k6j",
			coinDenom:        "ario",
			undelegateAmount: math.NewIntFromUint64(1_000).MulRaw(1_000_000_000_000_000_000),
		},
		{
			oldOperator:      "realiovaloper13jrrtkfuuvzdak6zxmr95hek9c228ug50sdsvs", // Teshy 136k Self Bonded RST
			multiStakerAddr:  "realio13nxg9kc48lrvzm5ms5xanup5hnexzppla4dfnh",
			coinDenom:        "arst",
			undelegateAmount: math.NewIntFromUint64(1_000).MulRaw(1_000_000_000_000_000_000),
		},
	}

	baseCtx := newHeaderCtx(realioApp, initialHeight, proposerAddr, setupTime)

	// --- 0. the real pre-existing UBD in this genesis (for validator 1,
	// completion_time 2026-08-27, already in the past relative to setupTime)
	// must already be gone — settled during SetupWithRealGenesis's own
	// FinalizeBlock, exactly like a real node resuming the chain today. ---
	overdueDelegator, err := sdk.AccAddressFromBech32("realio1ysfea6fmkk3quvqval7k85a2fxxfmsc2ldgpkh")
	require.NoError(t, err)
	overdueValAddr, err := sdk.ValAddressFromBech32(targets[0].oldOperator)
	require.NoError(t, err)
	overdueBefore, err := realioApp.StakingKeeper.GetUnbondingDelegationsFromValidator(baseCtx, overdueValAddr)
	require.NoError(t, err)
	require.Empty(t, overdueBefore, "the genesis's one pre-existing (overdue) UBD must have already settled at genesis load")
	overdueBalance := realioApp.BankKeeper.GetBalance(baseCtx, overdueDelegator, targets[0].coinDenom)
	require.Truef(t, overdueBalance.IsPositive(), "delegator of the overdue genesis UBD must have received its %s payout at genesis load", targets[0].coinDenom)

	origRotations, origHeight := validatorRotations, ValidatorRotationHeight
	t.Cleanup(func() { validatorRotations, ValidatorRotationHeight = origRotations, origHeight })

	rotationHeight := initialHeight + 1
	ValidatorRotationHeight = rotationHeight

	newValAddrs := make(map[string]sdk.ValAddress, len(targets))
	var rotations []struct {
		OldOperator      string
		NewOperator      string
		NewConsPubKeyB64 string
	}
	for _, tgt := range targets {
		newValAddr := sdk.ValAddress(testutil.GenAddress())
		newValAddrs[tgt.oldOperator] = newValAddr
		newConsPriv := ed25519.GenPrivKey()
		rotations = append(rotations, struct {
			OldOperator      string
			NewOperator      string
			NewConsPubKeyB64 string
		}{
			OldOperator:      tgt.oldOperator,
			NewOperator:      newValAddr.String(),
			NewConsPubKeyB64: pubKeyB64(t, newConsPriv.PubKey().(*ed25519.PubKey)),
		})
	}
	validatorRotations = rotations

	// --- 1a. create a genuinely in-flight UBD on each OLD validator right
	// before rotation, via the real multistaking Undelegate path (real
	// lock -> real unlock entry -> real staking UnbondingDelegation with a
	// future completion_time). ---
	for _, tgt := range targets {
		oldValAddr, err := sdk.ValAddressFromBech32(tgt.oldOperator)
		require.NoError(t, err)

		_, err = realioApp.MultiStakingKeeper.Undelegate(baseCtx, &stakingtypes.MsgUndelegate{
			DelegatorAddress: tgt.multiStakerAddr,
			ValidatorAddress: tgt.oldOperator,
			Amount:           sdk.NewCoin(tgt.coinDenom, tgt.undelegateAmount),
		})
		require.NoErrorf(t, err, "creating a fresh in-flight UBD on %s must succeed", tgt.oldOperator)

		pending, err := realioApp.StakingKeeper.GetUnbondingDelegationsFromValidator(baseCtx, oldValAddr)
		require.NoError(t, err)
		require.Lenf(t, pending, 1, "expected exactly the freshly created in-flight UBD on %s", tgt.oldOperator)
	}

	ctx := newHeaderCtx(realioApp, rotationHeight, proposerAddr, baseCtx.BlockTime())
	_, err = realioApp.BeginBlocker(ctx)
	require.NoError(t, err)

	for _, tgt := range targets {
		newValAddr := newValAddrs[tgt.oldOperator]
		_, err := realioApp.StakingKeeper.GetValidator(ctx, newValAddr)
		require.NoErrorf(t, err, "new validator for %s must exist after rotation", tgt.oldOperator)

		oldValAddr, err := sdk.ValAddressFromBech32(tgt.oldOperator)
		require.NoError(t, err)
		stillOnOld, err := realioApp.StakingKeeper.GetUnbondingDelegationsFromValidator(ctx, oldValAddr)
		require.NoError(t, err)
		require.Empty(t, stillOnOld, "no unbonding delegation should remain addressable under the old operator %s", tgt.oldOperator)
	}

	// --- 1b. time-fly the freshly created in-flight UBDs past maturity
	// through the REAL EndBlocker (module manager EndBlock, including
	// multistaking's own layered logic) ---
	type pendingCheck struct {
		delegator sdk.AccAddress
		denom     string
	}
	var checks []pendingCheck
	for _, tgt := range targets {
		newValAddr := newValAddrs[tgt.oldOperator]
		ubds, err := realioApp.StakingKeeper.GetUnbondingDelegationsFromValidator(ctx, newValAddr)
		require.NoError(t, err)
		require.Lenf(t, ubds, 1, "the in-flight UBD created before rotation must have migrated onto %s", newValAddr)
		for _, ubd := range ubds {
			delAddr, err := sdk.AccAddressFromBech32(ubd.DelegatorAddress)
			require.NoError(t, err)
			checks = append(checks, pendingCheck{delAddr, tgt.coinDenom})
		}
	}
	require.NotEmpty(t, checks)

	balancesBefore := make(map[string]sdk.Coin, len(checks))
	for _, c := range checks {
		balancesBefore[c.delegator.String()+c.denom] = realioApp.BankKeeper.GetBalance(ctx, c.delegator, c.denom)
	}

	// Advance well past every entry's completion_time (unbonding_time in
	// this genesis is 7 days) and run the REAL EndBlocker.
	farFuture := ctx.BlockTime().Add(30 * 24 * time.Hour)
	matureCtx := newHeaderCtx(realioApp, rotationHeight+1, proposerAddr, farFuture)
	require.NotPanics(t, func() {
		_, err := realioApp.EndBlocker(matureCtx)
		require.NoError(t, err)
	})

	for _, c := range checks {
		balAfter := realioApp.BankKeeper.GetBalance(matureCtx, c.delegator, c.denom)
		before := balancesBefore[c.delegator.String()+c.denom]
		require.Truef(t, balAfter.Amount.GT(before.Amount),
			"delegator %s must have received %s funds after the migrated unbonding matured (before=%s after=%s)",
			c.delegator, c.denom, before, balAfter)
	}

	for _, tgt := range targets {
		newValAddr := newValAddrs[tgt.oldOperator]
		remaining, err := realioApp.StakingKeeper.GetUnbondingDelegationsFromValidator(matureCtx, newValAddr)
		require.NoError(t, err)
		require.Empty(t, remaining, "migrated unbonding delegations for %s must be fully paid out, none stuck", tgt.oldOperator)
	}

	// --- 2. a user of each NEW validator fully undelegates afterwards ---
	for _, tgt := range targets {
		newValAddr := newValAddrs[tgt.oldOperator]

		delegations, err := realioApp.StakingKeeper.GetValidatorDelegations(matureCtx, newValAddr)
		require.NoError(t, err)
		require.NotEmpty(t, delegations, "expected at least one migrated real delegation on %s", newValAddr)

		del, err := sdk.AccAddressFromBech32(delegations[0].DelegatorAddress)
		require.NoError(t, err)

		delBalanceBefore := realioApp.BankKeeper.GetBalance(matureCtx, del, tgt.coinDenom)

		lockID := multistakingtypes.MultiStakingLockID(delegations[0].DelegatorAddress, newValAddr.String())
		lock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(matureCtx, lockID)
		require.Truef(t, found, "migrated self-delegation on %s must have a real MultiStakingLock carried over", newValAddr)

		_, uErr := realioApp.MultiStakingKeeper.Undelegate(matureCtx, &stakingtypes.MsgUndelegate{
			DelegatorAddress: delegations[0].DelegatorAddress,
			ValidatorAddress: newValAddr.String(),
			Amount:           sdk.NewCoin(lock.LockedCoin.Denom, lock.LockedCoin.Amount),
		})
		require.NoErrorf(t, uErr, "user of migrated validator %s must be able to fully undelegate", newValAddr)

		fullUbd, err := realioApp.StakingKeeper.GetUnbondingDelegation(matureCtx, del, newValAddr)
		require.NoError(t, err)
		require.NotEmpty(t, fullUbd.Entries)
		unbondTime := fullUbd.Entries[len(fullUbd.Entries)-1].CompletionTime.Add(time.Second)

		afterFullUnbond := newHeaderCtx(realioApp, rotationHeight+2, proposerAddr, unbondTime.Add(time.Second))
		require.NotPanics(t, func() {
			_, err := realioApp.EndBlocker(afterFullUnbond)
			require.NoError(t, err)
		})

		delBalanceAfter := realioApp.BankKeeper.GetBalance(afterFullUnbond, del, tgt.coinDenom)
		require.Truef(t, delBalanceAfter.Amount.GT(delBalanceBefore.Amount),
			"delegator %s must receive funds after fully undelegating from migrated validator %s (before=%s after=%s)",
			del, newValAddr, delBalanceBefore, delBalanceAfter)

		_, err = realioApp.StakingKeeper.GetDelegation(afterFullUnbond, del, newValAddr)
		require.Error(t, err, "delegator %s should have no delegation left on %s after fully undelegating", del, newValAddr)
	}
}
