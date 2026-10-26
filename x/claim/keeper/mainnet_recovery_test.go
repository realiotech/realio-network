package keeper_test

import (
	"encoding/csv"
	"encoding/json"
	"math/big"
	"os"
	"sort"
	"strings"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"

	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	erc20types "github.com/cosmos/evm/x/erc20/types"
	"github.com/cosmos/evm/x/vm/statedb"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/realiotech/realio-network/testutil"
	blacklisttypes "github.com/realiotech/realio-network/x/blacklist/types"
	"github.com/realiotech/realio-network/x/claim/keeper"
	"github.com/realiotech/realio-network/x/claim/types"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

const (
	// attackerAddress is the one blacklisted address that is the
	// attacker's own rather than a victim's. It gets no recovery wallet,
	// no link, and no claim -- the module must leave it frozen exactly as
	// the fork left it.
	//
	// Nothing in this repo marks it: leaked_addresses.json is a flat list
	// of 33,606 entries, the fork runbook names no one, and the export's
	// authz grants were already revoked. It is identified here by its
	// profile in the mainnet state -- it holds roughly four times the RIO
	// of any other blacklisted address and has nothing staked at all --
	// and confirmed out of band. Changing it is a one-line edit.
	attackerAddress = "realio1uzkdrfnjv53rt0cf4ltszffpd7mvkpd2cv794j"

	// dstrxContract is the only ERC-20 registered with x/erc20 on real
	// mainnet (Districts Token, DSTRX), and the only one deployed with
	// real bytecode rather than as a native-coin precompile -- so it is
	// the only token on the chain MsgClaimTokens can actually sweep.
	dstrxContract = "0xb841F365D5221Bed66d60E69094418D8C2aa5A44"

	// attackerInflowPath is the incident's own accounting of what each
	// address lost: how much RIO it sent to the attacker. A genesis
	// snapshot cannot supply this -- it records final balances, not the
	// transfers behind them -- so the figures come from transaction
	// analysis done off-chain.
	//
	// Their sum is exactly the attacker's balance, to the unit, which is
	// what makes a full recovery possible and is asserted below.
	attackerInflowPath = "../testdata/attacker-inflow-wallets-balances.csv"

	// dstrxBalancesSlot is where DSTRX keeps its _balances mapping, i.e.
	// the contract is laid out the way OpenZeppelin's ERC20 is: _balances
	// 0, _allowances 1, _totalSupply 2, _name 3, _symbol 4. Confirmed
	// against the export's own slots 2-4, which hold the supply and the
	// strings "Districts Token" and "DSTRX".
	//
	// Only the fixture reads storage directly. The claim path itself never
	// does -- it calls balanceOf through the EVM like any other caller.
	dstrxBalancesSlot = 0
)

// ---------------------------------------------------------------------------
// Seeding real mainnet state
// ---------------------------------------------------------------------------

// seedMainnetBlacklist puts every address the fork froze onto the test
// chain's blacklist. The whole list goes in, not just the addresses a test
// touches, so that nothing here can quietly pass by operating on an
// address that was never actually frozen.
func (suite *KeeperTestSuite) seedMainnetBlacklist(appState map[string]json.RawMessage) []string {
	var gs blacklisttypes.GenesisState
	suite.Require().NoError(json.Unmarshal(appState["blacklist"], &gs))
	suite.Require().NotEmpty(gs.Addresses)

	for _, a := range gs.Addresses {
		addr, err := sdk.AccAddressFromBech32(a)
		suite.Require().NoError(err)
		suite.Require().NoError(suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, addr))
	}
	return gs.Addresses
}

// seedMainnetERC20Contract installs the real DSTRX contract -- its
// deployed bytecode and all of its storage, holder balances included --
// into the test EVM, and registers the token pair x/erc20 has for it on
// mainnet. Everything afterwards reads those balances by calling the
// contract, so the sweep runs against real holdings and real bytecode
// rather than a fixture token minted for the occasion.
func (suite *KeeperTestSuite) seedMainnetERC20Contract(appState map[string]json.RawMessage) common.Address {
	var evmGenesis evmtypes.GenesisState
	suite.Require().NoError(suite.app.AppCodec().UnmarshalJSON(appState["evm"], &evmGenesis))

	contract := common.HexToAddress(dstrxContract)
	var found *evmtypes.GenesisAccount
	for i := range evmGenesis.Accounts {
		if common.HexToAddress(evmGenesis.Accounts[i].Address) == contract {
			found = &evmGenesis.Accounts[i]
			break
		}
	}
	suite.Require().NotNil(found, "expected %s in the mainnet evm genesis", dstrxContract)

	code := common.Hex2Bytes(found.Code)
	suite.Require().NotEmpty(code)
	codeHash := crypto.Keccak256Hash(code)

	suite.app.EvmKeeper.SetCode(suite.ctx, codeHash.Bytes(), code)
	suite.Require().NoError(suite.app.EvmKeeper.SetAccount(suite.ctx, contract, statedb.Account{
		Nonce:    1,
		Balance:  uint256.NewInt(0),
		CodeHash: codeHash.Bytes(),
	}))
	for _, s := range found.Storage {
		suite.app.EvmKeeper.SetState(suite.ctx, contract,
			common.HexToHash(s.Key), common.HexToHash(s.Value).Bytes())
	}

	var erc20Genesis erc20types.GenesisState
	suite.Require().NoError(suite.app.AppCodec().UnmarshalJSON(appState["erc20"], &erc20Genesis))
	var registered bool
	for _, pair := range erc20Genesis.TokenPairs {
		if common.HexToAddress(pair.Erc20Address) != contract {
			continue
		}
		suite.Require().True(pair.IsNativeERC20(),
			"fixture assumption broken: %s is not a deployed ERC-20 on mainnet", dstrxContract)
		suite.Require().NoError(suite.app.Erc20Keeper.SetToken(suite.ctx, pair))
		registered = true
	}
	suite.Require().True(registered, "expected a mainnet x/erc20 token pair for %s", dstrxContract)

	suite.T().Logf("seeded real DSTRX contract %s (%d bytes of code, %d storage slots)",
		dstrxContract, len(code), len(found.Storage))
	return contract
}

// mainnetBankHolders returns blacklisted addresses that hold a bank
// balance on real mainnet, with the exact coins they held, largest first
// and the attacker excluded.
type mainnetBankHolder struct {
	addr  sdk.AccAddress
	bech  string
	coins sdk.Coins
}

func (suite *KeeperTestSuite) mainnetBankHolders(appState map[string]json.RawMessage, limit int) []mainnetBankHolder {
	var gs banktypes.GenesisState
	suite.Require().NoError(suite.app.AppCodec().UnmarshalJSON(appState["bank"], &gs))
	leaked := suite.readLeakedAddresses()

	holders := make([]mainnetBankHolder, 0, limit)
	for _, b := range gs.Balances {
		if !leaked[b.Address] || b.Address == attackerAddress || b.Coins.IsZero() {
			continue
		}
		addr, err := sdk.AccAddressFromBech32(b.Address)
		suite.Require().NoError(err)
		holders = append(holders, mainnetBankHolder{addr: addr, bech: b.Address, coins: b.Coins})
	}
	// Largest first, so a small sample still exercises real amounts rather
	// than whatever dust happens to sort first.
	sort.Slice(holders, func(i, j int) bool {
		return holders[i].coins.AmountOf(multiStakingCoinDenom).
			GT(holders[j].coins.AmountOf(multiStakingCoinDenom))
	})
	if limit > 0 && len(holders) > limit {
		holders = holders[:limit]
	}
	return holders
}

// readAttackerInflows returns how much RIO each address lost to the
// attacker, keyed by bech32 address, in base units.
func (suite *KeeperTestSuite) readAttackerInflows() map[string]math.Int {
	f, err := os.Open(attackerInflowPath)
	if os.IsNotExist(err) {
		suite.T().Skipf("%s not present in this checkout", attackerInflowPath)
	}
	suite.Require().NoError(err)
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	suite.Require().NoError(err)
	suite.Require().Greater(len(records), 1, "expected a header and at least one row")

	header := records[0]
	col := func(name string) int {
		for i, h := range header {
			if strings.TrimSpace(h) == name {
				return i
			}
		}
		suite.FailNowf("column not found", "%s has no %q column", attackerInflowPath, name)
		return -1
	}
	addrCol, sentCol := col("address"), col("sent_to_attacker_ario")

	inflows := make(map[string]math.Int, len(records)-1)
	for _, r := range records[1:] {
		addr := strings.TrimSpace(r[addrCol])
		amount, ok := math.NewIntFromString(strings.TrimSpace(r[sentCol]))
		suite.Require().True(ok, "unparseable sent_to_attacker_ario for %s", addr)
		suite.Require().True(amount.IsPositive(), "%s has a non-positive inflow", addr)
		inflows[addr] = amount
	}
	return inflows
}

// mainnetCoinsOf returns the exact coins one address held on real
// mainnet, or nil if it held none.
func (suite *KeeperTestSuite) mainnetCoinsOf(appState map[string]json.RawMessage, bech string) sdk.Coins {
	var gs banktypes.GenesisState
	suite.Require().NoError(suite.app.AppCodec().UnmarshalJSON(appState["bank"], &gs))
	for _, b := range gs.Balances {
		if b.Address == bech {
			return b.Coins
		}
	}
	return nil
}

// dstrxBalanceOf reads a DSTRX balance straight out of contract storage.
// Used only to pick fixtures and to assert against; the module under test
// goes through the EVM.
func (suite *KeeperTestSuite) dstrxBalanceOf(contract common.Address, addr sdk.AccAddress) *big.Int {
	slot := crypto.Keccak256Hash(
		common.LeftPadBytes(common.BytesToAddress(addr.Bytes()).Bytes(), 32),
		common.LeftPadBytes(big.NewInt(dstrxBalancesSlot).Bytes(), 32),
	)
	return suite.app.EvmKeeper.GetState(suite.ctx, contract, slot).Big()
}

// mainnetDSTRXHolders finds which blacklisted addresses actually hold
// DSTRX, by reading the real contract storage rather than by hard-coding a
// list that would rot the moment the export is refreshed.
func (suite *KeeperTestSuite) mainnetDSTRXHolders(contract common.Address, blacklisted []string) []sdk.AccAddress {
	var holders []sdk.AccAddress
	for _, b := range blacklisted {
		if b == attackerAddress {
			continue
		}
		addr, err := sdk.AccAddressFromBech32(b)
		if err != nil {
			continue
		}
		if suite.dstrxBalanceOf(contract, addr).Sign() > 0 {
			holders = append(holders, addr)
		}
	}
	return holders
}

// ---------------------------------------------------------------------------
// The test
// ---------------------------------------------------------------------------

// TestMainnetRecoveryAcrossAllAssetKinds is the end-to-end rehearsal of a
// real recovery against real state: the actual frozen address list, the
// actual balances those addresses held, their actual delegations, and the
// actual ERC-20 contract they held tokens in.
//
// For every blacklisted address in the sample EXCEPT the attacker's own,
// it mints a fresh recovery wallet, links it as the admin would, and then
// claims each kind of asset with the message that owns it. It finishes by
// checking the attacker: no link, nothing moved, still frozen.
func (suite *KeeperTestSuite) TestMainnetRecoveryAcrossAllAssetKinds() {
	// 0 means every eligible address in the export. Set it to a small
	// number to shorten a local run; the assertions do not change.
	const sampleSize = 0

	appState := suite.readMainnetExport()
	suite.seedRealBankDenomMetadata(appState)
	suite.seedRealAssetTokens(appState)

	start := time.Now()
	blacklisted := suite.seedMainnetBlacklist(appState)
	suite.T().Logf("seeded %d real blacklisted addresses in %s", len(blacklisted), time.Since(start))

	attacker, err := sdk.AccAddressFromBech32(attackerAddress)
	suite.Require().NoError(err)
	attackerCoins := suite.mainnetCoinsOf(appState, attackerAddress)
	suite.Require().False(attackerCoins.IsZero(), "expected the attacker to hold coins on mainnet")
	suite.Require().True(suite.app.BlacklistKeeper.IsBlacklisted(suite.ctx, attacker),
		"fixture assumption broken: the attacker should be on the real blacklist")

	contract := suite.seedMainnetERC20Contract(appState)

	// What each address actually lost, and the fact that the pool covers
	// all of it exactly.
	inflows := suite.readAttackerInflows()
	owed := math.ZeroInt()
	for _, amount := range inflows {
		owed = owed.Add(amount)
	}
	suite.Require().True(owed.Equal(attackerCoins.AmountOf(multiStakingCoinDenom)),
		"the inflow figures (%s) and the attacker's balance (%s) disagree; one of the two is stale",
		owed, attackerCoins.AmountOf(multiStakingCoinDenom))
	suite.T().Logf("INFLOWS: %d addresses lost %s %s to the attacker, which is exactly what it holds",
		len(inflows), owed, multiStakingCoinDenom)

	// ---- pick the sample, per asset kind, from real holdings ----
	bankHolders := suite.mainnetBankHolders(appState, sampleSize)
	suite.Require().NotEmpty(bankHolders)

	delegators := suite.mainnetLeakedSamplesForDenom(appState, multiStakingCoinDenom)
	suite.Require().NotEmpty(delegators)
	if sampleSize > 0 && len(delegators) > sampleSize {
		delegators = delegators[:sampleSize]
	}

	dstrxHolders := suite.mainnetDSTRXHolders(contract, blacklisted)
	suite.Require().NotEmpty(dstrxHolders, "expected real blacklisted DSTRX holders in the mainnet export")
	suite.T().Logf("sample: %d bank holders, %d delegators, %d DSTRX holders",
		len(bankHolders), len(delegators), len(dstrxHolders))

	// ---- put the real holdings onto the test chain ----
	//
	// The attacker's balance goes on too. It is never linked and never
	// claimed, but without it on the chain there is nothing to check at
	// the end, and "the attacker was left alone" would pass by default.
	suite.Require().NoError(suite.app.ClaimKeeper.SetAttacker(suite.ctx, attacker))
	suite.Require().NoError(suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, attackerCoins))
	suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, attacker, attackerCoins))
	suite.T().Logf("ATTACKER: funded with its real mainnet holdings (%s), excluded from every claim", attackerCoins)

	for _, h := range bankHolders {
		suite.Require().NoError(suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, h.coins))
		suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, h.addr, h.coins))
	}
	for _, d := range delegators {
		suite.delegate(d.old, d.amount)
	}

	// ---- one recovery wallet per victim, never for the attacker ----
	recovery := make(map[string]sdk.AccAddress)
	victims := make([]sdk.AccAddress, 0)
	addVictim := func(old sdk.AccAddress) {
		if old.Equals(attacker) {
			suite.FailNow("the attacker was selected as a recovery subject", old.String())
		}
		if _, seen := recovery[old.String()]; seen {
			return
		}
		recovery[old.String()] = testutil.GenAddress()
		victims = append(victims, old)
	}
	for _, h := range bankHolders {
		addVictim(h.addr)
	}
	for _, d := range delegators {
		addVictim(d.old)
	}
	for _, h := range dstrxHolders {
		addVictim(h)
	}
	// Addresses the attacker drained to zero hold nothing to sweep, so they
	// appear in no list above -- and they are exactly the ones for whom
	// ClaimFromAttacker is the only thing that recovers anything.
	inflowOrder := make([]string, 0, len(inflows))
	for bech := range inflows {
		inflowOrder = append(inflowOrder, bech)
	}
	sort.Strings(inflowOrder)
	for _, bech := range inflowOrder {
		addr, err := sdk.AccAddressFromBech32(bech)
		suite.Require().NoError(err)
		addVictim(addr)
	}

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	linkStart := time.Now()
	for _, old := range victims {
		_, err := srv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
			Admin:      suite.admin,
			OldAddress: old.String(),
			NewAddress: recovery[old.String()].String(),
		})
		suite.Require().NoError(err, "LinkAddress failed for %s", old)
	}
	suite.T().Logf("linked %d recovery wallets in %s", len(victims), time.Since(linkStart))

	// ---- claim bank ----
	//
	// First, pin the reason the claim needs its module-account hop at all,
	// on the holder with the most RIO -- a plain bond denom, which no
	// x/asset Token governs, so the blacklist is unambiguously what
	// refuses the transfer.
	richest := bankHolders[0]
	directCtx, _ := suite.ctx.CacheContext()
	suite.Require().ErrorContains(suite.app.BankKeeper.SendCoins(
		directCtx, richest.addr, recovery[richest.bech],
		sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, richest.coins.AmountOf(multiStakingCoinDenom))),
	), "blacklisted")

	bankSwept := sdk.NewCoins()
	bankAwarded := sdk.NewCoins()
	attackerBefore := suite.app.BankKeeper.GetBalance(suite.ctx, attacker, multiStakingCoinDenom).Amount

	for _, h := range bankHolders {
		newAddr := recovery[h.bech]
		newRIOBefore := suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, multiStakingCoinDenom).Amount

		// Every one of these addresses is refused a direct transfer out.
		// The reason varies with what they hold -- the blacklist for plain
		// denoms, x/asset's own restriction first for a permissioned one
		// like almx or arst -- and either way the claim is the only way
		// the funds move.
		directCtx, _ := suite.ctx.CacheContext()
		suite.Require().Error(suite.app.BankKeeper.SendCoins(directCtx, h.addr, newAddr, h.coins),
			"a direct transfer out of frozen %s should have been refused", h.bech)

		res, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
			Admin:      suite.admin,
			OldAddress: h.bech,
		})
		suite.Require().NoError(err, "MigrateBankBalance failed for %s", h.bech)
		suite.Require().Equal(newAddr.String(), res.NewAddress)

		// The sweep took the whole balance, with no amount given for it.
		suite.Require().True(res.Swept.Equal(h.coins),
			"%s: swept %s, expected its whole balance %s", h.bech, res.Swept, h.coins)

		// Then, separately, what the attacker took comes back out of the
		// attacker's own account. Only RIO: the attacker holds no LMX, no
		// RST and no IBC coin, so those can only ever return via the sweep.
		bankSwept = bankSwept.Add(res.Swept...)

		// The old address is empty in every denom it held.
		for _, c := range h.coins {
			suite.Require().True(suite.app.BankKeeper.GetBalance(suite.ctx, h.addr, c.Denom).IsZero(),
				"%s still holds %s", h.bech, c.Denom)
		}
		// The new address holds what the sweep brought, on top of
		// anything it already had.
		for _, c := range h.coins {
			suite.Require().True(
				suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, c.Denom).Amount.Equal(newRIOBefore.Add(c.Amount)),
				"%s did not receive the right %s", newAddr, c.Denom)
		}
	}

	suite.T().Logf("MIGRATE BANK: %d/%d holders drained; swept %s", len(bankHolders), len(bankHolders), bankSwept)

	// ---- claim what the attacker took, address by address ----
	//
	// Separate from the sweep above because it answers a separate
	// question: not "what is still here" but "what was taken". The figure
	// is the incident's own, not derived from chain state.
	for _, bech := range inflowOrder {
		newAddr := recovery[bech]
		lost := sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, inflows[bech]))
		before := suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, multiStakingCoinDenom).Amount

		res, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
			Admin:        suite.admin,
			OldAddress:   bech,
			FromAttacker: lost,
		})
		suite.Require().NoError(err, "ClaimFromAttacker failed for %s", bech)
		suite.Require().Equal(newAddr.String(), res.NewAddress)

		suite.Require().True(
			suite.app.BankKeeper.GetBalance(suite.ctx, newAddr, multiStakingCoinDenom).Amount.Equal(before.Add(inflows[bech])),
			"%s was not made whole for what it lost", bech)
		bankAwarded = bankAwarded.Add(lost...)
	}

	// Every claim came out of the attacker, and the pool covered all of
	// them exactly -- the attacker is left with nothing.
	attackerAfter := suite.app.BankKeeper.GetBalance(suite.ctx, attacker, multiStakingCoinDenom).Amount
	suite.Require().True(
		attackerAfter.Equal(attackerBefore.Sub(bankAwarded.AmountOf(multiStakingCoinDenom))),
		"the attacker's balance does not match what was paid out of it: %s vs %s",
		attackerAfter, attackerBefore.Sub(bankAwarded.AmountOf(multiStakingCoinDenom)))
	suite.Require().True(attackerAfter.IsZero(),
		"the attacker should be emptied exactly, %s left", attackerAfter)
	suite.T().Logf("CLAIM FROM ATTACKER: paid %s to %d addresses; attacker emptied to zero",
		bankAwarded, len(inflowOrder))

	// ---- claim multistaking ----
	stakeMoved := math.ZeroInt()
	for _, d := range delegators {
		newAddr := recovery[d.old.String()]

		sharesBefore, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, d.old, suite.validator)
		suite.Require().NoError(err)

		_, err = srv.ClaimDelegations(suite.ctx, &types.MsgClaimDelegations{
			Admin:      suite.admin,
			OldAddress: d.old.String(),
		})
		suite.Require().NoError(err, "ClaimDelegations failed for %s", d.old)

		_, err = suite.app.StakingKeeper.GetDelegation(suite.ctx, d.old, suite.validator)
		suite.Require().Error(err, "%s still has a delegation", d.old)

		newDel, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, newAddr, suite.validator)
		suite.Require().NoError(err, "%s did not receive the delegation", newAddr)
		suite.Require().True(newDel.Shares.Equal(sharesBefore.Shares))
		stakeMoved = stakeMoved.Add(d.amount)

		lock, found := suite.app.MultiStakingKeeper.GetMultiStakingLock(suite.ctx,
			multistakingtypes.MultiStakingLockID(newAddr.String(), suite.validator.String()))
		suite.Require().True(found, "%s did not receive the multi-staking lock", newAddr)
		suite.Require().True(lock.LockedCoin.Amount.Equal(d.amount))
		_, stillThere := suite.app.MultiStakingKeeper.GetMultiStakingLock(suite.ctx,
			multistakingtypes.MultiStakingLockID(d.old.String(), suite.validator.String()))
		suite.Require().False(stillThere, "%s still holds its multi-staking lock", d.old)
	}
	suite.T().Logf("MULTISTAKING: %d/%d delegations re-keyed, %s %s moved, every lock followed",
		len(delegators), len(delegators), stakeMoved, multiStakingCoinDenom)

	// ---- claim ERC-20 ----
	dstrxMoved := new(big.Int)
	for _, old := range dstrxHolders {
		newAddr := recovery[old.String()]
		held := suite.dstrxBalanceOf(contract, old)
		suite.Require().Positive(held.Sign())
		newBefore := suite.dstrxBalanceOf(contract, newAddr)

		res, err := srv.ClaimTokens(suite.ctx, &types.MsgClaimTokens{
			Admin:          suite.admin,
			OldAddress:     old.String(),
			Erc20Contracts: []string{dstrxContract},
		})
		suite.Require().NoError(err, "ClaimTokens failed for %s", old)
		suite.Require().Equal(uint64(1), res.Erc20ContractsSwept)

		suite.Require().Zero(suite.dstrxBalanceOf(contract, old).Sign(), "%s still holds DSTRX", old)
		suite.Require().Equal(
			new(big.Int).Add(newBefore, held),
			suite.dstrxBalanceOf(contract, newAddr),
			"%s did not receive the DSTRX", newAddr)
		dstrxMoved.Add(dstrxMoved, held)
	}
	suite.T().Logf("ERC-20: %d/%d DSTRX holders swept, %s units moved, every old balance now zero",
		len(dstrxHolders), len(dstrxHolders), dstrxMoved)

	// ---- what is still frozen when the dust settles ----
	//
	// Everything recoverable has moved, so whatever RIO is still sitting on
	// a blacklisted address is there on purpose or because nothing in this
	// module can reach it. Both are worth a number rather than a shrug.
	residual := math.ZeroInt()
	for _, b := range blacklisted {
		addr, err := sdk.AccAddressFromBech32(b)
		suite.Require().NoError(err)
		residual = residual.Add(suite.app.BankKeeper.GetBalance(suite.ctx, addr, multiStakingCoinDenom).Amount)
	}
	attackerHeld := suite.app.BankKeeper.GetBalance(suite.ctx, attacker, multiStakingCoinDenom).Amount
	suite.T().Logf("RESIDUAL: %s %s left on all %d blacklisted addresses, attacker included",
		residual, multiStakingCoinDenom, len(blacklisted))

	// Not one frozen address is left holding RIO -- the attacker included,
	// because what it held was exactly what it owed.
	suite.Require().True(residual.IsZero(),
		"%s RIO is still sitting on frozen addresses", residual)
	suite.Require().True(attackerHeld.IsZero())

	// Spelled out per address, so a failure names the address rather than
	// just a total that does not add up.
	for _, b := range blacklisted {
		addr, err := sdk.AccAddressFromBech32(b)
		suite.Require().NoError(err)
		suite.Require().True(suite.app.BankKeeper.GetBalance(suite.ctx, addr, multiStakingCoinDenom).IsZero(),
			"%s still holds RIO after the recovery", b)
	}

	// ---- the attacker is left exactly as the fork left it ----
	suite.Require().False(suite.app.ClaimKeeper.IsLinked(suite.ctx, attacker),
		"the attacker must never be linked to a recovery wallet")
	suite.Require().True(suite.app.BlacklistKeeper.IsBlacklisted(suite.ctx, attacker),
		"the attacker must still be frozen after the recovery run")

	// And with no link, every claim message refuses it outright -- the
	// exclusion is enforced by the module, not merely by this test
	// remembering to skip it.
	for name, call := range map[string]func() error{
		"ClaimDelegations": func() error {
			_, err := srv.ClaimDelegations(suite.ctx, &types.MsgClaimDelegations{Admin: suite.admin, OldAddress: attackerAddress})
			return err
		},
		"MigrateBankBalance": func() error {
			_, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
				Admin: suite.admin, OldAddress: attackerAddress,
			})
			return err
		},
		"ClaimFromAttacker": func() error {
			_, err := srv.MigrateBankBalance(suite.ctx, &types.MsgMigrateBankBalance{
				Admin: suite.admin, OldAddress: attackerAddress,
				FromAttacker: sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, math.NewInt(1))),
			})
			return err
		},
		"ClaimTokens": func() error {
			_, err := srv.ClaimTokens(suite.ctx, &types.MsgClaimTokens{
				Admin: suite.admin, OldAddress: attackerAddress, Erc20Contracts: []string{dstrxContract},
			})
			return err
		},
	} {
		suite.Require().ErrorContains(call(), "is not linked", "%s should refuse the attacker", name)
	}

	suite.T().Logf("TOTAL: %d distinct victims recovered in %s; attacker left frozen and unlinked",
		len(victims), time.Since(start))
}
