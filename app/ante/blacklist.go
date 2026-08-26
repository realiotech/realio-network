package ante

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// BlacklistKeeper is the subset of x/blacklist's keeper both decorators
// below need: a single membership check against on-chain state (a KVStore
// set), not a compiled-in list. Addresses are added via genesis or a
// chain-upgrade handler — the module has no Msg service, so a compromised
// key can never remove itself from the blacklist. Checked by raw address
// bytes, not the bech32 string, to avoid an encode/decode round trip on
// every tx.
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
}

func NewCosmosBlacklistDecorator(keeper BlacklistKeeper) CosmosBlacklistDecorator {
	return CosmosBlacklistDecorator{keeper: keeper}
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

	return next(ctx, tx, simulate)
}
