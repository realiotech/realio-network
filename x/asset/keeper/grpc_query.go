package keeper

import (
	"github.com/realiotech/network/x/asset/types"
)

var _ types.QueryServer = Keeper{}
