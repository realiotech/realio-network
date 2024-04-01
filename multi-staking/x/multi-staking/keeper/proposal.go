package keeper

import (
	"fmt"

	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddMultiStakingCoinProposal handles the proposals to add a new bond token
func (k Keeper) AddMultiStakingCoinProposal(
	ctx sdk.Context,
	p *types.AddMultiStakingCoinProposal,
) error {
	_, found := k.GetBondWeight(ctx, p.Denom)
	if found {
		return fmt.Errorf("Error MultiStakingCoin %s already exist", p.Denom) //nolint:stylecheck
	}

	bondWeight := *p.BondWeight
	if bondWeight.LTE(sdk.ZeroDec()) {
		return fmt.Errorf("Error MultiStakingCoin BondWeight %s invalid", bondWeight) //nolint:stylecheck
	}

	k.SetBondWeight(ctx, p.Denom, bondWeight)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddMultiStakingCoin,
			sdk.NewAttribute(types.AttributeKeyDenom, p.Denom),
			sdk.NewAttribute(types.AttributeKeyBondWeight, p.BondWeight.String()),
		),
	)
	return nil
}

func (k Keeper) BondWeightProposal(
	ctx sdk.Context,
	p *types.UpdateBondWeightProposal,
) error {
	_, found := k.GetBondWeight(ctx, p.Denom)
	if !found {
		return fmt.Errorf("Error MultiStakingCoin %s not found", p.Denom) //nolint:stylecheck
	}

	bondWeight := *p.UpdatedBondWeight
	if bondWeight.LTE(sdk.ZeroDec()) {
		return fmt.Errorf("Error MultiStakingCoin BondWeight %s invalid", bondWeight) //nolint:stylecheck
	}

	k.SetBondWeight(ctx, p.Denom, bondWeight)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddMultiStakingCoin,
			sdk.NewAttribute(types.AttributeKeyDenom, p.Denom),
			sdk.NewAttribute(types.AttributeKeyBondWeight, p.UpdatedBondWeight.String()),
		),
	)
	return nil
}
