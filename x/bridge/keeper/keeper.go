package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/realiotech/realio-network/x/bridge/types"
)

// Keeper of the mint store
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	authKeeper   types.AccountKeeper
	bankKeeper   types.BankKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	Schema          collections.Schema
	Params          collections.Item[types.Params]
	EpochInfo       collections.Item[types.EpochInfo]
	RegisteredCoins collections.Map[string, types.RateLimit]
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec, storeService store.KVStoreService,
	ak types.AccountKeeper, bk types.BankKeeper,
	authority string,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:             cdc,
		storeService:    storeService,
		authKeeper:      ak,
		bankKeeper:      bk,
		authority:       authority,
		Params:          collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		EpochInfo:       collections.NewItem(sb, types.EpochInfoKey, "epoch_info", codec.CollValue[types.EpochInfo](cdc)),
		RegisteredCoins: collections.NewMap(sb, types.RegisteredCoinsPrefix, "registered_coins", collections.StringKey, codec.CollValue[types.RateLimit](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) UpdateInflow(ctx context.Context, coin sdk.Coin) error {
	epochInfo, err := k.EpochInfo.Get(ctx)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "failed to get bridge epoch info")
	}

	if epochInfo.EpochCountingStarted {
		ratelimit, err := k.RegisteredCoins.Get(ctx, coin.Denom)
		if err != nil {
			return errorsmod.Wrapf(types.ErrCoinNotRegister, "denom: %s", coin.Denom)
		}

		err = ratelimit.CheckAddInflowThreshold(coin.Amount)
		if err != nil {
			return err
		}

		return k.RegisteredCoins.Set(ctx, coin.Denom, ratelimit)
	}

	return nil
}
