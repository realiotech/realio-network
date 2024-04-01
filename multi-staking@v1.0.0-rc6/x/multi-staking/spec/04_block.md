# End-Block

## Complete Unbonding Delegations

### Calculate total `UnbondedAmount`

* Retrieve `matureUnbondingDelegations` which is the array of all `UnbondingDelegations` that complete in this block

### Staking module EndBlock

* Call `Staking` module `EndBlock` to `CompleteUnbonding`

### MultiStaking module EndBlock

* Iterate over `matureUnbondingDelegations` which was retrieve above

* For each iteration, we will:

    * Calculate amount of `unlockedCoin` that will be return to user by multiply the amount of `unbonded coin` and `bonded weight`

    * Burn the `remainingCoin` that remain on the `Lock` after send `unlockedCoin` to user

    * Delete `UnlockEntry`.

    
