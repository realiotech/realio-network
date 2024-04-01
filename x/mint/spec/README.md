<!--
order: 0
title: Mint Overview
parent:
  title: "mint"
-->

# `mint`

## Abstract

The `x/mint` module mints new Rio tokens every block according to the inflation parameter and rio distribution model. 

It is based & replaces the original cosmos/x/mint module. It removes the logic around dynamic inflation calculation 
from goal bonded, actual bonded, and instead calculates the provisions based on a parameritized inflation rate and 
remaining Rio supply.

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
