package app

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/testutil"
)

func TestEVMTokenBlacklistHook(t *testing.T) {
	realio := Setup(false, nil, 1)
	ctx := realio.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1})
	blacklistedOwner := testutil.GenAddress()
	cleanOwner := testutil.GenAddress()
	require.NoError(t, realio.BlacklistKeeper.SetBlacklisted(ctx, blacklistedOwner))

	hook := NewEVMTokenBlacklistHook(realio.BlacklistKeeper)
	for _, tc := range []struct {
		name      string
		topic     common.Hash
		owner     sdk.AccAddress
		wantError bool
	}{
		{"transfer from blacklisted owner", transferEventTopic, blacklistedOwner, true},
		{"approval by blacklisted owner", approvalEventTopic, blacklistedOwner, true},
		{"approval-for-all by blacklisted owner", approvalForAllEventTopic, blacklistedOwner, true},
		{"transfer from clean owner", transferEventTopic, cleanOwner, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			receipt := &ethtypes.Receipt{Logs: []*ethtypes.Log{{
				Address: common.HexToAddress("0x0000000000000000000000000000000000001234"),
				Topics: []common.Hash{
					tc.topic,
					common.BytesToHash(tc.owner.Bytes()),
					common.BytesToHash(cleanOwner.Bytes()),
				},
			}}}

			err := hook.PostTxProcessing(ctx, common.Address{}, core.Message{}, receipt)
			if tc.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.owner.String())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEVMTokenBlacklistHookRegistered(t *testing.T) {
	realio := Setup(false, nil, 1)
	require.True(t, realio.EvmKeeper.HasHooks())
}
