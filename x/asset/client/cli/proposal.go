package cli

import (
	"github.com/realiotech/realio-network/x/asset/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func NewCmdAddTokenManagerProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-token-manager [title] [description] [manager] [deposit]",
		Args:  cobra.ExactArgs(4),
		Short: "Submit an add token manager proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := types.NewAddTokenManager(args[0], args[1], args[2])

			deposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

func NewCmdRemoveTokenManagerProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-token-manager [title] [description] [manager] [deposit]",
		Args:  cobra.ExactArgs(4),
		Short: "Submit an remove token manager proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := types.NewRemoveTokenManager(args[0], args[1], args[2])

			deposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
