package mint

import (
	"context"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
	// "github.com/cosmos/cosmos-sdk/store/types"
)

func (mp MintPriviledge) MintToken(ctx sdk.Context, msg *MsgMintToken, tokenID string) error {

	mintedCoins := sdk.NewCoin(tokenID, sdk.NewIntFromUint64(msg.Amount))

	err := mp.bk.MintCoins(ctx, assettypes.ModuleName, sdk.NewCoins(mintedCoins))
	if err != nil {
		return err
	}

	err = mp.bk.SendCoinsFromModuleToAccount(ctx, assettypes.ModuleName, sdk.MustAccAddressFromBech32(msg.ToAccount), sdk.NewCoins(mintedCoins))

	return err
}

func (mp MintPriviledge) MsgHandler(context context.Context, msg proto.Message, tokenID string, privAcc sdk.AccAddress) (proto.Message, error) {
	ctx := sdk.UnwrapSDKContext(context)

	switch msg := msg.(type) {
	case *MsgMintToken:
		return nil, mp.MintToken(ctx, msg, tokenID)
	default:
		errMsg := fmt.Sprintf("unrecognized message type: %T for Mint priviledge", msg)
		return nil, errors.Errorf(errMsg)
	}
}
