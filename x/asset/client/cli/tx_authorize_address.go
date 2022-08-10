package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/realiotech/realio-network/v1/x/asset/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdAuthorizeAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorize-address [symbol] [address]",
		Short: "Broadcast message AuthorizeAddress",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argSymbol := args[0]
			argAddress := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAuthorizeAddress(
				clientCtx.GetFromAddress().String(),
				argSymbol,
				argAddress,
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
