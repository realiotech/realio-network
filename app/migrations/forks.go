package migrations

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	blacklistmoduletypes "github.com/realiotech/realio-network/x/blacklist/types"
)

var (
	ForkHeight       = int64(5989487)
	OneEternityLater = time.Date(9999, 9, 9, 9, 9, 9, 9, time.UTC)

	// BlacklistForkHeight seeds x/blacklist with the addresses whose keys
	// leaked. It only ever adds blacklist entries — no bank/staking/
	// multistaking state is touched or moved. It also doubles as the height
	// at which the x/blacklist store itself gets created (see
	// BlacklistStoreUpgrades, wired in setupUpgradeHandlers, app/upgrades.go)
	// — a validator must swap to this binary right as it commits the block
	// before this height, same as any other hardcoded fork.
	BlacklistForkHeight = int64(19573266)

	// BlacklistStoreUpgrades tells the store loader that x/blacklist is a
	// brand new store as of BlacklistForkHeight, not one that should already
	// exist in a chain database that predates this fork.
	BlacklistStoreUpgrades = storetypes.StoreUpgrades{
		Added: []string{blacklistmoduletypes.StoreKey},
	}

	//go:embed leaked_addresses.json
	LeakedAddressesJSON []byte
)

// ScheduleForkUpgrade executes any necessary fork logic for based upon the current
// block height and chain ID (mainnet or testnet). It sets an upgrade plan once
// the chain reaches the pre-defined upgrade height.
//
// CONTRACT: for this logic to work properly it is required to:
//
//  1. Release a non-breaking patch version so that the chain can set the scheduled upgrade plan at upgrade-height.
//  2. Release the software defined in the upgrade-info
func ScheduleForkUpgrade(ctx sdk.Context, k Keepers) {
	if ctx.BlockHeight() == ForkHeight {

		// remove duplicate UnbondingQueueKey
		removeDuplicateValueUnbondingQueueKey(k, ctx)
		removeDuplicateValueRedelegationQueueKey(k, ctx)
		removeDuplicateUnbondingValidator(k, ctx)
	}

	if ctx.BlockHeight() == BlacklistForkHeight {
		// Parsed once and threaded through below — every one of these needs
		// the same decoded address list, and re-parsing/re-decoding
		// leaked_addresses.json (33k+ entries) independently in each was
		// pure repeated work for no benefit.
		leaked := decodeLeakedAddresses()

		seedLeakedAddressBlacklist(k, ctx, leaked)
		rotateAssetManagers(k, ctx)
		unauthorizeLeakedAddresses(k, ctx, leaked)
		rotateBridgeAuthority(k, ctx)
		revokeLeakedAuthzGrants(k, ctx, leaked)
	}
}

// ScheduleValidatorRotation runs the validator-rotation fork (see
// validator_rotation.go) and returns any ABCI validator updates it produced
// (the "old key -> power 0" half of a rotation — see RotateValidators). It
// must be called AFTER the module manager's BeginBlock — see the comment on
// the call site in app/app.go's BeginBlocker for why.
func ScheduleValidatorRotation(ctx sdk.Context, k Keepers) []abci.ValidatorUpdate {
	if ctx.BlockHeight() == ValidatorRotationHeight {
		return RotateValidators(k, ctx)
	}
	return nil
}

func removeDuplicateValueRedelegationQueueKey(k Keepers, ctx sdk.Context) {
	sk := k.StakingKeeper
	cdc := k.Codec
	store := ctx.KVStore(k.StakingStoreKey)

	redelegationTimesliceIterator, err := sk.RedelegationQueueIterator(ctx, OneEternityLater) // make sure to iterate all queue
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

func removeDuplicateUnbondingValidator(k Keepers, ctx sdk.Context) {
	valIter, err := k.StakingKeeper.ValidatorQueueIterator(ctx, OneEternityLater, 99999999999999)
	if err != nil {
		panic(err)
	}

	defer valIter.Close()

	for ; valIter.Valid(); valIter.Next() {
		addrs := stakingtypes.ValAddresses{}
		k.Codec.MustUnmarshal(valIter.Value(), &addrs)

		vals := map[string]bool{}
		for _, valAddr := range addrs.Addresses {
			vals[valAddr] = true
		}

		uniqueAddrs := []string{}
		for valAddr := range vals {
			uniqueAddrs = append(uniqueAddrs, valAddr)
		}
		sort.Strings(uniqueAddrs)

		ctx.KVStore(k.StakingStoreKey).Set(valIter.Key(), k.Codec.MustMarshal(&stakingtypes.ValAddresses{Addresses: uniqueAddrs}))
	}
}

func removeDuplicateValueUnbondingQueueKey(k Keepers, ctx sdk.Context) {
	sk := k.StakingKeeper
	cdc := k.Codec
	store := ctx.KVStore(k.StakingStoreKey)

	unbondingTimesliceIterator, err := sk.UBDQueueIterator(ctx, OneEternityLater) // make sure to iterate all queue
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

// seedLeakedAddressBlacklist adds every address in leaked to x/blacklist. It
// does not touch bank, staking, or multistaking state — the leaked keys'
// funds are left exactly where they are; blacklisting only prevents those
// addresses from ever signing another transaction.
func seedLeakedAddressBlacklist(k Keepers, ctx sdk.Context, leaked []sdk.AccAddress) {
	for _, accAddr := range leaked {
		if err := k.BlacklistKeeper.SetBlacklisted(ctx, accAddr); err != nil {
			panic(fmt.Errorf("blacklist fork: failed to blacklist %q: %w", accAddr, err))
		}
	}
}

// ParseLeakedAddresses unmarshals leaked_addresses.json into its raw bech32
// strings. A malformed file panics rather than silently seeding an empty
// (or partial) blacklist.
func ParseLeakedAddresses() []string {
	var addrs []string
	if err := json.Unmarshal(LeakedAddressesJSON, &addrs); err != nil {
		panic(fmt.Errorf("blacklist fork: failed to parse leaked_addresses.json: %w", err))
	}
	return addrs
}

// decodeLeakedAddresses parses leaked_addresses.json and decodes every
// entry to sdk.AccAddress once, so every fork function that needs the
// leaked-address list at BlacklistForkHeight (seedLeakedAddressBlacklist,
// unauthorizeLeakedAddresses, revokeLeakedAuthzGrants) can share a single
// parsed+decoded copy instead of each re-parsing and re-decoding the same
// 33k+-entry file independently.
func decodeLeakedAddresses() []sdk.AccAddress {
	raw := ParseLeakedAddresses()
	addrs := make([]sdk.AccAddress, len(raw))
	for i, addr := range raw {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			panic(fmt.Errorf("blacklist fork: invalid address %q: %w", addr, err))
		}
		addrs[i] = accAddr
	}
	return addrs
}
