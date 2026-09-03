package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
