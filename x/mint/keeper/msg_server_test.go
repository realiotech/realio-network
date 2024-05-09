package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/x/mint/types"
)

func (suite *KeeperTestSuite) TestUpdateParams() {
	testCases := []struct {
		name      string
		request   *types.MsgUpdateParams
		expectErr bool
	}{
		{
			name: "set invalid authority",
			request: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
		},
		{
			name: "set invalid params",
			request: &types.MsgUpdateParams{
				Authority: suite.app.MintKeeper.GetAuthority(),
				Params: types.Params{
					MintDenom:     sdk.DefaultBondDenom,
					InflationRate: sdk.NewDecWithPrec(-13, 2),
					BlocksPerYear: uint64(60 * 60 * 8766 / 5),
				},
			},
			expectErr: true,
		},
		{
			name: "set full valid params",
			request: &types.MsgUpdateParams{
				Authority: suite.app.MintKeeper.GetAuthority(),
				Params: types.Params{
					MintDenom:     sdk.DefaultBondDenom,
					InflationRate: sdk.NewDecWithPrec(8, 2),
					BlocksPerYear: uint64(60 * 60 * 8766 / 5),
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.DoSetupTest(suite.T())

			_, err := suite.msgServer.UpdateParams(suite.ctx, tc.request)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
