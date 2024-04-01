package keeper_test

import (
	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (suite *KeeperTestSuite) TestAddMultiStakingCoinProposal() {
	bondWeight := sdk.NewDec(1)

	for _, tc := range []struct {
		desc      string
		malleate  func(p *types.AddMultiStakingCoinProposal)
		proposal  *types.AddMultiStakingCoinProposal
		shouldErr bool
	}{
		{
			desc: "Success",
			malleate: func(p *types.AddMultiStakingCoinProposal) {
				_, found := suite.msKeeper.GetBondWeight(suite.ctx, p.Denom)
				suite.Require().False(found)
			},
			proposal: &types.AddMultiStakingCoinProposal{
				Title:       "Add multistaking coin",
				Description: "Add new multistaking coin",
				Denom:       "stake1",
				BondWeight:  &bondWeight,
			},
			shouldErr: false,
		},
		{
			desc: "Error multistaking coin already exists",
			malleate: func(p *types.AddMultiStakingCoinProposal) {
				suite.msKeeper.SetBondWeight(suite.ctx, p.Denom, *p.BondWeight)
			},
			proposal: &types.AddMultiStakingCoinProposal{
				Title:       "Add multistaking coin",
				Description: "Add new multistaking coin",
				Denom:       "stake2",
				BondWeight:  &bondWeight,
			},
			shouldErr: true,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			suite.SetupTest()
			tc.malleate(tc.proposal)

			legacyProposal, err := govv1types.NewLegacyContent(tc.proposal, authtypes.NewModuleAddress(govtypes.ModuleName).String())
			suite.Require().NoError(err)

			if !tc.shouldErr {
				// store proposal
				_, err = suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{legacyProposal}, "")
				suite.Require().NoError(err)

				// execute proposal
				handler := suite.govKeeper.LegacyRouter().GetRoute(tc.proposal.ProposalRoute())
				err = handler(suite.ctx, tc.proposal)
				suite.Require().NoError(err)

				_, found := suite.msKeeper.GetBondWeight(suite.ctx, tc.proposal.Denom)
				suite.Require().True(found)
			} else {
				// store proposal
				_, err = suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{legacyProposal}, "")
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateBondWeightProposal() {
	bondWeight := sdk.NewDec(1)

	for _, tc := range []struct {
		desc      string
		malleate  func(p *types.UpdateBondWeightProposal)
		proposal  *types.UpdateBondWeightProposal
		shouldErr bool
	}{
		{
			desc: "Success",
			malleate: func(p *types.UpdateBondWeightProposal) {
				oldBondWeight := sdk.NewDec(2)
				suite.msKeeper.SetBondWeight(suite.ctx, p.Denom, oldBondWeight)
			},
			proposal: &types.UpdateBondWeightProposal{
				Title:             "Add multistaking coin",
				Description:       "Add new multistaking coin",
				Denom:             "stake1",
				UpdatedBondWeight: &bondWeight,
			},
			shouldErr: false,
		},
		{
			desc:     "Error multistaking coin not exists",
			malleate: func(p *types.UpdateBondWeightProposal) {},
			proposal: &types.UpdateBondWeightProposal{
				Title:             "Add multistaking coin",
				Description:       "Add new multistaking coin",
				Denom:             "stake2",
				UpdatedBondWeight: &bondWeight,
			},
			shouldErr: true,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			suite.SetupTest()
			tc.malleate(tc.proposal)

			legacyProposal, err := govv1types.NewLegacyContent(tc.proposal, authtypes.NewModuleAddress(govtypes.ModuleName).String())
			suite.Require().NoError(err)

			if !tc.shouldErr {
				// store proposal
				_, err = suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{legacyProposal}, "")
				suite.Require().NoError(err)

				// execute proposal
				handler := suite.govKeeper.LegacyRouter().GetRoute(tc.proposal.ProposalRoute())
				err = handler(suite.ctx, tc.proposal)
				suite.Require().NoError(err)

				weight, found := suite.msKeeper.GetBondWeight(suite.ctx, tc.proposal.Denom)
				suite.Require().True(found)
				suite.Require().True(weight.Equal(bondWeight))
			} else {
				// store proposal
				_, err = suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{legacyProposal}, "")
				suite.Require().Error(err)
			}
		})
	}
}
