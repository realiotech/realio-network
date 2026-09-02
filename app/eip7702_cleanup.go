package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// clearLeakedEIP7702Delegations removes delegation designators already
// installed on leaked EOAs before the fork. The ante decorator prevents new
// authorizations after restart, but without this cleanup a clean relayer could
// continue invoking pre-existing delegated wallet code at the leaked address.
func clearLeakedEIP7702Delegations(app *RealioNetwork, ctx sdk.Context, leaked []sdk.AccAddress) {
	for _, addr := range leaked {
		evmAddr := common.BytesToAddress(addr.Bytes())
		codeHash := app.EvmKeeper.GetCodeHash(ctx, evmAddr)
		code := app.EvmKeeper.GetCode(ctx, codeHash)
		if _, delegated := ethtypes.ParseDelegation(code); delegated {
			// Only unlink this account from the code. The code blob is keyed by
			// hash and may be shared by another delegated account.
			app.EvmKeeper.DeleteCodeHash(ctx, evmAddr)
		}
	}
}
