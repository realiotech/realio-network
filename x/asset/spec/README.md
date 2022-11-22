<!--
order: 0
title: Mint Overview
parent:
  title: "mint"
-->

# `mint`

## Abstract

The `x/asset` module enables the creation and management of on chain assets in the Realio Network.

With this module, you can create assets that represent digitally native and real-world assets such as security tokens and stablecoins. 
There is functionality to place transfer restrictions via whitelists on an asset that help support securities, compliance, and certification use cases.


## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
    * [Minter](02_state.md#minter)
    * [Params](02_state.md#params)
3. **[Begin-Block](03_begin_block.md)**
    * [NextAnnualProvisions](03_begin_block.md#nextannualprovisions)
    * [BlockProvision](03_begin_block.md#blockprovision)
4. **[Parameters](04_params.md)**
5. **[Events](05_events.md)**
    * [BeginBlocker](05_events.md#beginblocker)
6. **[Client](06_client.md)**
    * [CLI](06_client.md#cli)
    * [gRPC](06_client.md#grpc)
    * [REST](06_client.md#rest)
