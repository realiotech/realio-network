package app

import (
	"testing"

	"cosmossdk.io/x/feegrant"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/realiotech/realio-network/testutil"
	realionetworktypes "github.com/realiotech/realio-network/types"
)

func TestRevokeLeakedFeeGrants(t *testing.T) {
	realio := Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1})
	leakedGranter := testutil.GenAddress()
	cleanGranter := testutil.GenAddress()
	grantee := testutil.GenAddress()
	allowance := &feegrant.BasicAllowance{SpendLimit: sdk.NewCoins(sdk.NewInt64Coin(realionetworktypes.BaseDenom, 100))}

	require.NoError(t, realio.FeeGrantKeeper.GrantAllowance(ctx, leakedGranter, grantee, allowance))
	require.NoError(t, realio.FeeGrantKeeper.GrantAllowance(ctx, cleanGranter, grantee, allowance))

	revokeLeakedFeeGrants(realio, ctx, []sdk.AccAddress{leakedGranter})

	_, err := realio.FeeGrantKeeper.GetAllowance(ctx, leakedGranter, grantee)
	require.Error(t, err)
	remaining, err := realio.FeeGrantKeeper.GetAllowance(ctx, cleanGranter, grantee)
	require.NoError(t, err)
	require.NotNil(t, remaining)
}

func TestDisableLeakedEVMFeeSponsor(t *testing.T) {
	realio := Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1})
	leakedSponsor := testutil.GenAddress()
	realio.FeeSponsorKeeper.SetFeePayerToStore(ctx, leakedSponsor)

	disableLeakedEVMFeeSponsor(realio, ctx, []sdk.AccAddress{leakedSponsor})

	_, found := realio.FeeSponsorKeeper.GetFeePayer(ctx)
	require.False(t, found)
}
