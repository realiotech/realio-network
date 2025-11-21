package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const (
	SendMethod      = "send"
	MultiSendMethod = "multiSend"
)

func (p Precompile) Send(
	ctx sdk.Context,
	sender common.Address,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	receiver, coins, err := ParseSendArgs(args)
	if err != nil {
		return nil, err
	}

	// Create multistaking delegation message
	msgServer := bankkeeper.NewMsgServerImpl(p.bankKeeper)
	msg := &banktypes.MsgSend{
		FromAddress: sdk.AccAddress(sender.Bytes()).String(),
		ToAddress:   sdk.AccAddress(receiver.Bytes()).String(),
		Amount:      coins,
	}

	_, err = msgServer.Send(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Return success
	return method.Outputs.Pack(true)
}

func (p Precompile) MultiSend(
	ctx sdk.Context,
	sender common.Address,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse arguments
	coins, outputs, err := ParseMultiSendArgs(args)
	if err != nil {
		return nil, err
	}

	bankOutputs := make([]banktypes.Output, 0)
	for _, output := range outputs {
		amount, err := sdk.ParseCoinsNormalized(output.Amount)
		if err != nil {
			return nil, err
		}
		bankOutputs = append(bankOutputs, banktypes.Output{
			Address: sdk.AccAddress(output.Addr.Bytes()).String(),
			Coins:   amount,
		})
	}

	// Create multisend message
	msgServer := bankkeeper.NewMsgServerImpl(p.bankKeeper)
	msg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: sdk.AccAddress(sender.Bytes()).String(),
				Coins:   coins,
			},
		},
		Outputs: bankOutputs,
	}

	_, err = msgServer.MultiSend(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Return success
	return method.Outputs.Pack(true)
}
