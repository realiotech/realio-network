package cli

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/realiotech/realio-network/v2/x/asset/types"
	"github.com/spf13/cobra"
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

	cmd.AddCommand(CmdMsgCreateToken())
	cmd.AddCommand(CmdMsgUpdateToken())
	cmd.AddCommand(CmdCreateToken())
	cmd.AddCommand(CmdUpdateToken())
	cmd.AddCommand(CmdAuthorizeAddress())
	cmd.AddCommand(CmdUnAuthorizeAddress())
	cmd.AddCommand(CmdTransferToken())
	// this line is used by starport scaffolding # 1

	return cmd
}
