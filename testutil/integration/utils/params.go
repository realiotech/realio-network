package utils

import (
	"fmt"

	"github.com/cosmos/evm/testutil/integration/os/factory"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/realiotech/realio-network/testutil/integration/network"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"cosmossdk.io/math"
	commonfactory "github.com/cosmos/evm/testutil/integration/common/factory"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
)

type UpdateParamsInput struct {
	Tf      factory.TxFactory
	Network network.Network
	Pk      cryptotypes.PrivKey
	Params  interface{}
}

var authority = authtypes.NewModuleAddress("gov").String()

// UpdateEvmParams helper function to update the EVM module parameters
// It submits an update params proposal, votes for it, and waits till it passes
func UpdateEvmParams(input UpdateParamsInput) error {
	return updateModuleParams[evmtypes.Params](input, evmtypes.ModuleName)
}

// updateModuleParams helper function to update module parameters
// It submits an update params proposal, votes for it, and waits till it passes
func updateModuleParams[T interface{}](input UpdateParamsInput, moduleName string) error {
	newParams, ok := input.Params.(T)
	if !ok {
		return fmt.Errorf("invalid params type %T for module %s", input.Params, moduleName)
	}

	proposalMsg := createProposalMsg(newParams, moduleName)

	title := fmt.Sprintf("Update %s params", moduleName)
	proposalID, err := SubmitProposal(input.Tf, input.Network, input.Pk, title, proposalMsg)
	if err != nil {
		return err
	}

	return ApproveProposal(input.Tf, input.Network, input.Pk, proposalID)
}

// createProposalMsg creates the module-specific update params message
func createProposalMsg(params interface{}, name string) sdk.Msg {
	switch name {
	case evmtypes.ModuleName:
		return &evmtypes.MsgUpdateParams{Authority: authority, Params: params.(evmtypes.Params)}
	case govtypes.ModuleName:
		return &govv1types.MsgUpdateParams{Authority: authority, Params: params.(govv1types.Params)}
	case feemarkettypes.ModuleName:
		return &feemarkettypes.MsgUpdateParams{Authority: authority, Params: params.(feemarkettypes.Params)}
	case erc20types.ModuleName:
		return &erc20types.MsgUpdateParams{Authority: authority, Params: params.(erc20types.Params)}
	default:
		return nil
	}
}

func RegisterMultistakingEVMBondDenom(input UpdateParamsInput, contractAddr string, weight math.LegacyDec, proposer sdk.AccAddress) error {
	addMultiStakingProposal := multistakingtypes.NewAddMultiStakingEVMCoinProposal("tittle", "des", contractAddr, weight)

	// Submit governance proposal
	govMsg, err := govv1beta1.NewMsgSubmitProposal(
		addMultiStakingProposal,
		sdk.NewCoins(sdk.NewCoin(input.Network.GetBaseDenom(), math.NewInt(1e18).Quo(evmtypes.GetEVMCoinDecimals().ConversionFactor()))),
		proposer,
	)
	if err != nil {
		return err
	}

	txArgs := commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{govMsg},
	}
	proposalId, err := submitProposal(input.Tf, input.Network, input.Pk, txArgs)
	if err != nil {
		return err
	}
	return ApproveProposal(input.Tf, input.Network, input.Pk, proposalId)

}
