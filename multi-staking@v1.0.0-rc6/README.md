# multi-staking-module

The multi-staking-module is a module that allows the cosmos-sdk staking system to support many types of token 

## Multi staking design

Given the fact that several core sdk modules such as distribution or slashing is dependent on the sdk staking module, we design the multi staking module as a wrapper around the sdk staking module so that there's no need to replace the sdk staking module

The multi staking module has the following features:
- Staking with many diffrent type of tokens
- Bond denom selection via Gov proposal
- A validator's delegations can only be in one denom