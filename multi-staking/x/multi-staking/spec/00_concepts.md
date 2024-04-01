# multi-staking-module

The multi-staking-module is a module that allows the cosmos-sdk staking system to support many types of coin 

## Features

- Staking with many diffrent type of coins
- Bond denom selection via Gov proposal
- A validator can only be delegated using with one type of coin
- All usecases of the sdk staking module

## Multi staking design

Given the fact that several core sdk modules such as distribution or slashing is dependent on the sdk staking module, we design the multi staking module as a wrapper around the sdk staking module so that there's no need to replace the sdk staking module and its related modules.

The mechanism of this module is that it still uses the sdk staking module for all the logic related to staking. But since the sdk staking module doesn't allow delegate with different types of coin, in order to support such feature, the multi-staking module will convert (lock and mint) those different coin into the one bond coin that is used by the sdk staking module and then stake with the converted bond coin.

![design](https://hackmd.io/_uploads/B1BduYEh6.png)

## Concepts and Terms

### Bond coin

Bond coin is the only coin that the sdk staking module accepts for delegation. In our design, the bond coin is just a virtual coin used only in the sdk staking layer, serving no other purposes than that. No user accounts are allowed to access to the bond coin.

### Multistaking coin

Multistaking coin refers to the instance of coin that is used to delegate via the multi-staking module. 

It is represented by this [struct](../types/multi_staking.pb.go), which is almost identical to the bank coin, except that it has an additional field called `bond weight`.

### Bond Weight

Each `multistaking coin` instance is associated with a `bond weight`. The `bond weight` value shows the conversion ratio to `bond coin` of that `multistaking coin` instance. It's different than the `bond weight` value set by government prop which specifies the current global `bond weight` value of that type of coin rather than `bond weight` value for a specific instance of `multistaking coin`.

We mentioned above that for each delegation the multi-staking will lock the `multistaking coin` and mint a calculated ammount of `bond token`. The calculation here is a multiplication: minted bond token ammount = multistaking coin amount * bond weight.

### Multistaking lock

`MultistakingLock` is used to keep tracks of the multi-staking coin that is locked for each delegation. `MultistakingLock` contains `LockID` refering to delegation ID (delegator, validator) of the corresponding delegation, and `MultistakingCoin` refering to the instance of `multistaking coin` that is locked.

### Multistaking unlock

`MultistakingUnlock` is used to keep tracks of the multi-staking coin that is unlocking for each unbonding delegation. `MultistakingUnLock` contains `UnLockID` refering to unbonding delegation ID (delegator, validator) of the corresponding unbonding delegation, and `Entries` refering to the instances of `multistaking coin` that is unlocking.