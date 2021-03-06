package keeper_test

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/realiotech/realio-network/simapp"
	"github.com/realiotech/realio-network/x/mint/types"
)

// returns context and an app with updated mint keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.MintKeeper.SetParams(ctx, types.DefaultParams())
	app.MintKeeper.SetMinter(ctx, types.DefaultInitialMinter())

	return app, ctx
}
