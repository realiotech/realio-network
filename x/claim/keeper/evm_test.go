package keeper_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/evm/contracts"
	"github.com/cosmos/evm/x/vm/statedb"

	"github.com/realiotech/realio-network/testutil"
	"github.com/realiotech/realio-network/x/claim/keeper"
	"github.com/realiotech/realio-network/x/claim/types"
)

var erc20TestABI = contracts.ERC20MinterBurnerDecimalsContract.ABI

// ---------------------------------------------------------------------------
// EVM test fixtures
// ---------------------------------------------------------------------------

// ensureAccount gives addr an auth account. CallEVM reads the caller's
// nonce through the account keeper, which errors on an address that has
// never been seen -- and a freshly generated test address has not been.
func (suite *KeeperTestSuite) ensureAccount(addr sdk.AccAddress) {
	if suite.app.AccountKeeper.GetAccount(suite.ctx, addr) == nil {
		suite.app.AccountKeeper.SetAccount(suite.ctx, suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
	}
}

// deployERC20 deploys the standard mintable ERC-20 x/erc20 ships, with
// deployer holding MINTER_ROLE, and returns where it landed.
func (suite *KeeperTestSuite) deployERC20(deployer sdk.AccAddress) common.Address {
	suite.ensureAccount(deployer)
	from := common.BytesToAddress(deployer.Bytes())

	nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, deployer)
	suite.Require().NoError(err)

	ctorArgs, err := erc20TestABI.Pack("", "ClaimTest", "CLM", uint8(18))
	suite.Require().NoError(err)
	initCode := make([]byte, 0, len(contracts.ERC20MinterBurnerDecimalsContract.Bin)+len(ctorArgs))
	initCode = append(initCode, contracts.ERC20MinterBurnerDecimalsContract.Bin...)
	initCode = append(initCode, ctorArgs...)

	stateDB := statedb.New(suite.ctx, suite.app.EvmKeeper, statedb.NewEmptyTxConfig())
	_, err = suite.app.EvmKeeper.CallEVMWithData(suite.ctx, stateDB, from, nil, initCode, true, false, nil)
	suite.Require().NoError(err)

	return crypto.CreateAddress(from, nonce)
}

// callContract makes a committing EVM call from caller, and requires it to
// succeed.
func (suite *KeeperTestSuite) callContract(caller sdk.AccAddress, contract common.Address, contractABI abi.ABI, method string, args ...interface{}) {
	suite.ensureAccount(caller)
	stateDB := statedb.New(suite.ctx, suite.app.EvmKeeper, statedb.NewEmptyTxConfig())
	_, err := suite.app.EvmKeeper.CallEVM(
		suite.ctx, stateDB, contractABI, common.BytesToAddress(caller.Bytes()),
		contract, true, false, nil, method, args...,
	)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) erc20Mint(minter sdk.AccAddress, contract common.Address, to sdk.AccAddress, amount int64) {
	suite.callContract(minter, contract, erc20TestABI, "mint", common.BytesToAddress(to.Bytes()), big.NewInt(amount))
}

func (suite *KeeperTestSuite) erc20BalanceOf(contract common.Address, account sdk.AccAddress) *big.Int {
	balance := suite.app.Erc20Keeper.BalanceOf(suite.ctx, erc20TestABI, contract, common.BytesToAddress(account.Bytes()))
	suite.Require().NotNil(balance, "balanceOf call failed")
	return balance
}

// link records old -> new without going through MsgLinkAddress, so the
// token tests do not drag delegation migration in with them.
func (suite *KeeperTestSuite) link(oldAddr, newAddr sdk.AccAddress) {
	suite.Require().NoError(suite.app.ClaimKeeper.SetLink(suite.ctx, oldAddr, newAddr))
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func (suite *KeeperTestSuite) TestClaimTokensERC20() {
	deployer := testutil.GenAddress()
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()

	contract := suite.deployERC20(deployer)
	suite.erc20Mint(deployer, contract, oldAddr, 1_000)
	suite.link(oldAddr, newAddr)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	res, err := srv.ClaimTokens(suite.ctx, &types.MsgClaimTokens{
		Admin:          suite.admin,
		OldAddress:     oldAddr.String(),
		Erc20Contracts: []string{contract.Hex()},
	})
	suite.Require().NoError(err)

	// The destination comes from the link, never from the message.
	suite.Require().Equal(newAddr.String(), res.NewAddress)
	suite.Require().Equal(uint64(1), res.Erc20ContractsSwept)

	suite.Require().Zero(suite.erc20BalanceOf(contract, oldAddr).Sign())
	suite.Require().Equal(big.NewInt(1_000), suite.erc20BalanceOf(contract, newAddr))
}

// TestClaimTokensERC20MergesIntoExistingBalance covers the new address
// already holding some of the same token: the sweep adds to it rather than
// needing the destination to be empty.
func (suite *KeeperTestSuite) TestClaimTokensERC20MergesIntoExistingBalance() {
	deployer := testutil.GenAddress()
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()

	contract := suite.deployERC20(deployer)
	suite.erc20Mint(deployer, contract, oldAddr, 1_000)
	suite.erc20Mint(deployer, contract, newAddr, 250)
	suite.link(oldAddr, newAddr)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.ClaimTokens(suite.ctx, &types.MsgClaimTokens{
		Admin:          suite.admin,
		OldAddress:     oldAddr.String(),
		Erc20Contracts: []string{contract.Hex()},
	})
	suite.Require().NoError(err)

	suite.Require().Zero(suite.erc20BalanceOf(contract, oldAddr).Sign())
	suite.Require().Equal(big.NewInt(1_250), suite.erc20BalanceOf(contract, newAddr))
}

// TestClaimTokensERC20ZeroBalance covers a contract the admin listed on
// spec: holding nothing is not an error, it just doesn't count as swept.
func (suite *KeeperTestSuite) TestClaimTokensERC20ZeroBalance() {
	deployer := testutil.GenAddress()
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()

	contract := suite.deployERC20(deployer)
	suite.link(oldAddr, newAddr)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	res, err := srv.ClaimTokens(suite.ctx, &types.MsgClaimTokens{
		Admin:          suite.admin,
		OldAddress:     oldAddr.String(),
		Erc20Contracts: []string{contract.Hex()},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), res.Erc20ContractsSwept)
}

// TestClaimTokensSweepsSeveralContracts: each listed contract is drained
// independently, and the count reports only the ones that held something.
func (suite *KeeperTestSuite) TestClaimTokensSweepsSeveralContracts() {
	deployer := testutil.GenAddress()
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()

	held := suite.deployERC20(deployer)
	alsoHeld := suite.deployERC20(deployer)
	empty := suite.deployERC20(deployer)
	suite.erc20Mint(deployer, held, oldAddr, 500)
	suite.erc20Mint(deployer, alsoHeld, oldAddr, 7)
	suite.link(oldAddr, newAddr)

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	res, err := srv.ClaimTokens(suite.ctx, &types.MsgClaimTokens{
		Admin:          suite.admin,
		OldAddress:     oldAddr.String(),
		Erc20Contracts: []string{held.Hex(), empty.Hex(), alsoHeld.Hex()},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(2), res.Erc20ContractsSwept)

	suite.Require().Equal(big.NewInt(500), suite.erc20BalanceOf(held, newAddr))
	suite.Require().Equal(big.NewInt(7), suite.erc20BalanceOf(alsoHeld, newAddr))
}

// TestClaimTokensFailsAtomically is the point of making this its own
// message: one bad entry undoes the whole sweep, so the admin never has to
// work out how far down the list it got.
func (suite *KeeperTestSuite) TestClaimTokensFailsAtomically() {
	deployer := testutil.GenAddress()
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()

	contract := suite.deployERC20(deployer)
	suite.erc20Mint(deployer, contract, oldAddr, 500)
	suite.link(oldAddr, newAddr)

	// Second entry is a mistyped address with nothing deployed at it.
	notAContract := common.BytesToAddress(testutil.GenAddress().Bytes())

	attemptCtx, _ := suite.ctx.CacheContext()
	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.ClaimTokens(attemptCtx, &types.MsgClaimTokens{
		Admin:          suite.admin,
		OldAddress:     oldAddr.String(),
		Erc20Contracts: []string{contract.Hex(), notAContract.Hex()},
	})
	suite.Require().ErrorContains(err, "has no contract code")

	// The first contract really did move, inside the failed attempt's
	// branch...
	suite.Require().Zero(suite.app.Erc20Keeper.BalanceOf(
		attemptCtx, erc20TestABI, contract, common.BytesToAddress(oldAddr.Bytes()),
	).Sign())
	// ...but that branch was never written back, which is what runMsgs does
	// to a message that returns an error.
	suite.Require().Equal(big.NewInt(500), suite.erc20BalanceOf(contract, oldAddr))
}

// TestClaimTokensSweepsAddressWithNoAuthAccount covers an address that only
// ever received tokens: x/auth has never heard of it, so the EVM call made
// on its behalf has no nonce to read unless the claim creates the account
// first.
func (suite *KeeperTestSuite) TestClaimTokensSweepsAddressWithNoAuthAccount() {
	deployer := testutil.GenAddress()
	oldAddr, newAddr := testutil.GenAddress(), testutil.GenAddress()

	contract := suite.deployERC20(deployer)
	suite.erc20Mint(deployer, contract, oldAddr, 42)
	suite.link(oldAddr, newAddr)
	suite.Require().Nil(suite.app.AccountKeeper.GetAccount(suite.ctx, oldAddr),
		"fixture assumption broken: oldAddr should have no auth account yet")

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	_, err := srv.ClaimTokens(suite.ctx, &types.MsgClaimTokens{
		Admin:          suite.admin,
		OldAddress:     oldAddr.String(),
		Erc20Contracts: []string{contract.Hex()},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(big.NewInt(42), suite.erc20BalanceOf(contract, newAddr))
}

func (suite *KeeperTestSuite) TestClaimTokensValidation() {
	deployer := testutil.GenAddress()
	linkedOld, linkedNew := testutil.GenAddress(), testutil.GenAddress()
	unlinked := testutil.GenAddress()
	contract := suite.deployERC20(deployer)
	suite.link(linkedOld, linkedNew)

	testCases := []struct {
		name      string
		msg       *types.MsgClaimTokens
		errString string
	}{
		{
			name:      "wrong admin",
			msg:       &types.MsgClaimTokens{Admin: testutil.GenAddress().String(), OldAddress: linkedOld.String(), Erc20Contracts: []string{contract.Hex()}},
			errString: "invalid admin",
		},
		{
			name:      "old_address not linked",
			msg:       &types.MsgClaimTokens{Admin: suite.admin, OldAddress: unlinked.String(), Erc20Contracts: []string{contract.Hex()}},
			errString: "is not linked",
		},
		{
			name:      "nothing listed",
			msg:       &types.MsgClaimTokens{Admin: suite.admin, OldAddress: linkedOld.String()},
			errString: "must not be empty",
		},
		{
			name:      "not a hex address",
			msg:       &types.MsgClaimTokens{Admin: suite.admin, OldAddress: linkedOld.String(), Erc20Contracts: []string{"nope"}},
			errString: "is not a hex address",
		},
		{
			name:      "nothing deployed at that address",
			msg:       &types.MsgClaimTokens{Admin: suite.admin, OldAddress: linkedOld.String(), Erc20Contracts: []string{common.BytesToAddress(testutil.GenAddress().Bytes()).Hex()}},
			errString: "has no contract code",
		},
	}

	srv := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			attemptCtx, _ := suite.ctx.CacheContext()
			_, err := srv.ClaimTokens(attemptCtx, tc.msg)
			suite.Require().ErrorContains(err, tc.errString)
		})
	}
}
