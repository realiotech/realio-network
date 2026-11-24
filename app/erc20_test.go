package app

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/cosmos/evm/x/erc20"
	erc20types "github.com/cosmos/evm/x/erc20/types"
)

// TestERC20GenesisRoundTrip checks that x/erc20's InitGenesis, which the
// module manager now runs instead of the InitChainer copy, restores real
// chain state exactly. The testnet export also has allowances and dynamic
// precompiles; the InitChainer copy dropped the precompiles.
func TestERC20GenesisRoundTrip(t *testing.T) {
	for name, path := range map[string]string{
		"mainnet": "upgrades/v1.8/testdata/exported_mainnet_after.json",
		"testnet": "upgrades/v1.8/testdata/testnet_2809.json",
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			var doc struct {
				AppState map[string]json.RawMessage `json:"app_state"`
			}
			require.NoError(t, json.Unmarshal(raw, &doc))

			realioApp, _, initialHeight, proposerAddr, blockTime := SetupWithGenesisFile(t, path)
			ctx := NewHeaderCtx(realioApp, initialHeight, proposerAddr, blockTime)

			var want erc20types.GenesisState
			realioApp.AppCodec().MustUnmarshalJSON(doc.AppState[erc20types.ModuleName], &want)
			got := erc20.ExportGenesis(ctx, realioApp.Erc20Keeper)

			require.Equal(t, want.Params, got.Params)
			require.ElementsMatch(t, want.TokenPairs, got.TokenPairs)
			require.ElementsMatch(t, want.NativePrecompiles, got.NativePrecompiles)
			require.ElementsMatch(t, want.DynamicPrecompiles, got.DynamicPrecompiles)
			require.Equal(t, allowanceKeys(want.Allowances), allowanceKeys(got.Allowances))
		})
	}
}

// TestExportIncludesERC20 checks that the module manager exports x/erc20.
func TestExportIncludesERC20(t *testing.T) {
	app := Setup(false, nil, 1)
	_, err := app.Commit()
	require.NoError(t, err)

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	var appState map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(exported.AppState, &appState))
	require.Contains(t, appState, erc20types.ModuleName)

	var got erc20types.GenesisState
	app.AppCodec().MustUnmarshalJSON(appState[erc20types.ModuleName], &got)
	ctx := app.NewContextLegacy(true, tmproto.Header{Height: app.LastBlockHeight()})
	want := erc20.ExportGenesis(ctx, app.Erc20Keeper)
	require.Equal(t, want.Params, got.Params)
	require.ElementsMatch(t, want.TokenPairs, got.TokenPairs)
	require.ElementsMatch(t, want.NativePrecompiles, got.NativePrecompiles)
}

// allowanceKeys returns the allowances as sorted strings, since math.Int
// values are not comparable with reflect.DeepEqual.
func allowanceKeys(allowances []erc20types.Allowance) []string {
	keys := make([]string, 0, len(allowances))
	for _, a := range allowances {
		keys = append(keys, strings.Join([]string{a.Erc20Address, a.Owner, a.Spender, a.Value.String()}, "/"))
	}
	slices.Sort(keys)
	return keys
}

func TestERC20MsgConvertERC20IsRejected(t *testing.T) {
	app := Setup(false, nil, 1)
	txConfig := app.GetTxConfig()

	encode := func(msg sdk.Msg) []byte {
		builder := txConfig.NewTxBuilder()
		require.NoError(t, builder.SetMsgs(msg))
		bz, err := txConfig.TxEncoder()(builder.GetTx())
		require.NoError(t, err)
		return bz
	}

	for _, tc := range []struct {
		msg      sdk.Msg
		accepted bool
	}{
		{&erc20types.MsgConvertCoin{}, true},
		{&erc20types.MsgRegisterERC20{}, true},
		{&erc20types.MsgToggleConversion{}, true},
		{&erc20types.MsgUpdateParams{}, true},
		{&erc20types.MsgConvertERC20{}, false},
	} {
		t.Run(sdk.MsgTypeURL(tc.msg), func(t *testing.T) {
			_, err := txConfig.TxDecoder()(encode(tc.msg))
			if tc.accepted {
				require.NoError(t, err)
				require.NotNil(t, app.MsgServiceRouter().Handler(tc.msg))
			} else {
				require.ErrorContains(t, err, "unable to resolve type URL")
				require.Nil(t, app.MsgServiceRouter().Handler(tc.msg))
			}
		})
	}

	// CheckTx rejects MsgConvertERC20, also when nested in authz, so it never
	// reaches the mempool.
	exec := authz.NewMsgExec(sdk.AccAddress(make([]byte, 20)), []sdk.Msg{&erc20types.MsgConvertERC20{}})
	for _, msg := range []sdk.Msg{&erc20types.MsgConvertERC20{}, &exec} {
		res, err := app.CheckTx(&abci.RequestCheckTx{Tx: encode(msg), Type: abci.CheckTxType_New})
		require.NoError(t, err)
		require.NotZero(t, res.Code)
		require.Contains(t, res.Log, "unable to resolve type URL /cosmos.evm.erc20.v1.MsgConvertERC20")
	}
}
