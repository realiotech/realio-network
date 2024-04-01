package types_test

import (
	"testing"

	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type ProposalTestSuite struct {
	suite.Suite
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalTestSuite))
}

func (suite *ProposalTestSuite) TestKeysTypes() {
	suite.Require().Equal("multistaking", (&types.AddMultiStakingCoinProposal{}).ProposalRoute())
	suite.Require().Equal("AddMultiStakingCoin", (&types.AddMultiStakingCoinProposal{}).ProposalType())
	suite.Require().Equal("multistaking", (&types.UpdateBondWeightProposal{}).ProposalRoute())
	suite.Require().Equal("UpdateBondWeight", (&types.UpdateBondWeightProposal{}).ProposalType())
}

func (suite *ProposalTestSuite) TestProposalString() {
	testTokenWeight := sdk.OneDec()
	testCases := []struct {
		msg           string
		proposal      govv1beta1.Content
		expectedValue string
	}{
		{
			msg: "Add Bond Token Proposal", proposal: &types.AddMultiStakingCoinProposal{Denom: "token", BondWeight: &testTokenWeight, Description: "Add token", Title: "Add #1"},
			expectedValue: "AddMultiStakingCoinProposal: Title: Add #1 Description: Add token Denom: token TokenWeight: 1.000000000000000000",
		},

		{
			msg: "Change Bond Token Weight Proposal", proposal: &types.UpdateBondWeightProposal{Denom: "token", UpdatedBondWeight: &testTokenWeight, Description: "Change Bond token weight", Title: "Change #2"},
			expectedValue: "UpdateBondWeightProposal: Title: Change #2 Description: Change Bond token weight Denom: token TokenWeight: 1.000000000000000000",
		},
	}

	for _, tc := range testCases {
		str_result := tc.proposal.String()
		suite.Require().Equal(str_result, tc.expectedValue)
	}
}

func (suite *ProposalTestSuite) TestAddMultiStakingCoinProposal() {
	testCases := []struct {
		msg         string
		title       string
		description string
		denom       string
		bondWeight  sdk.Dec
		expectPass  bool
	}{
		// Valid tests
		{msg: "Add bond token", title: "test", description: "test desc", denom: "token", bondWeight: sdk.OneDec(), expectPass: true},

		// Invalid tests
		{msg: "Add bond token - invalid token", title: "test", description: "test desc", denom: "", bondWeight: sdk.OneDec(), expectPass: false},
		{msg: "Add bond token - negative weight", title: "test", description: "test desc", denom: "token", bondWeight: sdk.MustNewDecFromStr("-1"), expectPass: false},
		{msg: "Add bond token - zero weight", title: "test", description: "test desc", denom: "token", bondWeight: sdk.ZeroDec(), expectPass: false},
	}

	for i, tc := range testCases {
		tx := types.NewAddMultiStakingCoinProposal(tc.title, tc.description, tc.denom, tc.bondWeight)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
		}
	}
}

func (suite *ProposalTestSuite) TestUpdateBondWeightProposal() {
	testCases := []struct {
		msg         string
		title       string
		description string
		denom       string
		bondWeight  sdk.Dec
		expectPass  bool
	}{
		// Valid tests
		{msg: "Change bond token weight", title: "test", description: "test desc", denom: "token", bondWeight: sdk.OneDec(), expectPass: true},

		// Invalid tests
		{msg: "Change bond token weight - invalid token", title: "test", description: "test desc", denom: "", bondWeight: sdk.OneDec(), expectPass: false},
		{msg: "Change bond token weight - negative weight", title: "test", description: "test desc", denom: "token", bondWeight: sdk.MustNewDecFromStr("-1"), expectPass: false},
		{msg: "Change bond token weight - zero weight", title: "test", description: "test desc", denom: "token", bondWeight: sdk.ZeroDec(), expectPass: false},
	}

	for i, tc := range testCases {
		tx := types.NewUpdateBondWeightProposal(tc.title, tc.description, tc.denom, tc.bondWeight)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
		}
	}
}
