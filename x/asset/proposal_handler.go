package asset

import (
	"github.com/realiotech/realio-network/x/asset/client/cli"
	"github.com/realiotech/realio-network/x/asset/keeper"
	"github.com/realiotech/realio-network/x/asset/types"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	AddTokenManagerHandler    = govclient.NewProposalHandler(cli.NewCmdAddTokenManagerProposal)
	RemoveTokenManagerHandler = govclient.NewProposalHandler(cli.NewCmdRemoveTokenManagerProposal)
)

// NewAssetProposalHandler creates a governance handler to manage asset proposals.
func NewAssetProposalHandler(k *keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.AddTokenManager:
			return k.AddTokenManager(ctx, c)
		case *types.RemoveTokenManager:
			return k.RemoveTokenManager(ctx, c)
		default:
			return sdkerrors.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}
