package keeper

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/x/asset/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		bankKeeper   types.BankKeeper
		ak           types.AccountKeeper
		allowAddrs   map[string]bool

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		Schema collections.Schema
		Params collections.Item[types.Params]
		Token  collections.Map[string, types.Token]
	}
)

// NewKeeper returns a new Keeper object with a given codec, dedicated
// store key, a BankKeeper implementation, an AccountKeeper implementation, and a parameter Subspace used to
// store and fetch module parameters. It also has an allowAddrs map[string]bool to skip restrictions for module addresses.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	bankKeeper types.BankKeeper,
	ak types.AccountKeeper,
	allowAddrs map[string]bool,
	authority string,
) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		bankKeeper:   bankKeeper,
		ak:           ak,
		allowAddrs:   allowAddrs,
		Params:       collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Token:        collections.NewMap(sb, types.TokenKeyPrefix, "token", collections.StringKey, codec.CollValue[types.Token](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return &k
}

func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) IsAddressAuthorizedToSend(ctx context.Context, symbol string, address sdk.AccAddress) (authorized bool) {
	lowerCased := strings.ToLower(symbol)
	token, err := k.Token.Get(ctx, lowerCased)
	if err != nil {
		return false
	}

	return token.AddressIsAuthorized(address)
}
