package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	keepertest "github.com/realiotech/realio-network/testutil/keeper"
	"github.com/realiotech/realio-network/x/rststaking/keeper"
	"github.com/realiotech/realio-network/x/rststaking/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestRstStakeMsgServerCreate(t *testing.T) {
	k, ctx := keepertest.RststakingKeeper(t)
	srv := keeper.NewMsgServerImpl(*k)
	wctx := sdk.WrapSDKContext(ctx)
	creator := "A"
	for i := 0; i < 5; i++ {
		expected := &types.MsgCreateRstStake{Creator: creator,
			Id: strconv.Itoa(i),
		}
		_, err := srv.CreateRstStake(wctx, expected)
		require.NoError(t, err)
		rst, found := k.GetRstStake(ctx,
			expected.Id,
		)
		require.True(t, found)
		require.Equal(t, expected.Creator, rst.Creator)
	}
}

func TestRstStakeMsgServerUpdate(t *testing.T) {
	creator := "A"

	for _, tc := range []struct {
		desc    string
		request *types.MsgUpdateRstStake
		err     error
	}{
		{
			desc: "Completed",
			request: &types.MsgUpdateRstStake{Creator: creator,
				Id: strconv.Itoa(0),
			},
		},
		{
			desc: "Unauthorized",
			request: &types.MsgUpdateRstStake{Creator: "B",
				Id: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "KeyNotFound",
			request: &types.MsgUpdateRstStake{Creator: creator,
				Id: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			k, ctx := keepertest.RststakingKeeper(t)
			srv := keeper.NewMsgServerImpl(*k)
			wctx := sdk.WrapSDKContext(ctx)
			expected := &types.MsgCreateRstStake{Creator: creator,
				Id: strconv.Itoa(0),
			}
			_, err := srv.CreateRstStake(wctx, expected)
			require.NoError(t, err)

			_, err = srv.UpdateRstStake(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				rst, found := k.GetRstStake(ctx,
					expected.Id,
				)
				require.True(t, found)
				require.Equal(t, expected.Creator, rst.Creator)
			}
		})
	}
}

func TestRstStakeMsgServerDelete(t *testing.T) {
	creator := "A"

	for _, tc := range []struct {
		desc    string
		request *types.MsgDeleteRstStake
		err     error
	}{
		{
			desc: "Completed",
			request: &types.MsgDeleteRstStake{Creator: creator,
				Id: strconv.Itoa(0),
			},
		},
		{
			desc: "Unauthorized",
			request: &types.MsgDeleteRstStake{Creator: "B",
				Id: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "KeyNotFound",
			request: &types.MsgDeleteRstStake{Creator: creator,
				Id: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			k, ctx := keepertest.RststakingKeeper(t)
			srv := keeper.NewMsgServerImpl(*k)
			wctx := sdk.WrapSDKContext(ctx)

			_, err := srv.CreateRstStake(wctx, &types.MsgCreateRstStake{Creator: creator,
				Id: strconv.Itoa(0),
			})
			require.NoError(t, err)
			_, err = srv.DeleteRstStake(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				_, found := k.GetRstStake(ctx,
					tc.request.Id,
				)
				require.False(t, found)
			}
		})
	}
}
