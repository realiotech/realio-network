package bridge_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/realiotech/realio-network/app"
	"github.com/realiotech/realio-network/x/bridge/types"
)

func TestRateLimitTrigger(t *testing.T) {
	// init app
	realio := app.Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContext(true)
	err := realio.BridgeKeeper.RegisteredCoins.Set(ctx, app.MultiStakingCoinA.Denom, types.RateLimit{
		Ratelimit:     math.NewInt(1000000000),
		CurrentInflow: math.NewInt(100000000),
	})
	require.NoError(t, err)
	_, err = realio.BridgeKeeper.RegisteredCoins.Get(ctx, app.MultiStakingCoinA.Denom)
	require.NoError(t, err)

	ver0 := realio.LastBlockHeight()
	// commit 10 blocks
	for i := int64(1); i <= 10; i++ {
		header := cmtproto.Header{
			Height:  ver0 + i,
			AppHash: realio.LastCommitID().Hash,
		}

		_, err = realio.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: header.Height,
		})
		require.NoError(t, err)
		_, err = realio.Commit()
		require.NoError(t, err)
	}

	require.Equal(t, ver0+10, realio.LastBlockHeight())

	ctx = realio.BaseApp.NewContext(true)
	epochInfo, err := realio.BridgeKeeper.EpochInfo.Get(ctx)
	require.NoError(t, err)
	require.True(t, epochInfo.EpochCountingStarted)
	_, err = realio.BridgeKeeper.RegisteredCoins.Get(ctx, app.MultiStakingCoinA.Denom)
	require.NoError(t, err)
}
