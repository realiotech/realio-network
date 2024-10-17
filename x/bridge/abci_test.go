package bridge_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/realiotech/realio-network/app"
)

func TestRateLimitTrigger(t *testing.T) {
	// init app
	realio := app.Setup(false, nil, 1)

	ver0 := realio.LastBlockHeight()
	// commit 10 blocks
	for i := int64(1); i <= 10; i++ {
		header := cmtproto.Header{
			Height:  ver0 + i,
			AppHash: realio.LastCommitID().Hash,
		}

		realio.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: header.Height,
		})
		realio.Commit()
	}

	require.Equal(t, ver0+10, realio.LastBlockHeight())

	ctx := realio.BaseApp.NewContext(true)
	epochInfo, err := realio.BridgeKeeper.EpochInfo.Get(ctx)
	require.NoError(t, err)
	require.True(t, epochInfo.EpochCountingStarted)
}
