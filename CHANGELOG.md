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

## Unreleased

### State Machine Breaking

### API Breaking

### Features

### Improvements

### Bug Fixes

## [v0.7.0] - 2023-01-16

### State Machine Breaking

- (deps) [6bbb25](https://github.com/realiotech/realio-network/commit/6bbb2584e1d855dba77cde49a415fd4dba282cb5) Bump `cosmos sdk` to [`v0.46.7`](https://github.com/realiotech/cosmos-sdk/releases/tag/v0.46.x-realio-alpha-0.6)
- (deps) [6bbb25](https://github.com/realiotech/realio-network/commit/6bbb2584e1d855dba77cde49a415fd4dba282cb5) Bump `ethermint` to [`v0.21.0-rc1`](https://github.com/evmos/ethermint/releases/tag/v0.21.0-rc1)
- (deps) [6bbb25](https://github.com/realiotech/realio-network/commit/6bbb2584e1d855dba77cde49a415fd4dba282cb5) Bump `ibc-go` to [`v6.1.0`](https://github.com/cosmos/ibc-go/releases/tag/v6.1.0)

### Bug Fixes

- (upgrade) [6bbb25](https://github.com/realiotech/realio-network/commit/6bbb2584e1d855dba77cde49a415fd4dba282cb5) Fix Ethermint params upgrade