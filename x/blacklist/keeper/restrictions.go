package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

// BlacklistSendRestriction blocks a blacklisted address from sending coins
// to another USER account. Registered as a bank send restriction (app.go),
// so it runs inside BankKeeper.SendCoins / SendCoinsFromAccountToModule
// regardless of what triggered the transfer — a top-level MsgSend, an authz
// MsgExec's inner message, an ERC-20 precompile's transferFrom, IBC, or
// anything else that eventually calls into the bank keeper. That's the one
// place all of those converge; the ante-level blacklist decorators only
// ever see the OUTER transaction's signers, which several of those paths
// deliberately aren't.
//
// Two deliberate carve-outs, both learned from a real EndBlocker panic
// during testing (multistaking's own maturity payout burns a matured
// delegator's locked "representation" coin FROM the delegator's account,
// as a bookkeeping step, before paying the real coin back to them — this
// is bookkeeping for an unbonding that predates the blacklisting, not a
// blacklisted account "acting", and it isn't wrapped to handle SendCoins
// failing):
//
//  1. toAddr is not checked at all: blocking incoming transfers would also
//     block the chain's own automatic payouts TO a blacklisted address (a
//     matured unbonding, a reward withdrawal someone else triggers on the
//     delegator's behalf, etc.).
//  2. fromAddr is only checked when toAddr is a real USER account, not a
//     module account: every automatic protocol settlement that debits a
//     blacklisted address's own balance (unbonding/reward bookkeeping,
//     burns, escrow) routes through a module account on the receiving end,
//     and none of those callers are prepared for SendCoins to fail. The
//     actual threat this closes — authz/erc20 moving a blacklisted
//     account's funds to an attacker-controlled wallet — always has a
//     plain user account as the destination, so this carve-out doesn't
//     reopen it. It does leave one narrower gap open: a feegrant allowance
//     a blacklisted address granted before being blacklisted still lets
//     the grantee draw fees from it (fromAddr=blacklisted,
//     toAddr=fee-collector, a module account) — bounded by the original
//     SpendLimit and closed separately at the ante layer instead (see
//     app/ante/blacklist.go), since that fix doesn't share this
//     module-account collision.
func (k Keeper) BlacklistSendRestriction(ctx context.Context, fromAddr, toAddr sdk.AccAddress, _ sdk.Coins) (sdk.AccAddress, error) {
	if k.allowAddrs[toAddr.String()] {
		return toAddr, nil
	}
	if k.IsBlacklisted(ctx, fromAddr) {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized, "address %s is blacklisted", fromAddr)
	}
	return toAddr, nil
}
