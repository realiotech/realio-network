package keeper

import (
	"github.com/realiotech/realio-network/x/asset/types"
)

var _ types.QueryServer = Keeper{}
