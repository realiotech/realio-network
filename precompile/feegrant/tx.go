package feegrant

import (
	"fmt"

	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// GrantEVM handles the grant of a fee allowance from the caller to a grantee.
func (p Precompile) GrantEVM(
	ctx sdk.Context,
	origin common.Address,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, err := NewGrantRequest(origin, args)
	if err != nil {
		return nil, err
	}

	// Execute grant using feegrant msgServer
	msgServer := feegrantkeeper.NewMsgServerImpl(p.feegrantKeeper)
	_, err = msgServer.GrantAllowance(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("feegrant grant failed: %v", err)
	}

	// Return success
	return method.Outputs.Pack(true)
}

// RevokeEVM handles the revocation of a fee allowance from the caller to a grantee
func (p Precompile) RevokeEVM(
	ctx sdk.Context,
	origin common.Address,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, err := NewRevokeRequest(origin, args)
	if err != nil {
		return nil, err
	}

	// Execute revoke using feegrant msgServer
	msgServer := feegrantkeeper.NewMsgServerImpl(p.feegrantKeeper)
	revokeRes, err := msgServer.RevokeAllowance(ctx, msg)
	fmt.Println("revokeRes", revokeRes)
	if err != nil {
		return nil, fmt.Errorf("feegrant revoke failed: %v", err)
	}

	// Return success
	return method.Outputs.Pack(true)
}
