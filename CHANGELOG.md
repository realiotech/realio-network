<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

ex: - (upgrade) [#1](https://github.com/realiotech/realio-network/pull/3) Fix Asset types

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [v1.2.0](https://github.com/realiotech/realio-network/releases/tag/v1.2.0) - 2025-04-25

### Bug Fixes

- (evm) [#212](https://github.com/realiotech/realio-network/pull/212): Add Berlin precomiles to memory

### Improvements

- (chore) [#211](https://github.com/realiotech/realio-network/pull/211): Update changelog from v1.0.1
- (feat) [#232](https://github.com/realiotech/realio-network/pull/232): Add bank end blocker

## [v1.0.5](https://github.com/realiotech/realio-network/releases/tag/v1.0.5)

- (chore)[213](https://github.com/realiotech/realio-network/pull/213): register geth tracers at app.go

## [v1.0.4](https://github.com/realiotech/realio-network/releases/tag/v1.0.4)

### Bug Fixes

- [#185](https://github.com/realiotech/realio-network/pull/185) Fix bridge module cli

## [v1.0.3](https://github.com/realiotech/realio-network/releases/tag/v1.0.3) - 2025-03-01

### Improvements

- (app) [#187](https://github.com/realiotech/realio-network/pull/187): Add gentx
- (chore) [#199](https://github.com/realiotech/realio-network/pull/199): Update local node start script
- (bump) [#204](https://github.com/realiotech/realio-network/pull/204): patch ibc-go v8.6.1


## [v1.0.2](https://github.com/realiotech/realio-network/releases/tag/v1.0.2) - 2024-12-19

### Improvements

- (bump) [#184](https://github.com/realiotech/realio-network/pull/184): Patch cosmos sdk v0.50.11

## [v1.0.1](https://github.com/realiotech/realio-network/releases/tag/1.0.1) - 2024-12-13

### Feature

- [#179](https://github.com/realiotech/realio-network/pull/179): Bridge module and upgrade sdk50
- [#180](https://github.com/realiotech/realio-network/pull/180): Update bridge params for testnet
- [#182](https://github.com/realiotech/realio-network/pull/182): Allow MsgBridgeIn transfer coin directly to reciever
- [#183](https://github.com/realiotech/realio-network/pull/183): Update bridge params for mainnet

### Improvements

- (chore) [#172](https://github.com/realiotech/realio-network/pull/172): Update local node scripts for multistaking
- (ci) [#173](https://github.com/realiotech/realio-network/pull/173): Remove unnecessary windows tests
- (chore) [#181](https://github.com/realiotech/realio-network/pull/181): Change upgrade folder naming

## [v0.9.3](https://github.com/realiotech/realio-network/releases/tag/v0.9.3) - 2024-6-26

### Improvements

- (chore) [#157](https://github.com/realiotech/realio-network/pull/157): update the validators' commissions if their commissions are less than 5%

## [v0.9.2](https://github.com/realiotech/realio-network/releases/tag/v0.9.1) - 2024-4-02

### Bug Fixes

- (app) [#133](https://github.com/realiotech/realio-network/pull/133) Update fork for multistaking upgrade

## [v0.9.1](https://github.com/realiotech/realio-network/releases/tag/v0.9.1) - 2024-4-01

### Bug Fixes

- (asset) [#131](https://github.com/realiotech/realio-network/pull/131) Fix send restriction


## [v0.9.0] - 2024-03-27

### Feature

- [#104](https://github.com/realiotech/realio-network/pull/104):  Intergrate [multi-staking module](https://github.com/realiotech/multi-staking) version v1.0.0-rc4

### Improvements

- (docs) [#127](https://github.com/realiotech/realio-network/pull/127): Update swagger docs
- (deps) [#126](https://github.com/realiotech/realio-network/pull/126): Bump multistaking module version to v1.0.0-rc6
- (chore) [#125](https://github.com/realiotech/realio-network/pull/125): Fix lint errors
- (deps) Bump cosmos-sdk version to `v0.46.12`
- (tests) [#85](https://github.com/realiotech/realio-network/pull/85) Add test to calculate expect mint amount each block

### Deprecated

- Unfork module `x/staking`
- (chore) [#84](https://github.com/realiotech/realio-network/pull/84)  Remove useless param from create-token cmd

## [v0.8.3] - 2023-07-10

### Bug Fixes

- (x/asset) [#80](https://github.com/realiotech/realio-network/pull/80) Update transfer token to use bank keepers BlockAddrs
- (x/asset) [#83](https://github.com/realiotech/realio-network/pull/83) Create token will check for denom existence in bank module state

## [v0.8.2] - 2023-06-9

### Bug Fixes

- (deps) [#79](https://github.com/realiotech/realio-network/pull/79) Barberry patch. Bump cosmos-sdk version to `v0.46.11-realio-4`.

## [v0.8.1] - 2023-06-1

### Improvements

- (deps) [#76](https://github.com/realioteach/realio-network/pull/76) Bump IBC-go version to v6.1.1

### Bug Fixes

- (deps) [#77](https://github.com/realiotech/realio-network/pull/77) Bump cosmos-sdk version to `v0.46.11-realio-3`.
  Modify redelegation logic in `x/staking` module.

## [v0.8.0-rc4] - 2023-04-2

### State Machine Breaking
- (asset) [9c78be6](https://github.com/realiotech/realio-network/commit/9c78be67e8fc06997c07a5c84559d41f67cf196f) x/asset modify token whitelist storage. add restriction module whitelist into assetKeeper
- (asset) [6529b19](https://github.com/realiotech/realio-network/commit/6529b19cba0b7abfefb5d476c628a1fe4224f5e5) x/asset add restriction support into bank keeper. clean up issuance logic
- (proto) [75f19ff](https://github.com/realiotech/realio-network/commit/75f19ff86aeff854fa853f4e06d5f72cb3193324) x/asset token model update, add query support for token

### API Breaking

### Features

### Improvements
- (deps) [fffc39](https://github.com/realiotech/realio-network/commit/fffc39c10369ae12691d58dd936d0d7f481dc486) migrate ethermint coin type

### Bug Fixes
## [v0.7.1] - 2023-01-24

### State Machine Breaking

- (deps) [6bbb25](https://github.com/realiotech/realio-network/commit/6bbb2584e1d855dba77cde49a415fd4dba282cb5) Bump `cosmos sdk` to [`v0.46.7`](https://github.com/realiotech/cosmos-sdk/releases/tag/v0.46.x-realio-alpha-0.6)
- (deps) [6bbb25](https://github.com/realiotech/realio-network/commit/6bbb2584e1d855dba77cde49a415fd4dba282cb5) Bump `ethermint` to [`v0.21.0-rc2`](https://github.com/evmos/ethermint/releases/tag/v0.21.0-rc1)
- (deps) [6bbb25](https://github.com/realiotech/realio-network/commit/6bbb2584e1d855dba77cde49a415fd4dba282cb5) Bump `ibc-go` to [`v6.1.0`](https://github.com/cosmos/ibc-go/releases/tag/v6.1.0)

### Bug Fixes

- (upgrade) [6bbb25](https://github.com/realiotech/realio-network/commit/6bbb2584e1d855dba77cde49a415fd4dba282cb5) Fix Ethermint params upgrade
