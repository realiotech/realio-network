package app

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	transferEventTopic       = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	approvalEventTopic       = crypto.Keccak256Hash([]byte("Approval(address,address,uint256)"))
	approvalForAllEventTopic = crypto.Keccak256Hash([]byte("ApprovalForAll(address,address,bool)"))
)

type evmBlacklistKeeper interface {
	IsBlacklisted(ctx context.Context, addr sdk.AccAddress) bool
}

// EVMTokenBlacklistHook rejects a successful EVM transaction if a standard
// ERC-20/ERC-721 event shows that it debited, or created delegated spending
// authority for, a blacklisted owner. Because this runs after EVM execution
// and returning an error reverts the transaction, it also covers calls routed
// through contracts as well as EIP-2612 permit + transferFrom flows.
type EVMTokenBlacklistHook struct {
	keeper evmBlacklistKeeper
}

func NewEVMTokenBlacklistHook(keeper evmBlacklistKeeper) EVMTokenBlacklistHook {
	return EVMTokenBlacklistHook{keeper: keeper}
}

func (h EVMTokenBlacklistHook) PostTxProcessing(
	ctx sdk.Context,
	_ common.Address,
	_ core.Message,
	receipt *ethtypes.Receipt,
) error {
	if h.keeper == nil || receipt == nil {
		return nil
	}

	for _, log := range receipt.Logs {
		if log == nil || len(log.Topics) < 2 || !isOwnerIndexedTokenEvent(log.Topics[0]) {
			continue
		}

		owner := sdk.AccAddress(common.BytesToAddress(log.Topics[1].Bytes()).Bytes())
		if h.keeper.IsBlacklisted(ctx, owner) {
			return errorsmod.Wrapf(errortypes.ErrUnauthorized, "token owner %s is blacklisted", owner)
		}
	}

	return nil
}

func isOwnerIndexedTokenEvent(topic common.Hash) bool {
	return topic == transferEventTopic || topic == approvalEventTopic || topic == approvalForAllEventTopic
}
