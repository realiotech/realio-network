package cli

import (
	"strconv"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/realiotech/network/x/asset/types"
)

var _ = strconv.Itoa(0)

func CmdSendToAlgorand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-to-algorand [index] [denom] [algorand-receiver] [amount]",
		Short: "Broadcast message sendToAlgorand",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argIndex := args[0]
			argDenom := args[1]
			argAlgorandReceiver := args[2]
			argAmount, err := cast.ToInt64E(args[3])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSendToAlgorand(
				clientCtx.GetFromAddress().String(),
				argIndex,
				argDenom,
				argAlgorandReceiver,
				argAmount,
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
