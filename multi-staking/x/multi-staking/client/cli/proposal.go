package cli

import (
	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func NewCmdSubmitAddMultiStakingCoinProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-multistaking-coin [title] [description] [denom] [bond_weight] [deposit]",
		Args:  cobra.ExactArgs(5),
		Short: "Submit an add multistaking coin proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bondWeight, err := sdk.NewDecFromStr(args[3])
			if err != nil {
				return err
			}
			from := clientCtx.GetFromAddress()
			content := types.NewAddMultiStakingCoinProposal(
				args[0], args[1], args[2], bondWeight,
			)

			deposit, err := sdk.ParseCoinsNormalized(args[4])
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

func NewCmdUpdateBondWeightProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-bond-weight [title] [description] [denom] [bond_weight] [deposit]",
		Args:  cobra.ExactArgs(5),
		Short: "Submit update bond weight for bond coin proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bondWeight, err := sdk.NewDecFromStr(args[3])
			if err != nil {
				return err
			}
			from := clientCtx.GetFromAddress()
			content := types.NewUpdateBondWeightProposal(
				args[0], args[1], args[2], bondWeight,
			)

			deposit, err := sdk.ParseCoinsNormalized(args[4])
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
