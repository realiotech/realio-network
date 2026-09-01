package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	blacklistmoduletypes "github.com/realiotech/realio-network/x/blacklist/types"
)

var (
	ForkHeight        = int64(5989487)
	oneEnternityLater = time.Date(9999, 9, 9, 9, 9, 9, 9, time.UTC)

	// BlacklistForkHeight seeds x/blacklist with the addresses whose keys
	// leaked. It only ever adds blacklist entries — no bank/staking/
	// multistaking state is touched or moved. It also doubles as the height
	// at which the x/blacklist store itself gets created (see
	// blacklistStoreUpgrades, wired in setupUpgradeHandlers, app/upgrades.go)
	// — a validator must swap to this binary right as it commits the block
	// before this height, same as any other hardcoded fork.
	BlacklistForkHeight = int64(19573266)

	// blacklistStoreUpgrades tells the store loader that x/blacklist is a
	// brand new store as of BlacklistForkHeight, not one that should already
	// exist in a chain database that predates this fork.
	blacklistStoreUpgrades = storetypes.StoreUpgrades{
		Added: []string{blacklistmoduletypes.StoreKey},
	}

	//go:embed leaked_addresses.json
	leakedAddressesJSON []byte
)

// ScheduleForkUpgrade executes any necessary fork logic for based upon the current
// block height and chain ID (mainnet or testnet). It sets an upgrade plan once
// the chain reaches the pre-defined upgrade height.
//
// CONTRACT: for this logic to work properly it is required to:
//
//  1. Release a non-breaking patch version so that the chain can set the scheduled upgrade plan at upgrade-height.
//  2. Release the software defined in the upgrade-info
func (app *RealioNetwork) ScheduleForkUpgrade(ctx sdk.Context) {
	if ctx.BlockHeight() == ForkHeight {

		// remove duplicate UnbondingQueueKey
		removeDuplicateValueUnbondingQueueKey(app, ctx)
		removeDuplicateValueRedelegationQueueKey(app, ctx)
		removeDuplicateUnbondingValidator(app, ctx)
	}

	if ctx.BlockHeight() == BlacklistForkHeight {
		seedLeakedAddressBlacklist(app, ctx)
		rotateAssetManagers(app, ctx)
	}
	// NOTE: there are no testnet forks for the existing versions
	// if !types.IsMainnet(ctx.ChainID()) {
	//	return
	//}
	//
	// upgradePlan := upgradetypes.Plan{
	//	Height: ctx.BlockHeight(),
	//}
	//
	//// handle mainnet forks with their corresponding upgrade name and info
	// switch ctx.BlockHeight() {
	// case v2.MainnetUpgradeHeight:
	//	upgradePlan.Name = v2.UpgradeName
	//	upgradePlan.Info = v2.UpgradeInfo
	//default:
	//	// No-op
	//	return
	//}
	//
	//// schedule the upgrade plan to the current block height, effectively performing
	//// a hard fork that uses the upgrade handler to manage the migration.
	// if err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan); err != nil {
	//	panic(
	//		fmt.Errorf(
	//			"failed to schedule upgrade %s during BeginBlock at height %d: %w",
	//			upgradePlan.Name, ctx.BlockHeight(), err,
	//		),
	//	)
	//}
}

// ScheduleValidatorRotation runs the validator-rotation fork (see
// app/validator_rotation.go). It must be called AFTER app.mm.BeginBlock —
// see the comment on that call site in app/app.go's BeginBlocker for why.
func (app *RealioNetwork) ScheduleValidatorRotation(ctx sdk.Context) {
	if ctx.BlockHeight() == ValidatorRotationHeight {
		rotateValidators(app, ctx)
	}
}

func removeDuplicateValueRedelegationQueueKey(app *RealioNetwork, ctx sdk.Context) {
	// Get Staking keeper, codec and staking store
	sk := app.StakingKeeper
	cdc := app.AppCodec()
	store := ctx.KVStore(app.keys[stakingtypes.ModuleName])

	redelegationTimesliceIterator, err := sk.RedelegationQueueIterator(ctx, oneEnternityLater) // make sure to iterate all queue
	if err != nil {
		panic(err)
	}

	defer redelegationTimesliceIterator.Close()

	for ; redelegationTimesliceIterator.Valid(); redelegationTimesliceIterator.Next() {
		timeslice := stakingtypes.DVVTriplets{}
		value := redelegationTimesliceIterator.Value()
		cdc.MustUnmarshal(value, &timeslice)

		triplets := removeDuplicateDVVTriplets(timeslice.Triplets)
		bz := cdc.MustMarshal(&stakingtypes.DVVTriplets{Triplets: triplets})

		store.Set(redelegationTimesliceIterator.Key(), bz)
	}
}

func removeDuplicateDVVTriplets(triplets []stakingtypes.DVVTriplet) []stakingtypes.DVVTriplet {
	var list []stakingtypes.DVVTriplet
	for _, item := range triplets {
		if !containsDVVTriplets(list, item) {
			list = append(list, item)
		}
	}
	return list
}

func containsDVVTriplets(s []stakingtypes.DVVTriplet, e stakingtypes.DVVTriplet) bool {
	for _, a := range s {
		if a.DelegatorAddress == e.DelegatorAddress &&
			a.ValidatorSrcAddress == e.ValidatorSrcAddress &&
			a.ValidatorDstAddress == e.ValidatorDstAddress {
			return true
		}
	}
	return false
}

func removeDuplicateUnbondingValidator(app *RealioNetwork, ctx sdk.Context) {
	valIter, err := app.StakingKeeper.ValidatorQueueIterator(ctx, oneEnternityLater, 99999999999999)
	if err != nil {
		panic(err)
	}

	defer valIter.Close()

	for ; valIter.Valid(); valIter.Next() {
		addrs := stakingtypes.ValAddresses{}
		app.appCodec.MustUnmarshal(valIter.Value(), &addrs)

		vals := map[string]bool{}
		for _, valAddr := range addrs.Addresses {
			vals[valAddr] = true
		}

		uniqueAddrs := []string{}
		for valAddr := range vals {
			uniqueAddrs = append(uniqueAddrs, valAddr)
		}
		sort.Strings(uniqueAddrs)

		ctx.KVStore(app.GetKey(stakingtypes.StoreKey)).Set(valIter.Key(), app.appCodec.MustMarshal(&stakingtypes.ValAddresses{Addresses: uniqueAddrs}))
	}
}

func removeDuplicateValueUnbondingQueueKey(app *RealioNetwork, ctx sdk.Context) {
	// Get Staking keeper, codec and staking store
	sk := app.StakingKeeper
	cdc := app.AppCodec()
	store := ctx.KVStore(app.keys[stakingtypes.ModuleName])

	unbondingTimesliceIterator, err := sk.UBDQueueIterator(ctx, oneEnternityLater) // make sure to iterate all queue
	if err != nil {
		panic(err)
	}

	defer unbondingTimesliceIterator.Close()

	for ; unbondingTimesliceIterator.Valid(); unbondingTimesliceIterator.Next() {
		timeslice := stakingtypes.DVPairs{}
		value := unbondingTimesliceIterator.Value()
		cdc.MustUnmarshal(value, &timeslice)

		dvPairs := removeDuplicatesDVPairs(timeslice.Pairs)
		bz := cdc.MustMarshal(&stakingtypes.DVPairs{Pairs: dvPairs})

		store.Set(unbondingTimesliceIterator.Key(), bz)
	}
}

func removeDuplicatesDVPairs(dvPairs []stakingtypes.DVPair) []stakingtypes.DVPair {
	var list []stakingtypes.DVPair
	for _, item := range dvPairs {
		if !containsDVPairs(list, item) {
			list = append(list, item)
		}
	}
	return list
}

func containsDVPairs(s []stakingtypes.DVPair, e stakingtypes.DVPair) bool {
	for _, a := range s {
		if a.DelegatorAddress == e.DelegatorAddress &&
			a.ValidatorAddress == e.ValidatorAddress {
			return true
		}
	}
	return false
}

// seedLeakedAddressBlacklist adds every address in leaked_addresses.json (a
// JSON array of bech32 address strings) to x/blacklist. It does not touch
// bank, staking, or multistaking state — the leaked keys' funds are left
// exactly where they are; blacklisting only prevents those addresses from
// ever signing another transaction.
func seedLeakedAddressBlacklist(app *RealioNetwork, ctx sdk.Context) {
	for _, addr := range parseLeakedAddresses() {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			panic(fmt.Errorf("blacklist fork: invalid address %q: %w", addr, err))
		}

		if err := app.BlacklistKeeper.SetBlacklisted(ctx, accAddr); err != nil {
			panic(fmt.Errorf("blacklist fork: failed to blacklist %q: %w", addr, err))
		}
	}
}

// parseLeakedAddresses unmarshals leaked_addresses.json. A malformed file
// panics rather than silently seeding an empty (or partial) blacklist.
func parseLeakedAddresses() []string {
	var addrs []string
	if err := json.Unmarshal(leakedAddressesJSON, &addrs); err != nil {
		panic(fmt.Errorf("blacklist fork: failed to parse leaked_addresses.json: %w", err))
	}
	return addrs
}
