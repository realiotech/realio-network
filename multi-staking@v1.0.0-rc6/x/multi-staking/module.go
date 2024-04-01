package multistaking

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/realio-tech/multi-staking-module/x/multi-staking/client/cli"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModule embeds the Cosmos SDK's x/staking AppModuleBasic.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the staking module's name.
func (AppModuleBasic) Name() string {
	return multistakingtypes.ModuleName
}

// RegisterLegacyAminoCodec register module codec
func (am AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	multistakingtypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module interface
func (am AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	multistakingtypes.RegisterInterfaces(reg)
}

// DefaultGenesis returns multi-staking module default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(multistakingtypes.DefaultGenesis())
}

// ValidateGenesis validate genesis state for multi-staking module
func (am AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState multistakingtypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", multistakingtypes.ModuleName, err)
	}

	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the staking module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := multistakingtypes.RegisterQueryHandlerClient(context.Background(), mux, multistakingtypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the staking module's root tx command.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd returns the multi-staking and staking module's root query command.
func (AppModuleBasic) GetQueryCmd() (queryCmd *cobra.Command) {
	return cli.GetQueryCmd()
}

// AppModule embeds the Cosmos SDK's x/staking AppModule where we only override
// specific methods.
type AppModule struct {
	AppModuleBasic
	// embed the Cosmos SDK's x/staking AppModule
	skAppModule staking.AppModule

	keeper multistakingkeeper.Keeper
	sk     stakingkeeper.Keeper
	ak     stakingtypes.AccountKeeper
	bk     stakingtypes.BankKeeper
}

// NewAppModule creates a new AppModule object using the native x/staking module
// AppModule constructor.
func NewAppModule(cdc codec.Codec, keeper multistakingkeeper.Keeper, sk stakingkeeper.Keeper, ak stakingtypes.AccountKeeper, bk stakingtypes.BankKeeper) AppModule {
	stakingAppMod := staking.NewAppModule(cdc, sk, ak, bk)
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		skAppModule:    stakingAppMod,
		keeper:         keeper,
		sk:             sk,
		ak:             ak,
		bk:             bk,
	}
}

// Name returns the staking module's name.
func (AppModule) Name() string {
	return multistakingtypes.ModuleName
}

// RegisterInvariants registers the staking module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	am.skAppModule.RegisterInvariants(ir)
	multistakingkeeper.RegisterInvariants(ir, am.keeper)
}

// Deprecated: Route returns the message routing key for the staking module.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the staking module's querier route name.
func (AppModule) QuerierRoute() string {
	return multistakingtypes.QuerierRoute
}

// LegacyQuerierHandler returns the staking module sdk.Querier.
// TODO: add legacy querier
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	stakingtypes.RegisterMsgServer(cfg.MsgServer(), multistakingkeeper.NewMsgServerImpl(am.keeper))
	multistakingtypes.RegisterQueryServer(cfg.QueryServer(), multistakingkeeper.NewQueryServerImpl(am.keeper))

	querier := stakingkeeper.Querier{Keeper: am.sk}
	stakingtypes.RegisterQueryServer(cfg.QueryServer(), querier)
}

// InitGenesis initial genesis state for multi-staking module
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState multistakingtypes.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	return am.keeper.InitGenesis(ctx, genesisState)
}

// ExportGenesis export multi-staking state as raw message for multi-staking module
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// BeginBlock returns the begin blocker for the multi-staking module.
func (am AppModule) BeginBlock(ctx sdk.Context, requestBeginBlock abci.RequestBeginBlock) {
	am.skAppModule.BeginBlock(ctx, requestBeginBlock)
}

// EndBlock returns the end blocker for the multi-staking module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, requestEndBlock abci.RequestEndBlock) []abci.ValidatorUpdate {
	// calculate the amount of coin
	matureUnbondingDelegations := am.keeper.GetMatureUnbondingDelegations(ctx)
	// staking endblock
	valUpdates := am.skAppModule.EndBlock(ctx, requestEndBlock)
	// update endblock multi-staking
	am.keeper.EndBlocker(ctx, matureUnbondingDelegations)

	return valUpdates
}

// ConsensusVersion return module consensus version
func (AppModule) ConsensusVersion() uint64 { return 1 }
