<!--
order: 5
-->

# Gov proposals

The asset module supports the following types of gov proposals:

## TokenCreatorWhitelistProposal

This proposal allows the community to decide on what addresses are allowed to create token via the asset module.

## AddPrilegeProposal

This proposal functioning similarly to the `SoftwareUpgradeProposal`. It specifies the `Privilege` that will be added into the privilege system. Then, It'll stop the chain at a specified height so that a new binary that contains logic for the specified `Privilege` is switched in to continue running the chain. After said process is finished, the new `Privilege` will be available on chain.

## ExecutePrivilege

These types of proposals is implemented as sdk.msg with respective handling methods following `gov.v1beta` standard.
