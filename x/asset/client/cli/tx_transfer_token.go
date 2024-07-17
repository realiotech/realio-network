package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/realiotech/realio-network/v2/x/asset/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdTransferToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-token [symbol] [from] [to] [amount]",
		Short: "Broadcast message TransferToken",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argSymbol := args[0]
			argFrom := args[1]
			argTo := args[2]
			argAmount := args[3]
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgTransferToken(
				argSymbol,
				argFrom,
				argTo,
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
