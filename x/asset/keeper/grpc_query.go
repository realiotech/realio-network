package keeper

import (
	"github.com/realiotech/realio-network/x/v1/asset/types"
)

var _ types.QueryServer = Keeper{}
