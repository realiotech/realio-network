package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/realiotech/realio-network/x/asset/types"
)

var DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(NewExecutePrivilegeMsgCmd())

	return cmd
}

// NewMultiSendTxCmd returns a CLI command handler for creating a MsgMultiSend transaction.
// For a better UX this command is limited to send funds from one account to two or more accounts.
func NewExecutePrivilegeMsgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-msg <user_address> <token_id> <path_to_json_file>",
		Short: "Execute privilege message for the given user address, token id and message",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Execute a privilege message.
Execute message should be defined in a JSON file.

Example:
$ %s tx asset execute-msg rio1... tokenID1 path/to/message.json

Where message.json contains:

{
  "message": 
    {
      "from_address": "rio1...",
      "to_address": "rio2...",
      "amount":[{"denom": "stake","amount": "10"}]
    }
  ,
  "message_type": "cosmos.bank.v1beta1.MsgSend",
}
`,
			)),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			message, err := parseMsgContent(args[2])
			if err != nil {
				return err
			}

			userAddress := clientCtx.FromAddress.String()

			msg := &types.MsgExecutePrivilege{
				Address:      userAddress,
				TokenId:      args[1],
				PrivilegeMsg: message,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
