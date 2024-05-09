package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyMintDenom     = []byte("MintDenom")
	KeyInflationRate = []byte("InflationRate")
	KeyBlocksPerYear = []byte("BlocksPerYear")
)

// ParamTable for minting module.
//
// NOTE: Deprecated.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}


// Implements params.ParamSet
//
// NOTE: Deprecated.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyInflationRate, &p.InflationRate, validateInflationRate),
		paramtypes.NewParamSetPair(KeyBlocksPerYear, &p.BlocksPerYear, validateBlocksPerYear),
	}
}