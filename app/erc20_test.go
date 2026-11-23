package app

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	erc20types "github.com/cosmos/evm/x/erc20/types"
)

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
