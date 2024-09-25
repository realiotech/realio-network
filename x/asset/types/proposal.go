package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Proposal types
const (
	ProposalTypeAddTokenManager    string = "AddTokenManager"
	ProposalTypeRemoveTokenManager string = "RemoveTokenManager"
)

var (
	_ govv1beta1.Content = &AddTokenManager{}
	_ govv1beta1.Content = &RemoveTokenManager{}
)

func init() {
	govv1beta1.RegisterProposalType(ProposalTypeAddTokenManager)
	govv1beta1.RegisterProposalType(ProposalTypeRemoveTokenManager)
}

// NewAddAddTokenManager returns new instance of AddTokenManager proposal
func NewAddTokenManager(title, description, manager string) govv1beta1.Content {
	return &AddTokenManager{
		Title:          title,
		Description:    description,
		ManagerAddress: manager,
	}
}

// GetTitle returns the title of a AddTokenManager
func (atmp *AddTokenManager) GetTitle() string { return atmp.Title }

// GetDescription returns the description of a AddTokenManager
func (atmp *AddTokenManager) GetDescription() string { return atmp.Description }

// ProposalRoute returns router key for a AddTokenManager
func (*AddTokenManager) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for a AddTokenManager
func (*AddTokenManager) ProposalType() string {
	return ProposalTypeAddTokenManager
}

// ValidateBasic runs basic stateless validity checks
func (atmp *AddTokenManager) ValidateBasic() error {
	err := govv1beta1.ValidateAbstract(atmp)
	if err != nil {
		return err
	}
	if _, err = sdk.AccAddressFromBech32(atmp.ManagerAddress); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (atmp AddTokenManager) String() string {
	return fmt.Sprintf("AddTokenManager: Title: %s Description: %s Manager: %s", atmp.Title, atmp.Description, atmp.ManagerAddress)
}

// NewRemoveTokenManager returns new instance of RemoveTokenManager
func NewRemoveTokenManager(title, description, manager string) govv1beta1.Content {
	return &RemoveTokenManager{
		Title:          title,
		Description:    description,
		ManagerAddress: manager,
	}
}

// GetTitle returns the title of a RemoveTokenManager
func (rtmp *RemoveTokenManager) GetTitle() string { return rtmp.Title }

// GetDescription returns the description of a RemoveTokenManager
func (rtmp *RemoveTokenManager) GetDescription() string { return rtmp.Description }

// ProposalRoute returns router key for a RemoveTokenManager
func (*RemoveTokenManager) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for a RemoveTokenManager
func (*RemoveTokenManager) ProposalType() string {
	return ProposalTypeRemoveTokenManager
}

// String implements the Stringer interface.
func (rtmp RemoveTokenManager) String() string {
	return fmt.Sprintf("UpdateBondWeightProposal: Title: %s Description: %s Manager: %s", rtmp.Title, rtmp.Description, rtmp.ManagerAddress)
}

// ValidateBasic runs basic stateless validity checks
func (rtmp *RemoveTokenManager) ValidateBasic() error {
	err := govv1beta1.ValidateAbstract(rtmp)
	if err != nil {
		return err
	}
	if _, err = sdk.AccAddressFromBech32(rtmp.ManagerAddress); err != nil {
		return err
	}

	return nil
}
