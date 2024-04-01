## Government Proposal

### Add Bond Token Proposals

We can designate a token as a `multistaking coin` by submiting an `AddMultiStakingCoinProposal`. In this proposal, we specify the token's `denom` and its `BondWeight`, if the proposal passes, the specified token will become a `multistaking coin` with the designated `BondWeight`.

### Change Bond Token Weight Proposals

We can alter the `BondWeight` of a `multistaking coin` by submiting a `UpdateBondWeightProposal`. This proposal requires specifying `denom` of the `multistaking coin` and the new `BondWeight`, if the proposal is passed the specified `multistaking coin` have its `BondWeight` changed to new value that decleared by the proposal.