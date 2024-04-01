package keeper

import (
	"fmt"

	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetUnlockEntryAtCreationHeight(ctx sdk.Context, unlockID types.UnlockID, creationHeight int64) (types.UnlockEntry, bool) {
	// get unbonded record
	unlock, found := k.GetMultiStakingUnlock(ctx, unlockID)
	if !found {
		return types.UnlockEntry{}, false
	}
	var (
		unlockEntry      types.UnlockEntry
		foundUnlockEntry = false
	)

	for _, entry := range unlock.Entries {
		if entry.CreationHeight == creationHeight {
			unlockEntry = entry
			foundUnlockEntry = true
			break
		}
	}
	if !foundUnlockEntry {
		return types.UnlockEntry{}, false
	}

	return unlockEntry, foundUnlockEntry
}

// SetMultiStakingUnlockEntry adds an entry to the unbonding delegation at
// the given addresses. It creates the unbonding delegation if it does not exist.
func (k Keeper) SetMultiStakingUnlockEntry(
	ctx sdk.Context, unlockID types.UnlockID,
	multistakingCoin types.MultiStakingCoin,
) types.MultiStakingUnlock {
	unlock, found := k.GetMultiStakingUnlock(ctx, unlockID)
	if found {
		unlock.AddEntry(ctx.BlockHeight(), multistakingCoin)
	} else {
		unlock = types.NewMultiStakingUnlock(unlockID, ctx.BlockHeight(), multistakingCoin)
	}

	k.SetMultiStakingUnlock(ctx, unlock)
	return unlock
}

func (k Keeper) DeleteUnlockEntryAtCreationHeight(
	ctx sdk.Context, unlockID types.UnlockID,
	creationHeight int64,
) error {
	unlock, found := k.GetMultiStakingUnlock(ctx, unlockID)
	if found {
		unlock.RemoveEntryAtCreationHeight(creationHeight)
	} else {
		return fmt.Errorf("can't found unlock entry")
	}

	if len(unlock.Entries) == 0 {
		k.DeleteMultiStakingUnlock(ctx, unlockID)
		return nil
	}

	k.SetMultiStakingUnlock(ctx, unlock)
	return nil
}

func (k Keeper) DecreaseUnlockEntryAmount(
	ctx sdk.Context, unlockID types.UnlockID,
	amount math.Int, creationHeight int64,
) (types.MultiStakingCoin, error) {
	unlockRecord, found := k.GetMultiStakingUnlock(ctx, unlockID)
	if !found {
		return types.MultiStakingCoin{}, fmt.Errorf("not found unlock recored")
	}

	unlockEntryIndex, found := unlockRecord.FindEntryIndexByHeight(creationHeight)
	// entryIndex exists
	if !found {
		return types.MultiStakingCoin{}, fmt.Errorf("unbonding delegation entry is not found at block height %d", creationHeight)
	}

	unlockEntry := unlockRecord.Entries[unlockEntryIndex]
	err := unlockRecord.RemoveCoinFromEntry(unlockEntryIndex, amount)
	if err != nil {
		return types.MultiStakingCoin{}, err
	}

	// set the unlocking record or remove it if there are no more entries
	if len(unlockRecord.Entries) == 0 {
		k.DeleteMultiStakingUnlock(ctx, unlockID)
	} else {
		k.SetMultiStakingUnlock(ctx, unlockRecord)
	}

	return types.NewMultiStakingCoin(unlockEntry.UnlockingCoin.Denom, amount, unlockEntry.GetBondWeight()), nil
}
