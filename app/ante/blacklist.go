package ante

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// BlacklistKeeper is the subset of x/blacklist's keeper both decorators
// below need: a single membership check against on-chain state (a KVStore
// set), not a compiled-in list. Addresses are added via genesis, a
// chain-upgrade handler, or MsgUpdateBlacklist — the latter is gated to the
// module's authority (x/gov by default), so a compromised key can never
// remove itself from the blacklist. Checked by raw address bytes, not the
// bech32 string, to avoid an encode/decode round trip on every tx.
type BlacklistKeeper interface {
	IsBlacklisted(ctx context.Context, addr sdk.AccAddress) bool
}

func blacklistedErr(addr sdk.AccAddress) error {
	return errorsmod.Wrapf(errortypes.ErrUnauthorized, "address %s is blacklisted", addr.String())
}

// EVMBlacklistDecorator rejects any raw Ethereum-style transaction
// (MsgEthereumTx) whose sender is a blacklisted address. It belongs on the
// EVM ante chain (newEthAnteHandler) — a Cosmos-tx chain never carries a
// MsgEthereumTx (evmosantecosmos.RejectMessagesDecorator already rejects
// those), so this decorator only makes sense there.
type EVMBlacklistDecorator struct {
	keeper BlacklistKeeper
}

func NewEVMBlacklistDecorator(keeper BlacklistKeeper) EVMBlacklistDecorator {
	return EVMBlacklistDecorator{keeper: keeper}
}

var _ sdk.AnteDecorator = EVMBlacklistDecorator{}

func (bd EVMBlacklistDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if bd.keeper == nil {
		return next(ctx, tx, simulate)
	}

	for _, msg := range tx.GetMsgs() {
		ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			continue
		}
		from := ethMsg.GetFrom()
		if bd.keeper.IsBlacklisted(ctx, from) {
			return ctx, blacklistedErr(from)
		}
	}

	return next(ctx, tx, simulate)
}

// CosmosBlacklistDecorator rejects any plain Cosmos SDK transaction
// (MsgSend, MsgDelegate, etc.) signed by a blacklisted address. It belongs
// on the Cosmos ante chain (newCosmosAnteHandler): a MsgEthereumTx doesn't
// implement authsigning.SigVerifiableTx (its GetSigners signature doesn't
// match), so this decorator never sees EVM txs — EVMBlacklistDecorator
// handles those separately.
type CosmosBlacklistDecorator struct {
	keeper BlacklistKeeper
	cdc    codec.Codec
}

func NewCosmosBlacklistDecorator(keeper BlacklistKeeper, cdc codec.Codec) CosmosBlacklistDecorator {
	return CosmosBlacklistDecorator{keeper: keeper, cdc: cdc}
}

var _ sdk.AnteDecorator = CosmosBlacklistDecorator{}

func (bd CosmosBlacklistDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if bd.keeper == nil {
		return next(ctx, tx, simulate)
	}

	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return next(ctx, tx, simulate)
	}

	signers, err := sigTx.GetSigners()
	if err != nil {
		return next(ctx, tx, simulate)
	}

	for _, s := range signers {
		addr := sdk.AccAddress(s)
		if bd.keeper.IsBlacklisted(ctx, addr) {
			return ctx, blacklistedErr(addr)
		}
	}

	if err := bd.checkAuthzSigners(ctx, tx.GetMsgs(), 1); err != nil {
		return ctx, err
	}

	// AuthInfo.Fee.Granter is a top-level tx field, not a signer: a
	// blacklisted address that granted a fee allowance before being
	// blacklisted never needs to sign again for the grantee to keep
	// drawing fees from it (DeductFeeDecorator reads this same field and
	// debits the granter — x/auth/ante/fee.go). BlacklistSendRestriction
	// (x/blacklist/keeper/restrictions.go) deliberately doesn't cover this
	// case: its destination is the fee-collector module account, the same
	// shape as every legitimate automatic protocol settlement it has to
	// leave open. Checked here instead, where the account is unambiguously
	// present as a real fee-granter, not overloaded with unrelated
	// module-account traffic.
	if feeTx, ok := tx.(sdk.FeeTx); ok {
		if granter := feeTx.FeeGranter(); granter != nil {
			addr := sdk.AccAddress(granter)
			if bd.keeper.IsBlacklisted(ctx, addr) {
				return ctx, blacklistedErr(addr)
			}
		}
	}

	return next(ctx, tx, simulate)
}

// checkAuthzSigners rejects MsgExec payloads whose inner message signer (the
// authz granter) is blacklisted. The outer transaction is signed only by the
// grantee, so checking transaction signers alone does not cover this path.
func (bd CosmosBlacklistDecorator) checkAuthzSigners(ctx sdk.Context, msgs []sdk.Msg, nestedLvl int) error {
	if nestedLvl >= maxNestedMsgs {
		return errorsmod.Wrapf(errortypes.ErrUnauthorized, "found more nested authz messages than permitted; limit is %d", maxNestedMsgs)
	}

	for _, msg := range msgs {
		exec, ok := msg.(*authz.MsgExec)
		if !ok {
			continue
		}

		innerMsgs, err := exec.GetMessages()
		if err != nil {
			return errorsmod.Wrapf(errortypes.ErrUnauthorized, "failed to unpack authz messages: %s", err)
		}

		for _, innerMsg := range innerMsgs {
			signers, _, err := bd.cdc.GetMsgV1Signers(innerMsg)
			if err != nil {
				return errorsmod.Wrapf(errortypes.ErrUnauthorized, "failed to extract authz message signers: %s", err)
			}
			for _, signer := range signers {
				addr := sdk.AccAddress(signer)
				if bd.keeper.IsBlacklisted(ctx, addr) {
					return blacklistedErr(addr)
				}
			}
		}

		if err := bd.checkAuthzSigners(ctx, innerMsgs, nestedLvl+1); err != nil {
			return err
		}
	}

	return nil
}
