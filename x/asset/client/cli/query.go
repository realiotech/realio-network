package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/realiotech/realio-network/x/asset/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group asset queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryTokens())
	cmd.AddCommand(CmdQueryToken())

	return cmd
}

// NewMultiSendTxCmd returns a CLI command handler for creating a MsgMultiSend transaction.
// For a better UX this command is limited to send funds from one account to two or more accounts.
func NewQueryPrivilegeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-privilege <privilege_name> <path_to_json_file>",
		Short: "Query privilege state",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a privilege states.
Query request should be defined in a JSON file.

Example:
$ %s query asset query-privilege mint path/to/request.json

Where request.json contains:

{
  "message": 
    {
      "address": "rio1...",
    }
  ,
  "message_type": "cosmos.bank.v1beta1.QueryBalances",
}
`,
			)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			message, err := parseMsgContent(args[1])
			if err != nil {
				return err
			}

			req := &types.QueryPrivilegeRequest{
				PrivilegeName: args[0],
				Request:       message,
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.QueryPrivilege(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
