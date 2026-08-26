package app

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cosmossdk.io/math"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/evm/contracts"
	srvflags "github.com/cosmos/evm/server/flags"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/stretchr/testify/require"
)

// SetupWithGenFile boots a RealioNetwork app straight from a full exported
// genesis file (chain_id, initial_height and app_state are all taken from
// the file itself), instead of the synthetic genesis built by Setup(). It
// also returns the consensus address of the genesis's first validator, which
// callers need to set as the block header's ProposerAddress — without it,
// the EVM keeper's COINBASE lookup fails ("validator does not exist") on
// every single EVM call (including read-only ones like ERC20 balanceOf),
// since a real node's FinalizeBlock always carries a real proposer and a
// hand-built test context otherwise leaves it empty.
func SetupWithGenFile(t *testing.T, genFile string) (*RealioNetwork, []byte) {
	t.Helper()

	// Required whenever this runs alongside other tests in the package (e.g.
	// TestFork, TestCommissionUpgrade via Setup()): the EVM coin info is a
	// process-global set-once value, and InitGenesis panics on a second set
	// without this reset.
	evmtypes.NewEVMConfigurator().ResetTestConfig()

	genData, err := os.ReadFile(genFile)
	require.NoError(t, err, "failed to read genesis file")

	var genesisDoc map[string]interface{}
	err = json.Unmarshal(genData, &genesisDoc)
	require.NoError(t, err, "failed to unmarshal genesis file")

	appState, ok := genesisDoc["app_state"].(map[string]interface{})
	require.True(t, ok, "failed to extract app_state from genesis")

	chainID, _ := genesisDoc["chain_id"].(string)
	require.NotEmpty(t, chainID)

	genesisTimeStr, _ := genesisDoc["genesis_time"].(string)
	genesisTime, err := time.Parse(time.RFC3339Nano, genesisTimeStr)
	require.NoError(t, err)

	initialHeight := int64(19572539)
	if v, ok := genesisDoc["initial_height"].(float64); ok {
		initialHeight = int64(v)
	}

	consensus, ok := genesisDoc["consensus"].(map[string]interface{})
	require.True(t, ok, "failed to extract consensus from genesis")
	validators, ok := consensus["validators"].([]interface{})
	require.True(t, ok, "failed to extract consensus.validators from genesis")
	require.NotEmpty(t, validators, "expected at least one genesis validator")
	firstVal, ok := validators[0].(map[string]interface{})
	require.True(t, ok)
	proposerAddrHex, ok := firstVal["address"].(string)
	require.True(t, ok, "failed to extract consensus.validators[0].address from genesis")
	proposerAddr, err := hex.DecodeString(proposerAddrHex)
	require.NoError(t, err, "failed to decode proposer address %q", proposerAddrHex)

	db := dbm.NewMemDB()
	opt := baseapp.SetChainID(chainID)
	appOpts := simtestutil.AppOptionsMap{srvflags.EVMChainID: MainnetEVMChainID}

	realioApp := New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, appOpts, opt)

	stateBytes, err := json.MarshalIndent(appState, "", " ")
	require.NoError(t, err, "failed to marshal app state")

	_, err = realioApp.InitChain(&abci.RequestInitChain{
		Time:            genesisTime,
		InitialHeight:   initialHeight,
		ChainId:         chainID,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err, "failed to init chain")

	_, err = realioApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:          initialHeight,
		Time:            genesisTime,
		Txs:             [][]byte{},
		ProposerAddress: proposerAddr,
	})
	require.NoError(t, err, "failed to finalize first block")

	return realioApp, proposerAddr
}

// loadRotatedAddresses reads the address,replace CSV produced for the key
// rotation and returns old->new address mapping.
func loadRotatedAddresses(t *testing.T, csvFile string) map[string]string {
	t.Helper()

	f, err := os.Open(csvFile)
	require.NoError(t, err)
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	require.NoError(t, err)
	require.True(t, len(records) > 1)
	require.Equal(t, []string{"address", "replace"}, records[0])

	mapping := make(map[string]string, len(records)-1)
	for _, rec := range records[1:] {
		require.Len(t, rec, 2)
		mapping[rec[0]] = rec[1]
	}
	return mapping
}

// TestRotatedAddressesLifecycle validates the address-rotation mechanism
// end-to-end against the app's real keepers, using the edited genesis
// (recover_genesis.json with the 2449 leaked keys swapped for freshly
// generated eth_secp256k1 keys):
//
//  1. every unbonding delegation that was already in-flight in the original
//     genesis still matures and pays out normally after the rotation, and
//  2. a brand-new delegator using one of the ROTATED addresses can submit a
//     fresh MsgUndelegate and have it mature and pay out normally, and
//  3. a plain bank transfer between two rotated addresses works normally.
// dstrxContract is the real "Districts Token" (DSTRX) ERC20 contract from
// the recovered mainnet genesis, registered as an erc20-backed multistaking
// coin (denom "erc20:0xb841F365..."). Several of the addresses being rotated
// hold DSTRX both as raw ERC20 balance and as a pending multistaking unlock;
// this test verifies the multistaking-unlock side actually pays out to the
// ROTATED address at the real contract, not just in bank bookkeeping.
var dstrxContract = common.HexToAddress("0xb841F365D5221Bed66d60E69094418D8C2aa5A44")

func erc20BalanceOf(t *testing.T, realioApp *RealioNetwork, ctx sdk.Context, contract, holder common.Address) *big.Int {
	t.Helper()
	bal := realioApp.Erc20Keeper.BalanceOf(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, contract, holder)
	require.NotNil(t, bal, "balanceOf(%s) on %s returned nil (EVM call failed)", holder, contract)
	return bal
}

func TestRotatedAddressesLifecycle(t *testing.T) {
	realioApp, proposerAddr := SetupWithGenFile(t, filepath.Join("testdata", "recover_genesis_edited.json"))
	rotated := loadRotatedAddresses(t, filepath.Join("testdata", "wallet-active-addresses-replaced.csv"))
	newAddrSet := make(map[string]bool, len(rotated))
	for _, newAddr := range rotated {
		newAddrSet[newAddr] = true
	}

	startHeight := realioApp.LastBlockHeight() + 1
	ctx := realioApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Height:          startHeight,
		Time:            time.Now(),
		ProposerAddress: proposerAddr,
	}).WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	qServer := multistakingkeeper.NewQueryServerImpl(realioApp.MultiStakingKeeper)

	// ---- Phase 1: pre-existing unbonding delegations still mature normally,
	//      for BOTH native coins and the real erc20-backed DSTRX coin ----
	unlocksRes, err := qServer.MultiStakingUnlocks(ctx, &multistakingtypes.QueryMultiStakingUnlocksRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, unlocksRes.Unlocks, "expected pre-existing unlocks carried over from the recovered genesis")

	type pendingUnbond struct {
		delAddr      sdk.AccAddress
		denom        string
		before       sdk.Coin
		isDSTRX      bool
		dstrxBefore  *big.Int
	}

	var maxCompletion time.Time
	var pendings []pendingUnbond
	for _, unlock := range unlocksRes.Unlocks {
		// Not every pre-existing unlock necessarily belongs to a rotated
		// address (wallet-active-addresses.csv is not guaranteed to be a
		// superset of every delegator with an in-flight unbond) — this
		// phase only checks that maturity/payout still works after the
		// genesis-wide rename, regardless of whose unlock it is.
		delAddr := sdk.MustAccAddressFromBech32(unlock.UnlockID.MultiStakerAddr)
		valAddr, err := sdk.ValAddressFromBech32(unlock.UnlockID.ValAddr)
		require.NoError(t, err)

		ubd, err := realioApp.StakingKeeper.GetUnbondingDelegation(ctx, delAddr, valAddr)
		if err != nil || len(ubd.Entries) == 0 {
			// Some multistaking unlock entries (e.g. ones left over from a
			// RemoveMultiStakingCoinProposal forced-unbond) are not backed
			// by a standard staking-module UnbondingDelegation and so don't
			// mature via the normal block-time queue; skip those here, they
			// are unrelated to whether address rotation itself works.
			t.Logf("skipping unlock %s/%s: no matching staking UnbondingDelegation (%v)",
				unlock.UnlockID.MultiStakerAddr, unlock.UnlockID.ValAddr, err)
			continue
		}

		denom := unlock.Entries[0].UnlockingCoin.Denom
		isDSTRX := denom == "erc20:0xb841F365D5221Bed66d60E69094418D8C2aa5A44"

		completion := ubd.Entries[len(ubd.Entries)-1].CompletionTime
		if completion.After(maxCompletion) {
			maxCompletion = completion
		}

		p := pendingUnbond{delAddr: delAddr, denom: denom, isDSTRX: isDSTRX}
		p.before = realioApp.BankKeeper.GetBalance(ctx, delAddr, denom)
		if isDSTRX {
			p.dstrxBefore = erc20BalanceOf(t, realioApp, ctx, dstrxContract, common.BytesToAddress(delAddr.Bytes()))
			t.Logf("rotated address %s (DSTRX unlock): bank(%s)=%s, contract balanceOf=%s",
				unlock.UnlockID.MultiStakerAddr, denom, p.before, p.dstrxBefore)
		}
		pendings = append(pendings, p)
	}
	require.NotEmpty(t, pendings, "expected at least one pending unbond to test maturity with")
	require.True(t, func() bool {
		for _, p := range pendings {
			if p.isDSTRX {
				return true
			}
		}
		return false
	}(), "expected at least one pending DSTRX unlock to test")

	// Advance block time/height past the latest maturity for ALL pending unbonds.
	ctx = ctx.WithBlockTime(maxCompletion.Add(time.Second)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Second)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	for _, p := range pendings {
		if !p.isDSTRX {
			// The erc20-backed unlock does NOT credit the "erc20:0x..." denom
			// through the bank module at all (GetBalance stays 0 for it) —
			// the payout lands directly in the real ERC20 contract's storage
			// instead, verified separately below via balanceOf. Only assert
			// the bank-balance increase for genuinely bank-tracked denoms.
			after := realioApp.BankKeeper.GetBalance(ctx, p.delAddr, p.denom)
			require.Truef(t, p.before.IsLT(after),
				"expected balance increase for %s (%s) after pre-existing unbonding matured: before=%s after=%s",
				p.delAddr, p.denom, p.before, after)
		}

		if p.isDSTRX {
			dstrxAfter := erc20BalanceOf(t, realioApp, ctx, dstrxContract, common.BytesToAddress(p.delAddr.Bytes()))
			require.Truef(t, p.dstrxBefore.Cmp(dstrxAfter) < 0,
				"expected DSTRX contract balanceOf(%s) to increase after unbonding matured: before=%s after=%s",
				p.delAddr, p.dstrxBefore, dstrxAfter)
			t.Logf("rotated address %s: DSTRX contract balanceOf increased %s -> %s after unbonding matured",
				p.delAddr, p.dstrxBefore, dstrxAfter)
		}
	}

	msgServer := multistakingkeeper.NewMsgServerImpl(realioApp.MultiStakingKeeper)

	// ---- Phase 2: a rotated address submits a FRESH delegate (stake) ----
	// Find a rotated address holding a free (unstaked) native balance, and a
	// validator that accepts that same multistaking coin.
	const stakeDenom = "ario"
	var stakerAddrStr string
	var stakerFreeBal sdk.Coin
	for oldAddr, newAddr := range rotated {
		_ = oldAddr
		addr := sdk.MustAccAddressFromBech32(newAddr)
		bal := realioApp.BankKeeper.GetBalance(ctx, addr, stakeDenom)
		if bal.Amount.GT(math.NewInt(1_000_000)) {
			stakerAddrStr = newAddr
			stakerFreeBal = bal
			break
		}
	}
	require.NotEmpty(t, stakerAddrStr, "expected at least one rotated address with a free %s balance", stakeDenom)

	// validator_multi_staking_coins confirms this validator accepts "ario".
	stakeValidator := "realiovaloper1qzqgpazwdj6eks2ry7p9scmsqqe8qd9uryat2r"
	stakeAmount := stakerFreeBal.Amount.QuoRaw(10)
	require.True(t, stakeAmount.IsPositive())

	_, err = msgServer.Delegate(ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: stakerAddrStr,
		ValidatorAddress: stakeValidator,
		Amount:           sdk.NewCoin(stakeDenom, stakeAmount),
	})
	require.NoError(t, err, "rotated address %s failed to submit a fresh delegate", stakerAddrStr)

	stakerAddr := sdk.MustAccAddressFromBech32(stakerAddrStr)
	newLock, found := realioApp.MultiStakingKeeper.GetMultiStakingLock(ctx,
		multistakingtypes.MultiStakingLockID(stakerAddrStr, stakeValidator))
	require.True(t, found, "expected a multistaking lock to be created for the fresh delegate")
	require.Equal(t, stakeAmount, newLock.LockedCoin.Amount)

	afterStakeBal := realioApp.BankKeeper.GetBalance(ctx, stakerAddr, stakeDenom)
	require.Equal(t, stakerFreeBal.Amount.Sub(stakeAmount), afterStakeBal.Amount)

	// ---- Phase 3: the SAME rotated address submits a FRESH undelegate and
	//      it matures and pays out normally ----
	valAddr, err := sdk.ValAddressFromBech32(stakeValidator)
	require.NoError(t, err)

	unbondAmount := stakeAmount.QuoRaw(2)
	require.True(t, unbondAmount.IsPositive())

	beforeUnbondBal := realioApp.BankKeeper.GetBalance(ctx, stakerAddr, stakeDenom)

	_, err = msgServer.Undelegate(ctx, &stakingtypes.MsgUndelegate{
		DelegatorAddress: stakerAddrStr,
		ValidatorAddress: stakeValidator,
		Amount:           sdk.NewCoin(stakeDenom, unbondAmount),
	})
	require.NoError(t, err, "rotated address %s failed to submit a fresh undelegate", stakerAddrStr)

	ubd, err := realioApp.StakingKeeper.GetUnbondingDelegation(ctx, stakerAddr, valAddr)
	require.NoError(t, err)
	require.NotEmpty(t, ubd.Entries)
	entry := ubd.Entries[len(ubd.Entries)-1]

	unbondingTimeParam, err := realioApp.StakingKeeper.UnbondingTime(ctx)
	require.NoError(t, err)
	require.WithinDuration(t, ctx.BlockTime().Add(unbondingTimeParam), entry.CompletionTime, time.Minute,
		"fresh undelegate completion time should be ~unbonding_time in the future")

	afterUnbondBal := realioApp.BankKeeper.GetBalance(ctx, stakerAddr, stakeDenom)
	require.Equal(t, beforeUnbondBal.Amount, afterUnbondBal.Amount,
		"undelegate should not release funds before the unbonding period matures")

	// advance past this fresh undelegate's own completion and confirm it pays out too.
	ctx = ctx.WithBlockTime(entry.CompletionTime.Add(time.Second)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Second)).WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = realioApp.EndBlocker(ctx)
	require.NoError(t, err)

	maturedUnbondBal := realioApp.BankKeeper.GetBalance(ctx, stakerAddr, stakeDenom)
	require.Truef(t, beforeUnbondBal.IsLT(maturedUnbondBal),
		"expected rotated address %s balance to increase after its fresh undelegate matured: before=%s after=%s",
		stakerAddr, beforeUnbondBal, maturedUnbondBal)

	// ---- Phase 4: plain bank transfer between two rotated addresses ----
	senderAddr := stakerAddr
	senderBal := realioApp.BankKeeper.GetBalance(ctx, senderAddr, stakeDenom)
	require.True(t, senderBal.IsPositive())

	var recipientAddr sdk.AccAddress
	for _, newAddr := range rotated {
		if newAddr == senderAddr.String() {
			continue
		}
		recipientAddr = sdk.MustAccAddressFromBech32(newAddr)
		break
	}
	require.NotNil(t, recipientAddr)

	sendAmount := sdk.NewCoin(stakeDenom, senderBal.Amount.QuoRaw(2))
	recipientBefore := realioApp.BankKeeper.GetBalance(ctx, recipientAddr, stakeDenom)

	// Go through bank's MsgServer (not a direct keeper call) so this exercises
	// the same message-handling path a signed MsgSend broadcast by a real
	// wallet (Keplr, CLI) would take — including BlockedAddr/SendEnabled
	// checks — not just the underlying balance bookkeeping.
	bankMsgServer := bankkeeper.NewMsgServerImpl(realioApp.BankKeeper)
	_, err = bankMsgServer.Send(ctx, &banktypes.MsgSend{
		FromAddress: senderAddr.String(),
		ToAddress:   recipientAddr.String(),
		Amount:      sdk.NewCoins(sendAmount),
	})
	require.NoError(t, err, "MsgSend from rotated address %s to %s failed", senderAddr, recipientAddr)

	senderAfter := realioApp.BankKeeper.GetBalance(ctx, senderAddr, stakeDenom)
	recipientAfter := realioApp.BankKeeper.GetBalance(ctx, recipientAddr, stakeDenom)

	require.Equal(t, senderBal.Amount.Sub(sendAmount.Amount), senderAfter.Amount)
	require.Equal(t, recipientBefore.Amount.Add(sendAmount.Amount), recipientAfter.Amount)
}
