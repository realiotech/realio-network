package app

import (
	"fmt"

	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// revokeLeakedFeeGrants removes allowances funded by leaked accounts. The
// ante handler also rejects a blacklisted FeeGranter, but deleting the grants
// prevents them from becoming usable again if an address is later removed
// from the blacklist.
func revokeLeakedFeeGrants(app *RealioNetwork, ctx sdk.Context, leaked []sdk.AccAddress) {
	leakedSet := make(map[string]struct{}, len(leaked))
	for _, addr := range leaked {
		leakedSet[addr.String()] = struct{}{}
	}

	var toRevoke []feegrant.Grant
	if err := app.FeeGrantKeeper.IterateAllFeeAllowances(ctx, func(grant feegrant.Grant) bool {
		if _, leaked := leakedSet[grant.Granter]; leaked {
			toRevoke = append(toRevoke, grant)
		}
		return false
	}); err != nil {
		panic(fmt.Errorf("revoke leaked fee grants: failed to iterate allowances: %w", err))
	}

	msgServer := feegrantkeeper.NewMsgServerImpl(app.FeeGrantKeeper)
	for _, grant := range toRevoke {
		_, err := msgServer.RevokeAllowance(ctx, &feegrant.MsgRevokeAllowance{
			Granter: grant.Granter,
			Grantee: grant.Grantee,
		})
		if err != nil {
			panic(fmt.Errorf("revoke leaked fee grant: granter=%s grantee=%s: %w", grant.Granter, grant.Grantee, err))
		}
	}
}

// disableLeakedEVMFeeSponsor prevents the global EVM sponsor path from
// spending a leaked account's feegrant allowance on behalf of clean senders.
func disableLeakedEVMFeeSponsor(app *RealioNetwork, ctx sdk.Context, leaked []sdk.AccAddress) {
	feePayer, found := app.FeeSponsorKeeper.GetFeePayer(ctx)
	if !found {
		return
	}

	for _, addr := range leaked {
		if addr.Equals(sdk.AccAddress(feePayer)) {
			app.FeeSponsorKeeper.RemoveFeePayerFromStore(ctx)
			return
		}
	}
}
