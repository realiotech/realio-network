package types

import (
	errorsmod "cosmossdk.io/errors"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	evmostypes "github.com/evmos/os/types"
)

// constants
const (
	// ProposalTypeRegisterCoin is DEPRECATED, remove after v16 upgrade
	ProposalTypeRegisterCoin          string = "RegisterCoin"
	ProposalTypeRegisterERC20Owner    string = "RegisterERC20Owner"
	ProposalTypeToggleTokenConversion string = "ToggleTokenConversion" // #nosec
)

// Implements Proposal Interface
var (
	// RegisterCoinProposal is DEPRECATED, remove after v16 upgrade
	_ v1beta1.Content = &RegisterERC20OwnerProposal{}
)

func init() {
	v1beta1.RegisterProposalType(ProposalTypeRegisterERC20Owner)
}

// NewRegisterERC20OwnerProposal returns new instance of RegisterERC20OwnerProposal
func NewRegisterERC20OwnerProposal(title, description string, erc20Address, ownerAddress string) v1beta1.Content {
	return &RegisterERC20OwnerProposal{
		Title:        title,
		Description:  description,
		Erc20Address: erc20Address,
		Owner:        ownerAddress,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterERC20OwnerProposal) ProposalRoute() string { return ModuleName }

// ProposalType returns proposal type for this proposal
func (*RegisterERC20OwnerProposal) ProposalType() string {
	return ProposalTypeRegisterERC20Owner
}

// ValidateBasic performs a stateless check of the proposal fields
func (rtbp *RegisterERC20OwnerProposal) ValidateBasic() error {
	if err := evmostypes.ValidateAddress(rtbp.Erc20Address); err != nil {
		return errorsmod.Wrap(err, "ERC20 address")
	}

	if err := evmostypes.ValidateAddress(rtbp.Owner); err != nil {
		return errorsmod.Wrap(err, "ERC20 address")
	}

	return v1beta1.ValidateAbstract(rtbp)
}
