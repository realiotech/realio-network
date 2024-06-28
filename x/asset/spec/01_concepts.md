<!--
order: 1
-->

# Concepts

## The Realio Asset Token Model

The Realio Asset module is centered around a token model where certain whitelisted accounts can issue their own token. A token issued by this module will be managed by a set of privileged accounts. These privileged accounts are assigned by its manager (either an account or a module/contract).

### System of privileged accounts

Privileged accounts of a token are accounts that can execute certain actions for that token. There're are several types of privileges, each has its own logic to define the actions which accounts of said type can execute. We wanna decouple the logic of these privileges from the `Asset module` logic, meaning that privileges will be defined in separate packages/modules, thus, developers can customize their type of privilege without modifying the `Asset Module`. Doing this allows our privileges system to be extensible while keeping the core logic of `Asset Module` untouched and simple, avoiding complicated migration when we expand our privileges system.

In order for a privilege to integrate into the `Asset Module`. It has to implement the `Privilege` interface and has its implementation registered via calling `RegisterPrivilege`. Once that is done, we can make said privilege available onchain by executing `AddPrivilege` gov proposals. This procedure is similar to the `SoftwareUpgrade` via gov proposals, however, we don't need to worry about having to write an `upgrade handler`, just import the new privilege into `Asset Module` and we're good to go.

It's important to note that the token manager can choose what privileges it wants to disable for its token Which is specified by the token manager when creating the token. After creating the token, all the enabled privileges will be assigned to the token manager in default but the token manager can assign privileges to different accounts later on.

We have already defined basic privileges: "mint", "freeze", "clawback", "transfer_auth". These privileges will be included in the default settings of the module.