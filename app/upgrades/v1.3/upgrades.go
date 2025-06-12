package v3

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/evm/crypto/ethsecp256k1"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1.3.0
func CreateUpgradeHandler(
	_ *module.Manager,
	_ module.Configurator,
	accountKeeper authkeeper.AccountKeeper,
	evmKeeper evmkeeper.Keeper,

) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting upgrade for v1.3.0...")

		params := evmKeeper.GetParams(sdkCtx)
		params.ExtraEIPs = []int64{3855}
		err := evmKeeper.SetParams(sdkCtx, params)
		if err != nil {
			return nil, err
		}

		// Migrate pubkey to cosmos/evm
		var accounts []sdk.AccountI
		accountKeeper.IterateAccounts(ctx, func(account sdk.AccountI) (stop bool) {
			accounts = append(accounts, account)
			return false
		})
		// Process each account
		for _, account := range accounts {
			// Skip accounts without public keys
			baseAcc, ok := account.(*authtypes.BaseAccount)
			if !ok {
				continue
			}
			oldPk := baseAcc.PubKey
			if oldPk == nil {
				continue
			}

			if oldPk.TypeUrl == "/os.crypto.v1.ethsecp256k1.PubKey" {
				oldPk.TypeUrl = "/cosmos.evm.crypto.v1.ethsecp256k1.PubKey"
				pk := ethsecp256k1.PubKey{
					Key: oldPk.Value,
				}
				account.SetPubKey(&pk)
				accountKeeper.SetAccount(ctx, account)
			}

		}

		// We have no version map changes so keep current vm
		return vm, nil
	}
}
