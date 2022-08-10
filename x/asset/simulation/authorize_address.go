package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/realiotech/realio-network/v1/x/asset/keeper"
	"github.com/realiotech/realio-network/v1/x/asset/types"
)

func SimulateMsgAuthorizeAddress(
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgAuthorizeAddress{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the AuthorizeAddress simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "AuthorizeAddress simulation not implemented"), nil, nil
	}
}
