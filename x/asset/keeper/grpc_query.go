package keeper

import (
	"github.com/realiotech/realio-network/v1/x/asset/types"
)

var _ types.QueryServer = Keeper{}
