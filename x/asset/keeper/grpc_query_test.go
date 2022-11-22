package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/asset/types"
)

func (suite *KeeperTestSuite) TestParamsQuery() {
	suite.SetupTest()

	k := suite.app.AssetKeeper
	wctx := sdk.WrapSDKContext(suite.ctx)

	params := types.DefaultParams()
	k.SetParams(suite.ctx, params)

	response, err := k.Params(wctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}

func (suite *KeeperTestSuite) TestTokensQuery() {
	var (
		req    *types.QueryTokensRequest
		expRes *types.QueryTokensResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no tokens",
			func() {
				req = &types.QueryTokensRequest{}
				expRes = &types.QueryTokensResponse{Tokens: []types.Token(nil)}
			},
			true,
		},
		{
			"1 token exists",
			func() {
				req = &types.QueryTokensRequest{}
				token := types.Token{
					Creator:               suite.testUser1Address,
					Name:                  "rst",
					Symbol:                "rst",
					Total:                 "1000",
					Decimals:              "1",
					AuthorizationRequired: false,
				}
				suite.app.AssetKeeper.SetToken(suite.ctx, token)

				expRes = &types.QueryTokensResponse{
					Tokens: []types.Token{token},
				}
			},
			true,
		},
		{
			"2 tokens exists",
			func() {
				req = &types.QueryTokensRequest{}
				token1 := types.Token{
					Creator:               suite.testUser1Address,
					Name:                  "rst",
					Symbol:                "rst",
					Total:                 "1000",
					Decimals:              "1",
					AuthorizationRequired: false,
				}
				suite.app.AssetKeeper.SetToken(suite.ctx, token1)

				token2 := types.Token{
					Creator:               suite.testUser1Address,
					Name:                  "bitcoinEtf",
					Symbol:                "btf",
					Total:                 "1000",
					Decimals:              "1",
					AuthorizationRequired: false,
				}
				suite.app.AssetKeeper.SetToken(suite.ctx, token2)

				expRes = &types.QueryTokensResponse{
					Tokens: []types.Token{token2, token1},
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.Tokens(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.Tokens, res.Tokens)
				suite.Require().ElementsMatch(expRes.Tokens, res.Tokens)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
