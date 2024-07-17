package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// asset message types
const (
	TypeMsgCreateToken       = "create_token"
	TypeMsgUpdateToken       = "update_token"
	TypeMsgAllocateToken     = "allocate_token"
	TypeMsgAssignPrivilege   = "assign_privilege"
	TypeMsgUnassignPrivilege = "unassign_privilege"
	TypeMsgDisablePrivilege  = "disable_privilege"
	TypeMsgExecutePrivilege  = "execute_privilege"
)

var (
	_ sdk.Msg = &MsgCreateToken{}
	_ sdk.Msg = &MsgUpdateToken{}
	_ sdk.Msg = &MsgAllocateToken{}
	_ sdk.Msg = &MsgAssignPrivilege{}
	_ sdk.Msg = &MsgUnassignPrivilege{}
	_ sdk.Msg = &MsgDisablePrivilege{}
	_ sdk.Msg = &MsgExecutePrivilege{}
)

func NewMsgCreateToken(creator string,
	manager string,
	name string,
	symbol string,
	decimal uint32,
	description string,
	executePrivileges []string,
	addNewPrivilege bool,
) *MsgCreateToken {
	return &MsgCreateToken{
		Creator:            creator,
		Manager:            manager,
		Name:               name,
		Symbol:             symbol,
		Decimal:            decimal,
		Description:        description,
		ExcludedPrivileges: executePrivileges,
		AddNewPrivilege:    addNewPrivilege,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgCreateToken) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgCreateToken) Type() string { return TypeMsgCreateToken }

// GetSigners implements the sdk.Msg interface.
func (msg MsgCreateToken) GetSigners() []sdk.AccAddress {
	creatorAddr, _ := sdk.ValAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{sdk.AccAddress(creatorAddr)}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgCreateToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgCreateToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid creator address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Manager); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid manager address: %s", err)
	}
	return nil
}

func NewMsgUpdateToken(
	manager string,
	name string,
	symbol string,
	description string,
	addNewPrivilege bool,
) *MsgUpdateToken {
	return &MsgUpdateToken{
		Manager:         manager,
		Name:            name,
		Symbol:          symbol,
		Description:     description,
		AddNewPrivilege: addNewPrivilege,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUpdateToken) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgUpdateToken) Type() string { return TypeMsgUpdateToken }

// GetSigners implements the sdk.Msg interface.
func (msg MsgUpdateToken) GetSigners() []sdk.AccAddress {
	managerAddr, _ := sdk.ValAddressFromBech32(msg.Manager)
	return []sdk.AccAddress{sdk.AccAddress(managerAddr)}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgUpdateToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgUpdateToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Manager); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid manager address: %s", err)
	}
	return nil
}

func NewMsgAllocateToken(
	manager string,
	tokenId string,
	balances []banktypes.Balance,
	vestingBalance []*codectypes.Any,
) *MsgAllocateToken {
	return &MsgAllocateToken{
		Manager:        manager,
		TokenId:        tokenId,
		Balances:       balances,
		VestingBalance: vestingBalance,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgAllocateToken) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgAllocateToken) Type() string { return TypeMsgAllocateToken }

// GetSigners implements the sdk.Msg interface.
func (msg MsgAllocateToken) GetSigners() []sdk.AccAddress {
	managerAddr, _ := sdk.ValAddressFromBech32(msg.Manager)
	return []sdk.AccAddress{sdk.AccAddress(managerAddr)}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgAllocateToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgAllocateToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Manager); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid manager address: %s", err)
	}
	return nil
}

func NewMsgAssignPrivilege(
	manager string,
	tokenId string,
	assignTo []string,
	privilege string,
) *MsgAssignPrivilege {
	return &MsgAssignPrivilege{
		Manager:    manager,
		TokenId:    tokenId,
		AssignedTo: assignTo,
		Privilege:  privilege,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgAssignPrivilege) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgAssignPrivilege) Type() string { return TypeMsgAssignPrivilege }

// GetSigners implements the sdk.Msg interface.
func (msg MsgAssignPrivilege) GetSigners() []sdk.AccAddress {
	managerAddr, _ := sdk.ValAddressFromBech32(msg.Manager)
	return []sdk.AccAddress{sdk.AccAddress(managerAddr)}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgAssignPrivilege) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgAssignPrivilege) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Manager); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid manager address: %s", err)
	}

	for _, assignee := range msg.AssignedTo {
		if _, err := sdk.AccAddressFromBech32(assignee); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid assignee address: %s", err)
		}
	}
	return nil
}

func NewMsgUnassignPrivilege(
	manager string,
	tokenId string,
	unassignedFrom []string,
	privilege string,
) *MsgUnassignPrivilege {
	return &MsgUnassignPrivilege{
		Manager:        manager,
		TokenId:        tokenId,
		UnassignedFrom: unassignedFrom,
		Privilege:      privilege,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUnassignPrivilege) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgUnassignPrivilege) Type() string { return TypeMsgUnassignPrivilege }

// GetSigners implements the sdk.Msg interface.
func (msg MsgUnassignPrivilege) GetSigners() []sdk.AccAddress {
	managerAddr, _ := sdk.ValAddressFromBech32(msg.Manager)
	return []sdk.AccAddress{sdk.AccAddress(managerAddr)}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgUnassignPrivilege) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgUnassignPrivilege) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Manager); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid manager address: %s", err)
	}

	for _, assignee := range msg.UnassignedFrom {
		if _, err := sdk.AccAddressFromBech32(assignee); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid assignee address: %s", err)
		}
	}
	return nil
}

func NewMsgDisablePrivilege(
	manager string,
	tokenId string,
	privilege string,
) *MsgDisablePrivilege {
	return &MsgDisablePrivilege{
		Manager:           manager,
		TokenId:           tokenId,
		DisabledPrivilege: privilege,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgDisablePrivilege) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgDisablePrivilege) Type() string { return TypeMsgDisablePrivilege }

// GetSigners implements the sdk.Msg interface.
func (msg MsgDisablePrivilege) GetSigners() []sdk.AccAddress {
	managerAddr, _ := sdk.ValAddressFromBech32(msg.Manager)
	return []sdk.AccAddress{sdk.AccAddress(managerAddr)}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgDisablePrivilege) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgDisablePrivilege) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Manager); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid manager address: %s", err)
	}

	return nil
}

func NewMsgExecutePrivilege(
	address string,
	tokenId string,
	privilegeMsg *codectypes.Any,
) *MsgExecutePrivilege {
	return &MsgExecutePrivilege{
		Address:      address,
		TokenId:      tokenId,
		PrivilegeMsg: privilegeMsg,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgExecutePrivilege) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgExecutePrivilege) Type() string { return TypeMsgExecutePrivilege }

// GetSigners implements the sdk.Msg interface.
func (msg MsgExecutePrivilege) GetSigners() []sdk.AccAddress {
	singerAddr, _ := sdk.ValAddressFromBech32(msg.Address)
	return []sdk.AccAddress{sdk.AccAddress(singerAddr)}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgExecutePrivilege) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgExecutePrivilege) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid signer address: %s", err)
	}

	return nil
}
