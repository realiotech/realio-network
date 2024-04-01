package types

import (
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/codec"
)

// String implements the Stringer interface for a Validator object.
func (v ValidatorInfo) String() string {
	bz, err := codec.ProtoMarshalJSON(&v, nil)
	if err != nil {
		panic(err)
	}

	out, err := yaml.JSONToYAML(bz)
	if err != nil {
		panic(err)
	}

	return string(out)
}
