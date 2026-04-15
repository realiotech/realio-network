package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

// NewParams returns Params instance with the given values.
func NewParams(authority string) Params {
	return Params{
		Authority: authority,
	}
}

// default bridge module parameters
func DefaultParams() Params {
	return Params{
		Authority: "",
	}
}

// validate params
func (p Params) Validate() error {
	if p.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(p.Authority); err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
