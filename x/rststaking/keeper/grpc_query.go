package keeper

import (
	"github.com/realiotech/realio-network/x/rststaking/types"
)

var _ types.QueryServer = Keeper{}
