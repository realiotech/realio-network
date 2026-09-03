package app

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/realiotech/realio-network/testutil"
)

func TestDisableLeakedEVMFeeSponsor(t *testing.T) {
	realio := Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1})
	leakedSponsor := testutil.GenAddress()
	realio.FeeSponsorKeeper.SetFeePayerToStore(ctx, leakedSponsor)

	disableLeakedEVMFeeSponsor(realio, ctx, []sdk.AccAddress{leakedSponsor})

	_, found := realio.FeeSponsorKeeper.GetFeePayer(ctx)
	require.False(t, found)
}
