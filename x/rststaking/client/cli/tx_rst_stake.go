package cli

import (
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/realiotech/realio-network/x/rststaking/types"
)

func CmdCreateRstStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-rst-stake [id] [address] [rst-amount] [rio-amount] [incoming-rst-txn-hash] [funded-rio-txn-hash] [rst-origin-chain] [rst-origin-address] [created] [status]",
		Short: "Create a new rstStake",
		Args:  cobra.ExactArgs(10),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// Get id
			id := args[0]

			// Get value arguments
			argAddress := args[1]
			argRstAmount, err := cast.ToInt64E(args[2])
			if err != nil {
				return err
			}
			argRioAmount, err := cast.ToInt64E(args[3])
			if err != nil {
				return err
			}
			argIncomingRstTxnHash := args[4]
			argFundedRioTxnHash := args[5]
			argRstOriginChain := args[6]
			argRstOriginAddress := args[7]
			argCreated, err := cast.ToInt64E(args[8])
			if err != nil {
				return err
			}
			argStatus := args[9]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateRstStake(
				clientCtx.GetFromAddress().String(),
				id,
				argAddress,
				argRstAmount,
				argRioAmount,
				argIncomingRstTxnHash,
				argFundedRioTxnHash,
				argRstOriginChain,
				argRstOriginAddress,
				argCreated,
				argStatus,
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

func CmdUpdateRstStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-rst-stake [id] [address] [rst-amount] [rio-amount] [incoming-rst-txn-hash] [funded-rio-txn-hash] [rst-origin-chain] [rst-origin-address] [created] [status]",
		Short: "Update a rstStake",
		Args:  cobra.ExactArgs(10),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// Get id
			id := args[0]

			// Get value arguments
			argAddress := args[1]
			argRstAmount, err := cast.ToInt64E(args[2])
			if err != nil {
				return err
			}
			argRioAmount, err := cast.ToInt64E(args[3])
			if err != nil {
				return err
			}
			argIncomingRstTxnHash := args[4]
			argFundedRioTxnHash := args[5]
			argRstOriginChain := args[6]
			argRstOriginAddress := args[7]
			argCreated, err := cast.ToInt64E(args[8])
			if err != nil {
				return err
			}
			argStatus := args[9]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateRstStake(
				clientCtx.GetFromAddress().String(),
				id,
				argAddress,
				argRstAmount,
				argRioAmount,
				argIncomingRstTxnHash,
				argFundedRioTxnHash,
				argRstOriginChain,
				argRstOriginAddress,
				argCreated,
				argStatus,
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

func CmdDeleteRstStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-rst-stake [id]",
		Short: "Delete a rstStake",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			id := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteRstStake(
				clientCtx.GetFromAddress().String(),
				id,
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
