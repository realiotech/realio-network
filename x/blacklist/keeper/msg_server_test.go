package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/testutil"
	"github.com/realiotech/realio-network/x/blacklist/keeper"
	"github.com/realiotech/realio-network/x/blacklist/types"
)

func (suite *KeeperTestSuite) TestUpdateBlacklist() {
	addrA := testutil.GenAddress()
	addrB := testutil.GenAddress()
	badAuthority := testutil.GenAddress().String()

	testCases := []struct {
		name         string
		msg          *types.MsgUpdateBlacklist
		setAuthority bool
		expectErr    bool
		errString    string
	}{
		{
			name: "valid: add addresses",
			msg: &types.MsgUpdateBlacklist{
				AddAddresses: []string{addrA.String(), addrB.String()},
			},
			setAuthority: true,
			expectErr:    false,
		},
		{
			name: "invalid: bad bech32 in add_addresses",
			msg: &types.MsgUpdateBlacklist{
				AddAddresses: []string{"not-a-valid-address"},
			},
			setAuthority: true,
			expectErr:    true,
			errString:    "invalid add_addresses entry",
		},
		{
			name: "invalid: bad bech32 in remove_addresses",
			msg: &types.MsgUpdateBlacklist{
				RemoveAddresses: []string{"not-a-valid-address"},
			},
			setAuthority: true,
			expectErr:    true,
			errString:    "invalid remove_addresses entry",
		},
		{
			name: "invalid: wrong authority",
			msg: &types.MsgUpdateBlacklist{
				Authority:    badAuthority,
				AddAddresses: []string{addrA.String()},
			},
			setAuthority: false,
			expectErr:    true,
			errString:    "invalid authority",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			srv := keeper.NewMsgServerImpl(suite.app.BlacklistKeeper)

			if tc.setAuthority {
				tc.msg.Authority = suite.authority
			}

			_, err := srv.UpdateBlacklist(suite.ctx, tc.msg)
			if tc.expectErr {
				suite.Require().ErrorContains(err, tc.errString)
				return
			}

			suite.Require().NoError(err)
			for _, addr := range tc.msg.AddAddresses {
				accAddr, decodeErr := sdk.AccAddressFromBech32(addr)
				suite.Require().NoError(decodeErr)
				suite.Require().True(suite.app.BlacklistKeeper.IsBlacklisted(suite.ctx, accAddr))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateBlacklistRemove() {
	suite.SetupTest()

	addr := testutil.GenAddress()
	srv := keeper.NewMsgServerImpl(suite.app.BlacklistKeeper)

	_, err := srv.UpdateBlacklist(suite.ctx, &types.MsgUpdateBlacklist{
		Authority:    suite.authority,
		AddAddresses: []string{addr.String()},
	})
	suite.Require().NoError(err)
	suite.Require().True(suite.app.BlacklistKeeper.IsBlacklisted(suite.ctx, addr))

	_, err = srv.UpdateBlacklist(suite.ctx, &types.MsgUpdateBlacklist{
		Authority:       suite.authority,
		RemoveAddresses: []string{addr.String()},
	})
	suite.Require().NoError(err)
	suite.Require().False(suite.app.BlacklistKeeper.IsBlacklisted(suite.ctx, addr))
}
