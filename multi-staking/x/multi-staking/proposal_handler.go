package multistaking

import (
	"github.com/realio-tech/multi-staking-module/x/multi-staking/client/cli"
	"github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	AddMultiStakingProposalHandler  = govclient.NewProposalHandler(cli.NewCmdSubmitAddMultiStakingCoinProposal)
	UpdateBondWeightProposalHandler = govclient.NewProposalHandler(cli.NewCmdUpdateBondWeightProposal)
)

// NewMultiStakingProposalHandler creates a governance handler to manage Mult-Staking proposals.
func NewMultiStakingProposalHandler(k *keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.AddMultiStakingCoinProposal:
			return k.AddMultiStakingCoinProposal(ctx, c)
		case *types.UpdateBondWeightProposal:
			return k.BondWeightProposal(ctx, c)
		default:
			return sdkerrors.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}
