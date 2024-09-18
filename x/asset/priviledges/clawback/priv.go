package clawback

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
)

const priv_name = "clawback"

type ClawbackPriviledge struct {
	bk BankKeeper
}

func NewClawbackPriviledge(bk BankKeeper) ClawbackPriviledge {
	return ClawbackPriviledge{
		bk: bk,
	}
}

func (cp ClawbackPriviledge) Name() string {
	return priv_name
}

func (cp ClawbackPriviledge) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*assettypes.PrivilegeMsgI)(nil),
		&MsgClawbackToken{},
	)
}
