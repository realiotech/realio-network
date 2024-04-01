package cli

import (
	"fmt"
	"strings"

	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryBondWeight(),
		GetCmdQueryMultiStakingCoinInfos(),
		GetCmdQueryMultiStakingLock(),
		GetCmdQueryMultiStakingLocks(),
		GetCmdQueryMultiStakingUnlock(),
		GetCmdQueryMultiStakingUnlocks(),
		GetCmdQueryValidatorMultiStakingCoin(),
		GetCmdQueryValidator(),
		GetCmdQueryValidators(),
	)

	return cmd
}

// GetCmdQueryBondWeight implements the command to query bond weight of specific denom
func GetCmdQueryBondWeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond-weight [denom]",
		Short: "Query Multi-staking coin bond weight",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryBondWeightRequest{
				Denom: args[0],
			}

			res, err := queryClient.BondWeight(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryMultiStakingCoinInfos implements the command to query all multistaking coin information
func GetCmdQueryMultiStakingCoinInfos() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coin-infos",
		Short: "Query all multistaking coin information",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryMultiStakingCoinInfosRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.MultiStakingCoinInfos(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "coin-infos")

	return cmd
}

func GetCmdQueryMultiStakingLock() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multistaking-lock [delegator] [validator]",
		Short: "Query Multi-staking lock of specific DV pair",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			_, err = sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			req := &types.QueryMultiStakingLockRequest{
				MultiStakerAddress: args[0],
				ValidatorAddress:   args[1],
			}

			res, err := queryClient.MultiStakingLock(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetCmdQueryMultiStakingLocks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multistaking-locks",
		Short: "Query all Multi-staking lock",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryMultiStakingLocksRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.MultiStakingLocks(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "multistaking-locks")

	return cmd
}

func GetCmdQueryMultiStakingUnlock() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multistaking-unlock [delegator] [validator]",
		Short: "Query Multi-staking unlock of specific DV pair",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			_, err = sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			req := &types.QueryMultiStakingUnlockRequest{
				MultiStakerAddress: args[0],
				ValidatorAddress:   args[1],
			}

			res, err := queryClient.MultiStakingUnlock(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetCmdQueryMultiStakingUnlocks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multistaking-unlocks",
		Short: "Query all Multi-staking unlock",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryMultiStakingUnlocksRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.MultiStakingUnlocks(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "multistaking-unlocks")

	return cmd
}

func GetCmdQueryValidatorMultiStakingCoin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-multistaking-coin [validator]",
		Short: "Query multistaking-coin for specific validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			_, err = sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			req := &types.QueryValidatorMultiStakingCoinRequest{
				ValidatorAddr: args[0],
			}

			res, err := queryClient.ValidatorMultiStakingCoin(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about an individual validator.

Example:
$ %s query multistaking validator %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryValidatorRequest{ValidatorAddr: addr.String()}
			res, err := queryClient.Validator(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Validator)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryValidators implements the query all validators command.
func GetCmdQueryValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about all validators on a network.

Example:
$ %s query multistaking validators
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			result, err := queryClient.Validators(cmd.Context(), &types.QueryValidatorsRequest{
				// Leaving status empty on purpose to query all validators.
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "validators")

	return cmd
}
