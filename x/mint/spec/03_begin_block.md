<!--
order: 3
-->

# Begin-Block

Minting inflation is paid at the beginning of each block.

## Inflation rate calculation

Inflation rate is set at genesis time and available via the params for the module. The rate is changeable via
a governance proposal


## NextAnnualProvisions

Calculate the annual provisions based on current remaining total supply of rio and inflation
rate. This parameter is calculated once per block.

```go
NextAnnualProvisions(params Params, totalSupply sdk.Dec) (provisions sdk.Dec) {
	return Inflation * remainingRioTotalSupply
```

## BlockProvision

Calculate the provisions generated for each block based on current annual provisions. The provisions are then minted by the `mint` module's `ModuleMinterAccount` and then transferred to the `auth`'s `FeeCollector` `ModuleAccount`.

```go
BlockProvision(params Params) sdk.Coin {
	provisionAmt = AnnualProvisions/ params.BlocksPerYear
	return sdk.NewCoin(params.MintDenom, provisionAmt.Truncate())
```
