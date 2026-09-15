package claim

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/realiotech/realio-network/x/claim/client/cli"
	"github.com/realiotech/realio-network/x/claim/keeper"
	"github.com/realiotech/realio-network/x/claim/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// ConsensusVersion defines the current x/claim module consensus version.
const ConsensusVersion = 1

// AppModuleBasic implements the module.AppModuleBasic interface. The
// old->new address links and the module's admin can only be set by
// genesis, MsgLinkAddress, or a chain-upgrade handler — the Query RPCs,
// however, are open to anyone.
type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

func (AppModuleBasic) GetQueryCmd() *cobra.Command { return cli.GetQueryCmd() }

func (AppModuleBasic) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(types.DefaultGenesis())
	if err != nil {
		panic(err)
	}
	return bz
}

func (AppModuleBasic) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return gs.Validate()
}

// AppModule implements the module.AppModule interface.
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

// IsOnePerModuleType and IsAppModule mark this struct as satisfying
// cosmossdk.io/core/appmodule.AppModule (a depinject-related tagging
// interface required by module.Manager in this SDK version).
func (AppModule) IsOnePerModuleType() {}
func (AppModule) IsAppModule()        {}

// RegisterServices registers the module's Msg and Query gRPC services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// ConsensusVersion implements module.HasConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

func (am AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, bz json.RawMessage) {
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		panic(fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err))
	}
	if err := gs.Validate(); err != nil {
		panic(err)
	}

	if err := am.keeper.SetAdmin(ctx, gs.Admin); err != nil {
		panic(err)
	}
	for _, link := range gs.Links {
		oldAddr, err := sdk.AccAddressFromBech32(link.OldAddress)
		if err != nil {
			panic(fmt.Errorf("invalid old_address %q in %s genesis: %w", link.OldAddress, types.ModuleName, err))
		}
		if err := am.keeper.SetLink(ctx, oldAddr, link.NewAddress); err != nil {
			panic(err)
		}
	}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	links, err := am.keeper.GetAllLinks(ctx)
	if err != nil {
		panic(err)
	}
	bz, err := json.Marshal(types.GenesisState{Admin: am.keeper.GetAdmin(ctx), Links: links})
	if err != nil {
		panic(err)
	}
	return bz
}
