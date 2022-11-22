package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/realiotech/realio-network/x/asset/types"
)

var _ = strconv.Itoa(0)

func CmdMsgCreateToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "msg-create-token [name] [symbol] [total] [decimals] [authorization-required]",
		Short: "Broadcast message MsgCreateToken",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argName := args[0]
			argSymbol := args[1]
			argTotal := args[2]
			argDecimals := args[3]
			argAuthorizationRequired, err := cast.ToBoolE(args[4])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateToken(
				clientCtx.GetFromAddress().String(),
				argName,
				argSymbol,
				argTotal,
				argDecimals,
				argAuthorizationRequired,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
