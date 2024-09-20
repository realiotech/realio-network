package mint

import (
	"context"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/gogo/protobuf/proto"
	assettypes "github.com/realiotech/realio-network/x/asset/types"
)

const priv_name = "mint"

type MintPriviledge struct {
	bk BankKeeper
}

func NewMintPriviledge(bk BankKeeper) MintPriviledge {
	return MintPriviledge{
		bk: bk,
	}
}

func (mp MintPriviledge) Name() string {
	return priv_name
}

func (mp MintPriviledge) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*assettypes.PrivilegeMsgI)(nil),
		&MsgMintToken{},
	)
}

func (mp MintPriviledge) QueryHandler() assettypes.QueryHandler {
	return func(context context.Context, privQuery proto.Message, tokenID string) (proto.Message, error) {
		return nil, nil
	}
}
