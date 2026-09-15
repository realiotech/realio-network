package keeper_test

import (
	"encoding/json"
	"os"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	"github.com/realiotech/realio-network/testutil"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
	"github.com/realiotech/realio-network/x/claim/keeper"
	"github.com/realiotech/realio-network/x/claim/types"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

const (
	mainnetExportPath   = "../testdata/exported_mainnet_after.json"
	leakedAddressesPath = "../../../app/migrations/leaked_addresses.json"
)

// readMainnetExport reads and decodes the app_state envelope of the real
// mainnet export, skipping the calling test if the (untracked, large) file
// isn't present in this checkout.
func (suite *KeeperTestSuite) readMainnetExport() map[string]json.RawMessage {
	if _, err := os.Stat(mainnetExportPath); os.IsNotExist(err) {
		suite.T().Skipf("skipping: %s not present in this checkout", mainnetExportPath)
	}

	raw, err := os.ReadFile(mainnetExportPath)
	suite.Require().NoError(err)

	var envelope struct {
		AppState map[string]json.RawMessage `json:"app_state"`
	}
	suite.Require().NoError(json.Unmarshal(raw, &envelope))
	return envelope.AppState
}

func (suite *KeeperTestSuite) readLeakedAddresses() map[string]bool {
	leakedRaw, err := os.ReadFile(leakedAddressesPath)
	suite.Require().NoError(err)
	var leakedList []string
	suite.Require().NoError(json.Unmarshal(leakedRaw, &leakedList))
	leaked := make(map[string]bool, len(leakedList))
	for _, a := range leakedList {
		leaked[a] = true
	}
	return leaked
}

// mainnetLeakedSample is one real leaked delegator pulled from the mainnet
// export together with its original locked-coin amount in a chosen denom.
type mainnetLeakedSample struct {
	old    sdk.AccAddress
	amount math.Int
}

// mainnetLeakedSamplesForDenom cross-references the real mainnet export
// against the real leaked-address list and returns every leaked
// delegator's original locked-coin amount for denom, summed across
// validators when a delegator locked that denom with more than one (the
// caller is expected to seed all samples onto a single test validator, so
// per-validator entries for the same delegator would otherwise become
// duplicate old addresses, which LinkAddress correctly rejects). Both
// input files are the actual incident-response data for this repo, not
// synthetic fixtures.
func (suite *KeeperTestSuite) mainnetLeakedSamplesForDenom(appState map[string]json.RawMessage, denom string) []mainnetLeakedSample {
	var msGenesis multistakingtypes.GenesisState
	suite.Require().NoError(suite.app.AppCodec().UnmarshalJSON(appState["multistaking"], &msGenesis))

	leaked := suite.readLeakedAddresses()

	type lockKey struct{ del, val string }
	locksByKey := make(map[lockKey]multistakingtypes.MultiStakingCoin, len(msGenesis.MultiStakingLocks))
	for _, l := range msGenesis.MultiStakingLocks {
		locksByKey[lockKey{l.LockID.MultiStakerAddr, l.LockID.ValAddr}] = l.LockedCoin
	}

	amountByAddr := make(map[string]math.Int)
	order := make([]string, 0)
	for _, d := range msGenesis.StakingGenesisState.Delegations {
		if !leaked[d.DelegatorAddress] {
			continue
		}
		lock, found := locksByKey[lockKey{d.DelegatorAddress, d.ValidatorAddress}]
		if !found || lock.Denom != denom {
			continue
		}
		if existing, ok := amountByAddr[d.DelegatorAddress]; ok {
			amountByAddr[d.DelegatorAddress] = existing.Add(lock.Amount)
		} else {
			amountByAddr[d.DelegatorAddress] = lock.Amount
			order = append(order, d.DelegatorAddress)
		}
	}

	samples := make([]mainnetLeakedSample, 0, len(order))
	for _, delAddr := range order {
		addr, err := sdk.AccAddressFromBech32(delAddr)
		suite.Require().NoError(err)
		samples = append(samples, mainnetLeakedSample{old: addr, amount: amountByAddr[delAddr]})
	}
	return samples
}

// seedRealAssetTokens loads the real x/asset genesis (app_state.asset) from
// the mainnet export and writes every Token record into the test app's
// asset store exactly as InitGenesis would, using only x/asset's existing
// exported Token collection -- nothing in x/asset is modified for this.
// Returns the decoded tokens so callers can find which symbol/denom is
// actually permissioned on real mainnet, instead of guessing or fabricating
// one.
func (suite *KeeperTestSuite) seedRealAssetTokens(appState map[string]json.RawMessage) []assettypes.Token {
	var assetGenesis assettypes.GenesisState
	suite.Require().NoError(suite.app.AppCodec().UnmarshalJSON(appState["asset"], &assetGenesis))

	for _, token := range assetGenesis.Tokens {
		suite.Require().NoError(suite.app.AssetKeeper.Token.Set(suite.ctx, assettypes.TokenKey(token.Symbol), token))
	}
	return assetGenesis.Tokens
}

// seedRealBankDenomMetadata loads the real bank denom metadata
// (app_state.bank.denom_metadata) from the mainnet export, so denom->symbol
// resolution (used both by x/asset's own AssetSendRestriction and by
// x/claim's authorizeAssetForDenom) works the same way it does on real
// mainnet instead of finding nothing in the test app's synthetic bank
// genesis.
func (suite *KeeperTestSuite) seedRealBankDenomMetadata(appState map[string]json.RawMessage) {
	var bankGenesis banktypes.GenesisState
	suite.Require().NoError(suite.app.AppCodec().UnmarshalJSON(appState["bank"], &bankGenesis))
	for _, md := range bankGenesis.DenomMetadata {
		suite.app.BankKeeper.SetDenomMetaData(suite.ctx, md)
	}
}

// createTestValidator creates a brand-new validator, self-bonded in denom,
// through the real multistaking CreateValidator message (so every hook a
// live chain would run -- including distribution's AfterValidatorCreated,
// which later delegations and reward withdrawals depend on -- actually
// runs, rather than hand-poking store entries into a state that merely
// looks right). Used to give the mainnet-sample tests a second validator to
// redelegate into, and a validator that accepts a denom the seed validator
// (suite.validator, "ario"-only) doesn't.
func (suite *KeeperTestSuite) createTestValidator(denom string) sdk.ValAddress {
	priv := ed25519.GenPrivKey()
	valAddr := sdk.ValAddress(priv.PubKey().Address())

	selfBond := math.NewInt(1_000_000_000_000)
	coins := sdk.NewCoins(sdk.NewCoin(denom, selfBond))
	suite.Require().NoError(suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))
	suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, sdk.AccAddress(valAddr), coins))

	createMsg, err := stakingtypes.NewMsgCreateValidator(
		valAddr.String(),
		priv.PubKey(),
		sdk.NewCoin(denom, selfBond),
		stakingtypes.Description{Moniker: "test-validator-" + denom},
		stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
		math.OneInt(),
	)
	suite.Require().NoError(err)

	msMsgServer := multistakingkeeper.NewMsgServerImpl(suite.app.MultiStakingKeeper)
	_, err = msMsgServer.CreateValidator(suite.ctx, createMsg)
	suite.Require().NoError(err)

	return valAddr
}

// TestMainnetSampleMigration replays LinkAddress/LinkAddresses against
// every real leaked mainnet delegator that locked "ario" -- the one denom
// the test suite's seed validator (suite.validator) accepts -- split across
// both call styles: the first individualBatchSize addresses one at a time
// via LinkAddress, the rest in groups of batchSize via LinkAddresses. For
// every migrated address it then proves the new address can fully operate
// the position it inherited, not just hold it: a real additional
// MsgDelegate, a real partial MsgUndelegate, and a real partial
// MsgBeginRedelegate into a second validator, each submitted by the new
// address itself through the actual multistaking message server.
func (suite *KeeperTestSuite) TestMainnetSampleMigration() {
	appState := suite.readMainnetExport()
	samples := suite.mainnetLeakedSamplesForDenom(appState, multiStakingCoinDenom)
	suite.Require().NotEmpty(samples, "expected at least one leaked mainnet delegator seedable onto the test validator")
	suite.T().Logf("loaded %d real leaked mainnet delegators (denom=%s)", len(samples), multiStakingCoinDenom)

	start := time.Now()

	// A redelegation destination: same denom as suite.validator, since
	// BeginRedelegate requires both sides to accept the same multistaking
	// coin.
	validator2 := suite.createTestValidator(multiStakingCoinDenom)

	// Seed every sample as a real delegation, through the actual
	// multistaking Delegate message -- so distribution's
	// DelegatorStartingInfo, the MultiStakingLock, and the native
	// Delegation all get created exactly as they would on a live chain.
	msMsgServer := multistakingkeeper.NewMsgServerImpl(suite.app.MultiStakingKeeper)
	for _, s := range samples {
		suite.delegate(s.old, s.amount)
	}
	suite.T().Logf("seeded %d delegations in %s", len(samples), time.Since(start))

	claimSrv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)

	const individualBatchSize = 20
	const batchSize = 10

	individual := samples
	if len(individual) > individualBatchSize {
		individual = individual[:individualBatchSize]
	}
	batched := samples[len(individual):]

	newAddrOf := make(map[string]sdk.AccAddress, len(samples))

	linkStart := time.Now()
	for _, s := range individual {
		newAddr := testutil.GenAddress()
		newAddrOf[s.old.String()] = newAddr
		_, err := claimSrv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
			Admin:      suite.admin,
			OldAddress: s.old.String(),
			NewAddress: newAddr.String(),
		})
		suite.Require().NoError(err, "individual LinkAddress failed for %s", s.old)
	}
	suite.T().Logf("linked %d addresses individually via LinkAddress in %s", len(individual), time.Since(linkStart))

	batchStart := time.Now()
	batchCount := 0
	for i := 0; i < len(batched); i += batchSize {
		end := i + batchSize
		if end > len(batched) {
			end = len(batched)
		}
		chunk := batched[i:end]

		links := make([]*types.LinkAddressPair, len(chunk))
		for j, s := range chunk {
			newAddr := testutil.GenAddress()
			newAddrOf[s.old.String()] = newAddr
			links[j] = &types.LinkAddressPair{OldAddress: s.old.String(), NewAddress: newAddr.String()}
		}

		_, err := claimSrv.LinkAddresses(suite.ctx, &types.MsgLinkAddresses{
			Admin: suite.admin,
			Links: links,
		})
		suite.Require().NoError(err, "batch LinkAddresses failed for batch starting at index %d", i)
		batchCount++
	}
	suite.T().Logf("linked %d addresses in %d batches of up to %d via LinkAddresses in %s", len(batched), batchCount, batchSize, time.Since(batchStart))

	// Verify every migration: old is fully gone, new holds exactly what
	// old had, and new can fully operate the inherited position --
	// delegate more, undelegate part of it, and redelegate part of it --
	// each as a real message new submits itself.
	verifyStart := time.Now()
	topUp := math.NewInt(1_000_000_000)
	for _, s := range samples {
		newAddr := newAddrOf[s.old.String()]

		_, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, s.old, suite.validator)
		suite.Require().ErrorIs(err, stakingtypes.ErrNoDelegation, "old address %s still has a delegation after migration", s.old)

		newDel, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, newAddr, suite.validator)
		suite.Require().NoError(err, "new address %s has no delegation after migration", newAddr)
		suite.Require().True(newDel.Shares.IsPositive())

		_, found := suite.app.MultiStakingKeeper.GetMultiStakingLock(suite.ctx,
			multistakingtypes.MultiStakingLockID(s.old.String(), suite.validator.String()))
		suite.Require().False(found, "old address %s still has a multi-staking lock after migration", s.old)

		newLock, found := suite.app.MultiStakingKeeper.GetMultiStakingLock(suite.ctx,
			multistakingtypes.MultiStakingLockID(newAddr.String(), suite.validator.String()))
		suite.Require().True(found, "new address %s has no multi-staking lock after migration", newAddr)
		suite.Require().True(newLock.LockedCoin.Amount.Equal(s.amount))

		// #1: new delegates more.
		coins := sdk.NewCoins(sdk.NewCoin(multiStakingCoinDenom, topUp))
		suite.Require().NoError(suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))
		suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, newAddr, coins))
		_, err = msMsgServer.Delegate(suite.ctx, &stakingtypes.MsgDelegate{
			DelegatorAddress: newAddr.String(),
			ValidatorAddress: suite.validator.String(),
			Amount:           sdk.NewCoin(multiStakingCoinDenom, topUp),
		})
		suite.Require().NoError(err, "new address %s could not submit its own MsgDelegate after migration", newAddr)

		grownDel, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, newAddr, suite.validator)
		suite.Require().NoError(err)
		suite.Require().True(grownDel.Shares.GT(newDel.Shares), "new address %s's delegation did not grow after its own top-up MsgDelegate", newAddr)

		// #2: new undelegates a small part of the inherited position.
		unbondAmt := sdk.NewCoin(multiStakingCoinDenom, math.NewInt(1_000_000))
		_, err = msMsgServer.Undelegate(suite.ctx, &stakingtypes.MsgUndelegate{
			DelegatorAddress: newAddr.String(),
			ValidatorAddress: suite.validator.String(),
			Amount:           unbondAmt,
		})
		suite.Require().NoError(err, "new address %s could not submit its own MsgUndelegate after migration", newAddr)
		ubd, err := suite.app.StakingKeeper.GetUnbondingDelegation(suite.ctx, newAddr, suite.validator)
		suite.Require().NoError(err, "new address %s has no unbonding delegation after its own MsgUndelegate", newAddr)
		suite.Require().Len(ubd.Entries, 1)

		// #3: new redelegates a small part of the inherited position into
		// a second validator.
		redelegateAmt := sdk.NewCoin(multiStakingCoinDenom, math.NewInt(1_000_000))
		_, err = msMsgServer.BeginRedelegate(suite.ctx, &stakingtypes.MsgBeginRedelegate{
			DelegatorAddress:    newAddr.String(),
			ValidatorSrcAddress: suite.validator.String(),
			ValidatorDstAddress: validator2.String(),
			Amount:              redelegateAmt,
		})
		suite.Require().NoError(err, "new address %s could not submit its own MsgBeginRedelegate after migration", newAddr)
		red, err := suite.app.StakingKeeper.GetRedelegation(suite.ctx, newAddr, suite.validator, validator2)
		suite.Require().NoError(err, "new address %s has no redelegation record after its own MsgBeginRedelegate", newAddr)
		suite.Require().Len(red.Entries, 1)
		dstDel, err := suite.app.StakingKeeper.GetDelegation(suite.ctx, newAddr, validator2)
		suite.Require().NoError(err, "new address %s has no delegation on the redelegation destination validator", newAddr)
		suite.Require().True(dstDel.Shares.IsPositive())
	}
	suite.T().Logf("verified %d migrations (delegate more + undelegate + redelegate, each submitted by the new address itself) in %s", len(samples), time.Since(verifyStart))
	suite.T().Logf("total: %s", time.Since(start))
}

// TestMainnetRealAssetAuthorization is the counterpart to
// TestMainnetSampleMigration for x/asset interop, using real data instead
// of a fabricated Token: the real mainnet asset genesis has an actual
// permissioned Token (symbol "rst", AuthorizationRequired = true) whose
// base bank denom is "arst" -- which is also a real multi-staking lock
// denom several leaked addresses hold. This seeds the real Token records
// (seedRealAssetTokens) and a validator that accepts "arst", migrates
// those real leaked "arst" delegators, and confirms every new address
// comes out already authorized on the real "rst" Token -- discovered from
// the data, not assumed or hand-registered.
func (suite *KeeperTestSuite) TestMainnetRealAssetAuthorization() {
	appState := suite.readMainnetExport()
	suite.seedRealBankDenomMetadata(appState)
	tokens := suite.seedRealAssetTokens(appState)

	const arstDenom = "arst"
	var arstToken *assettypes.Token
	if md, found := suite.app.BankKeeper.GetDenomMetaData(suite.ctx, arstDenom); found {
		// Token.Symbol ("rst") and bank Metadata.Symbol ("RST") disagree in
		// case on real mainnet data; TokenKey lowercases both when used as
		// a store key (and that's exactly what the production lookup in
		// authorizeAssetForDenom relies on), so match the same way here.
		for i := range tokens {
			if assettypes.TokenKey(tokens[i].Symbol) == assettypes.TokenKey(md.Symbol) {
				arstToken = &tokens[i]
				break
			}
		}
	}
	suite.Require().NotNil(arstToken, "expected the real mainnet asset genesis to have a Token for denom %s", arstDenom)
	suite.Require().True(arstToken.AuthorizationRequired, "expected %s (denom %s) to require authorization on real mainnet", arstToken.Symbol, arstDenom)
	suite.T().Logf("discovered real permissioned token %q (AuthorizationRequired=true) for denom %q", arstToken.Symbol, arstDenom)

	samples := suite.mainnetLeakedSamplesForDenom(appState, arstDenom)
	suite.Require().NotEmpty(samples, "expected at least one leaked mainnet delegator holding %s", arstDenom)
	suite.T().Logf("loaded %d real leaked mainnet delegators (denom=%s)", len(samples), arstDenom)

	arstValidator := suite.createTestValidator(arstDenom)

	for _, s := range samples {
		coins := sdk.NewCoins(sdk.NewCoin(arstDenom, s.amount))
		suite.Require().NoError(suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))
		suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, s.old, coins))
		msMsgServer := multistakingkeeper.NewMsgServerImpl(suite.app.MultiStakingKeeper)
		_, err := msMsgServer.Delegate(suite.ctx, &stakingtypes.MsgDelegate{
			DelegatorAddress: s.old.String(),
			ValidatorAddress: arstValidator.String(),
			Amount:           sdk.NewCoin(arstDenom, s.amount),
		})
		suite.Require().NoError(err)

		suite.Require().False(suite.app.AssetKeeper.IsAddressAuthorizedToSend(suite.ctx, arstToken.Symbol, s.old),
			"fixture assumption broken: leaked address %s was already authorized for %s on real mainnet", s.old, arstToken.Symbol)
	}

	claimSrv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	newAddrOf := make(map[string]sdk.AccAddress, len(samples))
	for _, s := range samples {
		newAddr := testutil.GenAddress()
		newAddrOf[s.old.String()] = newAddr
		_, err := claimSrv.LinkAddress(suite.ctx, &types.MsgLinkAddress{
			Admin:      suite.admin,
			OldAddress: s.old.String(),
			NewAddress: newAddr.String(),
		})
		suite.Require().NoError(err, "LinkAddress failed for %s", s.old)
	}

	for _, s := range samples {
		newAddr := newAddrOf[s.old.String()]
		suite.Require().True(suite.app.AssetKeeper.IsAddressAuthorizedToSend(suite.ctx, arstToken.Symbol, newAddr),
			"new address %s was not authorized for the real %s token after migration", newAddr, arstToken.Symbol)
	}
	suite.T().Logf("verified %d real leaked %s holders migrated with x/asset authorization for %q carried over", len(samples), arstDenom, arstToken.Symbol)
}
