package cli

import (
	"fmt"
	"strconv"
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
var TokenNameMaxLength = 100
var TokenSymbolMaxLength = 100
var TokenDescriptionMaxLength = 200

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
	cmd.AddCommand(NewDisablePrivilegeCmd())
	cmd.AddCommand(NewAssignPrivilegeCmd())
	cmd.AddCommand(NewUnassignPrivilegeCmd())
	cmd.AddCommand(NewAllocateTokenCmd())
	cmd.AddCommand(NewUpdateTokenCmd())
	cmd.AddCommand(NewCreateTokenCmd())

	return cmd
}

// NewCreateTokenCmd returns a CLI command handler for creating a MsgUpdateToken transaction.
func NewCreateTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-token <creator_address> <manager_address> <name> <symbol> <decimal> <description> <excluded_privilege_1> <excluded_privilege_2> ...",
		Short: "Create token given the creator, manager address, name, symbol, decimal and description",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create token.

Example: 
$ %s tx asset create-token rio1... rio2... new_name new_symbol 6 new_description

`,
			)),
		Args: cobra.MinimumNArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			creatorAddress := clientCtx.FromAddress.String()

			if len(args[2]) > TokenNameMaxLength || len(args[3]) > TokenSymbolMaxLength || len(args[5]) > TokenDescriptionMaxLength {
				return fmt.Errorf("name|symbol|description reach max length: name should be less than %v, symbol should be less than %v, description should be less than %v",
					TokenNameMaxLength, TokenSymbolMaxLength, TokenDescriptionMaxLength)
			}

			decimal, err := strconv.ParseUint(args[4], 10, 32)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateToken{
				Creator:            creatorAddress,
				Manager:            args[1],
				Name:               args[2],
				Symbol:             args[3],
				Description:        args[5],
				Decimal:            uint32(decimal),
				ExcludedPrivileges: args[6:],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUpdateTokenCmd returns a CLI command handler for creating a MsgUpdateToken transaction.
func NewUpdateTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-token <manager_address> <token_id> <name> <symbol> <description>",
		Short: "Update token metadata",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update token information.

Example: 
$ %s tx asset update-token rio1... tokenID1 update_name update_symbol update_description

`,
			)),
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			managerAddress := clientCtx.FromAddress.String()

			if len(args[2]) > TokenNameMaxLength || len(args[3]) > TokenSymbolMaxLength || len(args[4]) > TokenDescriptionMaxLength {
				return fmt.Errorf("name|symbol|description reach max length: name should be less than %v, symbol should be less than %v, description should be less than %v",
					TokenNameMaxLength, TokenSymbolMaxLength, TokenDescriptionMaxLength)
			}

			msg := &types.MsgUpdateToken{
				Manager:     managerAddress,
				TokenId:     args[1],
				Name:        args[2],
				Symbol:      args[3],
				Description: args[4],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewAllocateTokenCmd returns a CLI command handler for creating a MsgAllocateToken transaction.
func NewAllocateTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allocate-token <manager_address> <token_id> <path_to_balances_file>",
		Short: "Allocate token to a list of addresses given the amount for each address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Allocate token privilege for a list of addresses.
List of addresses are defined in a json file

Example: 
$ %s tx asset allocate-token rio1... tokenID1 path/to/balances.json

where balances.json contains:

{
	"balances": [
		{
			"address": "rio2...."
			"amount": 100
		}, {
			"address": "rio3...."
			"amount": 100
		}
	]
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

			managerAddress := clientCtx.FromAddress.String()

			balances, err := parseBalances(clientCtx.Codec, args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgAllocateToken{
				Manager:  managerAddress,
				TokenId:  args[1],
				Balances: balances,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewAssignPrivilegeCmd returns a CLI command handler for creating a MsgAssignPrivilege transaction.
func NewAssignPrivilegeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign-privilege <manager_address> <token_id> <privilege_name> <address_1> <address_2>..",
		Short: "Assign privilege a token's privilege for a list of addresses",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Assign a token privilege for a list of addresses.

Example: 
$ %s tx asset assign-privilege rio1... tokenID1 mint rio2... rio3... rio4... rio5...
`,
			)),
		Args: cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			managerAddress := clientCtx.FromAddress.String()

			msg := &types.MsgAssignPrivilege{
				Manager:    managerAddress,
				TokenId:    args[1],
				AssignedTo: args[3:],
				Privilege:  args[2],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUnassignPrivilegeCmd returns a CLI command handler for creating a MsgUnassignPrivilege transaction.
func NewUnassignPrivilegeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unassign-privilege <manager_address> <token_id> <privilege_name> <address_1> <address_2>..",
		Short: "Unassign privilege from a token's privilege from a list of addresses",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unassign from a token privilege for a list of addresses.

Example: 
$ %s tx asset unassign-privilege rio1... tokenID1 mint rio2... rio3... rio4... rio5...
`,
			)),
		Args: cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			managerAddress := clientCtx.FromAddress.String()

			msg := &types.MsgUnassignPrivilege{
				Manager:        managerAddress,
				TokenId:        args[1],
				UnassignedFrom: args[3:],
				Privilege:      args[2],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewDisablePrivilegeCmd returns a CLI command handler for creating a MsgDisablePrivilege transaction.
func NewDisablePrivilegeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable-privilege <manager_address> <token_id> <privilege_name>",
		Short: "Disable a privilege from a token given manager address and privilege name",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Disable a privilege for a token.

Example:
$ %s tx asset disable-privilege rio1... tokenID1 mint
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

			managerAddress := clientCtx.FromAddress.String()

			msg := &types.MsgDisablePrivilege{
				Manager:           managerAddress,
				TokenId:           args[1],
				DisabledPrivilege: args[2],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewExecutePrivilegeMsgCmd returns a CLI command handler for creating a MsgExecutePrivilegeMsg transaction.
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
