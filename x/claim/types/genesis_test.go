package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/realiotech/realio-network/x/claim/types"
)

func addr(seed byte) string {
	b := make([]byte, 20)
	b[0] = seed
	return sdk.AccAddress(b).String()
}

func TestGenesisValidateHackedAmounts(t *testing.T) {
	ok := sdk.NewCoins(sdk.NewInt64Coin("ario", 10))

	testCases := []struct {
		name    string
		gs      types.GenesisState
		wantErr string
	}{
		{name: "none", gs: types.GenesisState{}},
		{
			name: "valid",
			gs:   types.GenesisState{HackedAmounts: []types.HackedAmount{{Address: addr(1), Amount: ok}, {Address: addr(2), Amount: ok}}},
		},
		{
			name:    "bad address",
			gs:      types.GenesisState{HackedAmounts: []types.HackedAmount{{Address: "nope", Amount: ok}}},
			wantErr: "invalid hacked_amounts address",
		},
		{
			name:    "empty amount",
			gs:      types.GenesisState{HackedAmounts: []types.HackedAmount{{Address: addr(1)}}},
			wantErr: "must not be empty",
		},
		{
			name:    "invalid amount",
			gs:      types.GenesisState{HackedAmounts: []types.HackedAmount{{Address: addr(1), Amount: sdk.Coins{sdk.Coin{Denom: "bbb", Amount: math.NewInt(1)}, sdk.Coin{Denom: "aaa", Amount: math.NewInt(1)}}}}},
			wantErr: "invalid hacked_amounts amount",
		},
		{
			name:    "duplicate address",
			gs:      types.GenesisState{HackedAmounts: []types.HackedAmount{{Address: addr(1), Amount: ok}, {Address: addr(1), Amount: ok}}},
			wantErr: "duplicate hacked_amounts address",
		},
		{
			name:    "the attacker is owed nothing",
			gs:      types.GenesisState{Attacker: addr(9), HackedAmounts: []types.HackedAmount{{Address: addr(9), Amount: ok}}},
			wantErr: "is the attacker",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.gs.Validate()
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
