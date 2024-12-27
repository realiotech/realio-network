package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/bridge/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdBridgeIn())
	cmd.AddCommand(CmdBridgeOut())
	cmd.AddCommand(CmdRegisterNewCoins())
	cmd.AddCommand(CmdDeregisterCoins())
	cmd.AddCommand(CmdUpdateEpochDuration())
	// this line is used by starport scaffolding # 1

	return cmd
}

func CmdBridgeIn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge-in [amount]",
		Short: "Broadcast message BridgeIn",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			if err = coin.Validate(); err != nil {
				return err
			}

			msg := &types.MsgBridgeIn{
				Authority: clientCtx.GetFromAddress().String(),
				Coin:      coin,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdBridgeOut() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge-out [amount]",
		Short: "Broadcast message BridgeIn",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			if err = coin.Validate(); err != nil {
				return err
			}

			msg := &types.MsgBridgeOut{
				Signer: clientCtx.GetFromAddress().String(),
				Coin:   coin,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDeregisterCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deregister-coins [denom1 denom2 ...]",
		Short: "Broadcast message DeregisterCoins",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var denoms []string
			denoms = append(denoms, args[0:len(args)-1]...)

			msg := &types.MsgDeregisterCoins{
				Authority: clientCtx.GetFromAddress().String(),
				Denoms:    denoms,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdRegisterNewCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-coins [amount]",
		Short: "Broadcast message RegisterNewCoins",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgRegisterNewCoins{
				Authority: clientCtx.GetFromAddress().String(),
				Coins:     coins,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdUpdateEpochDuration() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-epoch-duration [duration]",
		Short: "Broadcast message UpdateEpochDuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			duration, err := time.ParseDuration(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateEpochDuration{
				Authority: clientCtx.GetFromAddress().String(),
				Duration:  duration,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
