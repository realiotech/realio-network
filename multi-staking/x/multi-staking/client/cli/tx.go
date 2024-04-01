package cli

import (
	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

// NewTxCmd returns a root CLI command handler for all x/exp transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "multi-staking transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		cli.NewCreateValidatorCmd(),
		cli.NewEditValidatorCmd(),
		cli.NewDelegateCmd(),
		cli.NewRedelegateCmd(),
		cli.NewUnbondCmd(),
		cli.NewCancelUnbondingDelegation(),
	)

	return txCmd
}
