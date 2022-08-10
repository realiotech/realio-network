package asset

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/realiotech/realio-network/testutil/sample"
	assetsimulation "github.com/realiotech/realio-network/x/v1/asset/simulation"
	"github.com/realiotech/realio-network/x/v1/asset/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = assetsimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgMsgCreateToken = "op_weight_msg_msg_create_token"
	// TODO: Determine the simulation weight value
	defaultWeightMsgMsgCreateToken int = 100

	opWeightMsgMsgUpdateToken = "op_weight_msg_msg_update_token"
	// TODO: Determine the simulation weight value
	defaultWeightMsgMsgUpdateToken int = 100

	opWeightMsgCreateToken = "op_weight_msg_create_token"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateToken int = 100

	opWeightMsgUpdateToken = "op_weight_msg_update_token"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateToken int = 100

	opWeightMsgAuthorizeAddress = "op_weight_msg_authorize_address"
	// TODO: Determine the simulation weight value
	defaultWeightMsgAuthorizeAddress int = 100

	opWeightMsgUnAuthorizeAddress = "op_weight_msg_un_authorize_address"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUnAuthorizeAddress int = 100

	opWeightMsgTransferToken = "op_weight_msg_transfer_token"
	// TODO: Determine the simulation weight value
	defaultWeightMsgTransferToken int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	assetGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&assetGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgMsgCreateToken int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgMsgCreateToken, &weightMsgMsgCreateToken, nil,
		func(_ *rand.Rand) {
			weightMsgMsgCreateToken = defaultWeightMsgMsgCreateToken
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgMsgCreateToken,
		assetsimulation.SimulateMsgCreateToken(am.bankKeeper, am.keeper),
	))

	var weightMsgMsgUpdateToken int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgMsgUpdateToken, &weightMsgMsgUpdateToken, nil,
		func(_ *rand.Rand) {
			weightMsgMsgUpdateToken = defaultWeightMsgMsgUpdateToken
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgMsgUpdateToken,
		assetsimulation.SimulateMsgUpdateToken(am.bankKeeper, am.keeper),
	))

	var weightMsgCreateToken int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateToken, &weightMsgCreateToken, nil,
		func(_ *rand.Rand) {
			weightMsgCreateToken = defaultWeightMsgCreateToken
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateToken,
		assetsimulation.SimulateCreateToken(am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateToken int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateToken, &weightMsgUpdateToken, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateToken = defaultWeightMsgUpdateToken
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateToken,
		assetsimulation.SimulateMsgUpdateToken(am.bankKeeper, am.keeper),
	))

	var weightMsgAuthorizeAddress int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgAuthorizeAddress, &weightMsgAuthorizeAddress, nil,
		func(_ *rand.Rand) {
			weightMsgAuthorizeAddress = defaultWeightMsgAuthorizeAddress
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgAuthorizeAddress,
		assetsimulation.SimulateMsgAuthorizeAddress(am.bankKeeper, am.keeper),
	))

	var weightMsgUnAuthorizeAddress int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUnAuthorizeAddress, &weightMsgUnAuthorizeAddress, nil,
		func(_ *rand.Rand) {
			weightMsgUnAuthorizeAddress = defaultWeightMsgUnAuthorizeAddress
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUnAuthorizeAddress,
		assetsimulation.SimulateMsgUnAuthorizeAddress(am.bankKeeper, am.keeper),
	))

	var weightMsgTransferToken int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgTransferToken, &weightMsgTransferToken, nil,
		func(_ *rand.Rand) {
			weightMsgTransferToken = defaultWeightMsgTransferToken
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgTransferToken,
		assetsimulation.SimulateMsgTransferToken(am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
