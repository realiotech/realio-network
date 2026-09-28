package v8

import "testing"

// TestRotationsFor covers the chain-id branching that picks
// TestnetValidatorRotations over ValidatorRotations (mainnet): this is
// package-internal, standalone logic with no keeper/app dependency, so it's
// tested directly rather than through the full RotateValidators path.
func TestRotationsFor(t *testing.T) {
	cases := []struct {
		name    string
		chainID string
		want    []ValidatorRotation
	}{
		{
			name:    "mainnet",
			chainID: "realionetwork_3301-1",
			want:    ValidatorRotations,
		},
		{
			name:    "testnet",
			chainID: "realionetwork_3300-1",
			want:    TestnetValidatorRotations,
		},
		{
			name:    "some other chain-id (e.g. a unit test's own) -> mainnet, not testnet",
			chainID: "realionetworklocal_7777-1",
			want:    ValidatorRotations,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rotationsFor(tc.chainID)
			if len(got) != len(tc.want) {
				t.Fatalf("rotationsFor(%q) returned %d entries, want %d", tc.chainID, len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("rotationsFor(%q)[%d] = %+v, want %+v", tc.chainID, i, got[i], tc.want[i])
				}
			}
		})
	}
}
