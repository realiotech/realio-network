package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateRstStake{}

func NewMsgCreateRstStake(
	creator string,
	index string,
	address string,
	rstAmount int64,
	rioAmount int64,
	incomingRstTxnHash string,
	fundedRioTxnHash string,
	rstOriginChain string,
	rstOriginAddress string,
	created int64,
	status string,

) *MsgCreateRstStake {
	return &MsgCreateRstStake{
		Creator:            creator,
		Index:              index,
		Address:            address,
		RstAmount:          rstAmount,
		RioAmount:          rioAmount,
		IncomingRstTxnHash: incomingRstTxnHash,
		FundedRioTxnHash:   fundedRioTxnHash,
		RstOriginChain:     rstOriginChain,
		RstOriginAddress:   rstOriginAddress,
		Created:            created,
		Status:             status,
	}
}

func (msg *MsgCreateRstStake) Route() string {
	return RouterKey
}

func (msg *MsgCreateRstStake) Type() string {
	return "CreateRstStake"
}

func (msg *MsgCreateRstStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateRstStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateRstStake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateRstStake{}

func NewMsgUpdateRstStake(
	creator string,
	index string,
	address string,
	rstAmount int64,
	rioAmount int64,
	incomingRstTxnHash string,
	fundedRioTxnHash string,
	rstOriginChain string,
	rstOriginAddress string,
	created int64,
	status string,

) *MsgUpdateRstStake {
	return &MsgUpdateRstStake{
		Creator:            creator,
		Index:              index,
		Address:            address,
		RstAmount:          rstAmount,
		RioAmount:          rioAmount,
		IncomingRstTxnHash: incomingRstTxnHash,
		FundedRioTxnHash:   fundedRioTxnHash,
		RstOriginChain:     rstOriginChain,
		RstOriginAddress:   rstOriginAddress,
		Created:            created,
		Status:             status,
	}
}

func (msg *MsgUpdateRstStake) Route() string {
	return RouterKey
}

func (msg *MsgUpdateRstStake) Type() string {
	return "UpdateRstStake"
}

func (msg *MsgUpdateRstStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateRstStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateRstStake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgDeleteRstStake{}

func NewMsgDeleteRstStake(
	creator string,
	index string,

) *MsgDeleteRstStake {
	return &MsgDeleteRstStake{
		Creator: creator,
		Index:   index,
	}
}
func (msg *MsgDeleteRstStake) Route() string {
	return RouterKey
}

func (msg *MsgDeleteRstStake) Type() string {
	return "DeleteRstStake"
}

func (msg *MsgDeleteRstStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteRstStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteRstStake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
