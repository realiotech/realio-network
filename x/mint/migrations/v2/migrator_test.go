package v2_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/realiotech/realio-network/x/mint"
	"github.com/realiotech/realio-network/x/mint/exported"
	v2 "github.com/realiotech/realio-network/x/mint/migrations/v2"
	"github.com/realiotech/realio-network/x/mint/types"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(_ sdk.Context, ps exported.ParamSet) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := storetypes.NewKVStoreKey(v2.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStoreService := runtime.NewKVStoreService(storeKey)
	store := kvStoreService.OpenKVStore(ctx)

	legacySubspace := newMockSubspace(types.DefaultParams())
	assert.NoError(t, v2.Migrate(ctx, store, legacySubspace, cdc))

	var res types.Params
	bz, err := store.Get(v2.ParamsKey)
	assert.NoError(t, err)
	assert.NoError(t, cdc.Unmarshal(bz, &res))
	assert.Equal(t, legacySubspace.ps, res)
}
