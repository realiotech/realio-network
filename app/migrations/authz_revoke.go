package migrations

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// revokeLeakedAuthzGrants deletes every authz grant where the GRANTER is a
// leaked address. Blacklisting alone doesn't cover this: CosmosBlacklistDecorator
// (app/ante/blacklist.go) checks sigTx.GetSigners(), which for an
// authz.MsgExec is only the grantee — the inner messages' signers are never
// inspected. So a leaked account with an outstanding authz grant (created by
// the attacker before the halt, using the leaked key) lets its grantee keep
// acting on its behalf — e.g. sending funds or staking — via MsgExec, fully
// bypassing the blacklist. Revoking every grant with a leaked granter here
// closes that regardless of whether the ante handler is ever taught to
// recurse into MsgExec.
func revokeLeakedAuthzGrants(k Keepers, ctx sdk.Context, leaked []sdk.AccAddress) {
	leakedSet := make(map[string]bool, len(leaked))
	for _, accAddr := range leaked {
		leakedSet[accAddr.String()] = true
	}

	type grantToRevoke struct {
		granter, grantee sdk.AccAddress
		msgType          string
	}
	var toRevoke []grantToRevoke
	k.AuthzKeeper.IterateGrants(ctx, func(granterAddr, granteeAddr sdk.AccAddress, grant authz.Grant) bool {
		if !leakedSet[granterAddr.String()] {
			return false
		}
		auth, err := grant.GetAuthorization()
		if err != nil {
			panic(fmt.Errorf("revoke leaked authz grants: failed to unpack authorization for granter %s: %w", granterAddr, err))
		}
		toRevoke = append(toRevoke, grantToRevoke{granterAddr, granteeAddr, auth.MsgTypeURL()})
		return false
	})

	// Deleting while iterating would mutate the store mid-iteration, so
	// collect first and delete afterwards.
	for _, g := range toRevoke {
		if err := k.AuthzKeeper.DeleteGrant(ctx, g.grantee, g.granter, g.msgType); err != nil {
			panic(fmt.Errorf("revoke leaked authz grants: failed to delete grant granter=%s grantee=%s msgType=%s: %w",
				g.granter, g.grantee, g.msgType, err))
		}
	}
}
