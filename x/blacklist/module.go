package blacklist

import (
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

	"github.com/realiotech/realio-network/x/blacklist/keeper"
	"github.com/realiotech/realio-network/x/blacklist/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// ConsensusVersion defines the current x/blacklist module consensus version.
const ConsensusVersion = 1

// AppModuleBasic implements the module.AppModuleBasic interface. This module
// has no Msg/Query service and no CLI commands: it is state (a set of
// blocked addresses) checked directly by the ante handler, never mutated by
// a user-submitted transaction.
type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

func (AppModuleBasic) RegisterInterfaces(cdctypes.InterfaceRegistry) {}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }

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

// RegisterServices is a no-op: this module registers no Msg/Query service.
func (am AppModule) RegisterServices(module.Configurator) {}

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
	for _, addr := range gs.Addresses {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			panic(fmt.Errorf("invalid address %q in %s genesis: %w", addr, types.ModuleName, err))
		}
		if err := am.keeper.SetBlacklisted(ctx, accAddr); err != nil {
			panic(err)
		}
	}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	addrs, err := am.keeper.GetAllBlacklisted(ctx)
	if err != nil {
		panic(err)
	}
	bz, err := json.Marshal(types.GenesisState{Addresses: addrs})
	if err != nil {
		panic(err)
	}
	return bz
}
