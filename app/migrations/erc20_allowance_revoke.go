package migrations

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

// revokeLeakedERC20Allowances removes every allowance in x/erc20's native
// precompile store whose owner is leaked. Runtime protection is provided by
// the EVM token blacklist hook; this cleanup also prevents stale allowances
// from becoming usable if an address is later removed from the blacklist.
func revokeLeakedERC20Allowances(k Keepers, ctx sdk.Context, leaked []sdk.AccAddress) {
	leakedSet := make(map[string]struct{}, len(leaked))
	for _, addr := range leaked {
		leakedSet[common.BytesToAddress(addr.Bytes()).Hex()] = struct{}{}
	}

	allowances := k.Erc20Keeper.GetAllowances(ctx)
	for _, allowance := range allowances {
		owner := common.HexToAddress(allowance.Owner)
		if _, leaked := leakedSet[owner.Hex()]; !leaked {
			continue
		}

		if err := k.Erc20Keeper.UnsafeSetAllowance(
			ctx,
			common.HexToAddress(allowance.Erc20Address),
			owner,
			common.HexToAddress(allowance.Spender),
			common.Big0,
		); err != nil {
			panic(fmt.Errorf("revoke leaked ERC20 allowance: token=%s owner=%s spender=%s: %w",
				allowance.Erc20Address, allowance.Owner, allowance.Spender, err))
		}
	}
}
