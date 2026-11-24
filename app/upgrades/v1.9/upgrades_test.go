package v9_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/prefix"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/erc20"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"

	"github.com/realiotech/realio-network/app"
	v9 "github.com/realiotech/realio-network/app/upgrades/v1.9"
)

// The mainnet and testnet exports taken after v1.8.0, which is the state the
// v1.9.0 upgrade runs on.
var genesisFixtures = map[string]string{
	"mainnet": "../v1.8/testdata/exported_mainnet_after.json",
	"testnet": "../v1.8/testdata/testnet_2809.json",
}

// TestV19UpgradeKeepsERC20State runs the v1.9.0 upgrade on real chain state
// and checks that x/erc20 keeps its state, so permissionless registration
// stays disabled.
func TestV19UpgradeKeepsERC20State(t *testing.T) {
	for name, path := range genesisFixtures {
		t.Run(name, func(t *testing.T) {
			realioApp, chainID, initialHeight, proposerAddr, blockTime := app.SetupWithGenesisFile(t, path)
			ctx := app.NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime).WithChainID(chainID)

			// Before v1.9.0, erc20, wasm and 08-wasm are not in the version map.
			versionMap := prefix.NewStore(ctx.KVStore(realioApp.GetKey(upgradetypes.StoreKey)), []byte{upgradetypes.VersionMapByte})
			for _, module := range []string{erc20types.ModuleName, wasmtypes.ModuleName, ibcwasmtypes.ModuleName} {
				versionMap.Delete([]byte(module))
			}

			before := erc20.ExportGenesis(ctx, realioApp.Erc20Keeper)
			require.False(t, before.Params.PermissionlessRegistration, "test premise: permissionless registration is off")
			require.NotEmpty(t, before.TokenPairs, "test premise: the chain has token pairs")

			plan := upgradetypes.Plan{Name: v9.UpgradeName, Height: ctx.BlockHeight()}
			require.NoError(t, realioApp.UpgradeKeeper.ScheduleUpgrade(ctx, plan))
			ctx = ctx.WithBlockTime(time.Now())
			require.NoError(t, realioApp.UpgradeKeeper.ApplyUpgrade(ctx, plan))

			// Params, token pairs, allowances and precompiles are unchanged.
			require.Equal(t, before, erc20.ExportGenesis(ctx, realioApp.Erc20Keeper))
			require.False(t, realioApp.Erc20Keeper.GetParams(ctx).PermissionlessRegistration)

			vm, err := realioApp.UpgradeKeeper.GetModuleVersionMap(ctx)
			require.NoError(t, err)
			require.Equal(t, erc20.AppModuleBasic{}.ConsensusVersion(), vm[erc20types.ModuleName])

			// Only gov can register an ERC20.
			_, err = realioApp.Erc20Keeper.RegisterERC20(ctx, &erc20types.MsgRegisterERC20{
				Signer:         sdk.AccAddress(make([]byte, 20)).String(),
				Erc20Addresses: []string{before.TokenPairs[0].Erc20Address},
			})
			require.ErrorContains(t, err, "invalid authority")
		})
	}
}
