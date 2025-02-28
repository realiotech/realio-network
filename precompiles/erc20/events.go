// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/os/x/evm/core/vm"

	auth "github.com/evmos/os/precompiles/authorization"
	cmn "github.com/evmos/os/precompiles/common"
)

const (
	// EventTypeTransfer defines the event type for the ERC-20 Transfer and TransferFrom transactions.
	EventTypeTransfer = "Transfer"
	EventTypeMint     = "Mint"
	EventTypeBurn     = "Burn"
)

// EmitTransferEvent creates a new Transfer event emitted on transfer and transferFrom transactions.
func (p Precompile) EmitTransferEvent(ctx sdk.Context, stateDB vm.StateDB, from, to common.Address, value *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeTransfer]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(from)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(to)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{event.Inputs[2]}
	packed, err := arguments.Pack(value)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // block height won't exceed uint64
	})

	return nil
}

// EmitApprovalEvent creates a new approval event emitted on Approve, IncreaseAllowance
// and DecreaseAllowance transactions.
func (p Precompile) EmitApprovalEvent(ctx sdk.Context, stateDB vm.StateDB, owner, spender common.Address, value *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[auth.EventTypeApproval]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(owner)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(spender)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{event.Inputs[2]}
	packed, err := arguments.Pack(value)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // block height won't exceed uint64
	})

	return nil
}

func (p Precompile) EmitMintEvent(ctx sdk.Context, stateDB vm.StateDB, to common.Address, value *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeMint]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(to)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{event.Inputs[1]}
	packed, err := arguments.Pack(value)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // block height won't exceed uint64
	})

	return nil
}

func (p Precompile) EmitBurnEvent(ctx sdk.Context, stateDB vm.StateDB, from common.Address, value *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeBurn]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(from)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{event.Inputs[0]}
	packed, err := arguments.Pack(value)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // block height won't exceed uint64
	})

	return nil
}
