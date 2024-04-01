package types

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewUnlockEntry(creationHeight int64, weightedCoin MultiStakingCoin) UnlockEntry {
	return UnlockEntry{
		CreationHeight: creationHeight,
		UnlockingCoin:  weightedCoin,
	}
}

// String implements the stringer interface for a UnlockEntry.
func (e UnlockEntry) String() string {
	out, _ := yaml.Marshal(e)
	return string(out)
}

func (u UnlockEntry) GetBondWeight() sdk.Dec {
	return u.UnlockingCoin.BondWeight
}

func (unlockEntry UnlockEntry) UnbondAmountToUnlockAmount(unbondAmount math.Int) math.Int {
	return sdk.NewDecFromInt(unbondAmount).Quo(unlockEntry.GetBondWeight()).TruncateInt()
}

func (unlockEntry UnlockEntry) UnlockAmountToUnbondAmount(unlockAmount math.Int) math.Int {
	return unlockEntry.GetBondWeight().MulInt(unlockAmount).TruncateInt()
}

// NewMultiStakingUnlock - create a new MultiStaking unlock object
//
//nolint:interfacer
func NewMultiStakingUnlock(
	unlockID UnlockID, creationHeight int64, weightedCoin MultiStakingCoin,
) MultiStakingUnlock {
	return MultiStakingUnlock{
		UnlockID: unlockID,
		Entries: []UnlockEntry{
			NewUnlockEntry(creationHeight, weightedCoin),
		},
	}
}

func (unlock MultiStakingUnlock) Validate() error {
	if _, err := sdk.AccAddressFromBech32(unlock.UnlockID.MultiStakerAddr); err != nil {
		return err
	}
	if _, err := sdk.ValAddressFromBech32(unlock.UnlockID.ValAddr); err != nil {
		return err
	}
	for _, entry := range unlock.Entries {
		if entry.CreationHeight <= 0 {
			return ErrInvalidMultiStakingUnlocksCreationHeight
		}

		if err := entry.UnlockingCoin.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (unlock *MultiStakingUnlock) FindEntryIndexByHeight(creationHeight int64) (int, bool) {
	for index, unlockEntry := range unlock.Entries {
		if unlockEntry.CreationHeight == creationHeight {
			return index, true
		}
	}
	return -1, false
}

// AddEntry - append entry to the unbonding delegation
func (unlock *MultiStakingUnlock) AddEntry(creationHeight int64, weightedCoin MultiStakingCoin) {
	// Check the entries exists with creation_height and complete_time
	entryIndex, found := unlock.FindEntryIndexByHeight(creationHeight)
	// entryIndex exists
	if found {
		unlockEntry := unlock.Entries[entryIndex]
		unlockEntry.UnlockingCoin = unlockEntry.UnlockingCoin.Add(weightedCoin)

		// update the entry
		unlock.Entries[entryIndex] = unlockEntry
	} else {
		// append the new unbond delegation entry
		entry := NewUnlockEntry(creationHeight, weightedCoin)
		unlock.Entries = append(unlock.Entries, entry)
	}
}

// RemoveCoinFromEntry - remove multi staking coin from unlocking entry
func (unlock *MultiStakingUnlock) RemoveCoinFromEntry(entryIndex int, amount math.Int) error {
	entriesLen := len(unlock.Entries)
	if entriesLen == 0 || entryIndex < 0 || entryIndex >= entriesLen {
		return fmt.Errorf("entry index is out of bound")
	}

	unlockEntry := unlock.Entries[entryIndex]
	if unlockEntry.UnlockingCoin.Amount.LT(amount) {
		return fmt.Errorf("cancel amount is greater than the unlocking entry amount")
	}

	updatedAmount := unlockEntry.UnlockingCoin.Amount.Sub(amount)
	if updatedAmount.IsZero() {
		unlock.RemoveEntry(entryIndex)
	} else {
		unlock.Entries[entryIndex].UnlockingCoin.Amount = updatedAmount
	}

	return nil
}

// RemoveEntry - remove entry at index i to the multi staking unlock
func (unlock *MultiStakingUnlock) RemoveEntry(i int) {
	unlock.Entries = append(unlock.Entries[:i], unlock.Entries[i+1:]...)
}

// RemoveEntryAtCreationHeight - remove entry at creation height to the multi staking unlock
func (unlock *MultiStakingUnlock) RemoveEntryAtCreationHeight(creationHeight int64) {
	// Check the entries exists with creation_height and complete_time
	entryIndex, found := unlock.FindEntryIndexByHeight(creationHeight)
	// entryIndex exists
	if found {
		unlock.RemoveEntry(entryIndex)
	}
}

// String returns a human readable string representation of an MultiStakingUnlock.
func (unlock MultiStakingUnlock) String() string {
	out := fmt.Sprintf(`Unlock ID: %s
	Entries:`, unlock.UnlockID)
	for i, entry := range unlock.Entries {
		out += fmt.Sprintf(`    Unbonding Delegation %d:
      Creation Height:           %v
     `, i, entry.CreationHeight,
		)
	}

	return out
}

// MultiStakingUnlocks is a collection of MultiStakingUnlock
// type MultiStakingUnlocks []UnlockEntry

// func (ubds MultiStakingUnlocks) String() (out string) {
// 	for _, u := range ubds {
// 		out += u.String() + "\n"
// 	}

// 	return strings.TrimSpace(out)
// }

// func NewUnbonedMultiStakingRecord( // ?
// 	delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress, creationHeight int64,
// 	completionTime time.Time, rate sdk.Dec, balance math.Int,
// ) Unlock {
// 	return UnbonedMultiStakingRecord{
// 		CreationHeight:  creationHeight,
// 		CompletionTime:  completionTime,
// 		ConversionRatio: rate,
// 		InitialBalance:  balance,
// 		Balance:         balance,
// 	}
// }

// // String implements the stringer interface for a UnlockEntry.
// func (e UnbonedMultiStakingRecord) String() string {
// 	out, _ := yaml.Marshal(e)
// 	return string(out)
// }
