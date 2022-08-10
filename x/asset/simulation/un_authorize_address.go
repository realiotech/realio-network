package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/realiotech/realio-network/x/v1/asset/keeper"
	"github.com/realiotech/realio-network/x/v1/asset/types"
)

func SimulateMsgUnAuthorizeAddress(
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgUnAuthorizeAddress{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the UnAuthorizeAddress simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "UnAuthorizeAddress simulation not implemented"), nil, nil
	}
}
