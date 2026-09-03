package ante_test

import (
	"math/big"

	"cosmossdk.io/x/feegrant"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	osecp256k1 "github.com/cosmos/evm/crypto/ethsecp256k1"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"

	realioante "github.com/realiotech/realio-network/app/ante"
)

// TestBlacklistDecorator verifies that a transaction signed by a blacklisted
// address is rejected outright at CheckTx, while a transaction from a
// non-blacklisted address passes the blacklist check (it may still fail
// later for unrelated reasons, e.g. insufficient funds, but must NOT be
// rejected as blacklisted).
func (suite *AnteTestSuite) TestBlacklistDecorator() {
	suite.SetupTest()

	testPrivKeys, testAddresses, err := generatePrivKeyAddressPairs(2)
	suite.Require().NoError(err)
	blacklistedAddr := testAddresses[0]
	cleanAddr := testAddresses[1]

	// Simulate this address being one of the leaked/compromised keys, seeded
	// into the x/blacklist module's on-chain state (not a compiled-in list).
	// CheckTx reads from the app's checkState branch, not suite.ctx (which
	// wraps deliverState) — write through a context over checkState instead
	// so the seeded entry is actually visible to the CheckTx calls below.
	checkCtx := suite.app.BaseApp.NewContextLegacy(true, suite.ctx.BlockHeader())
	err = suite.app.BlacklistKeeper.SetBlacklisted(checkCtx, blacklistedAddr)
	suite.Require().NoError(err)

	testcases := []struct {
		name        string
		from        sdk.AccAddress
		priv        *osecp256k1.PrivKey
		expectBlock bool // expect the ante handler to reject it as blacklisted
	}{
		{
			name:        "tx signed by a blacklisted address is rejected",
			from:        blacklistedAddr,
			priv:        testPrivKeys[0],
			expectBlock: true,
		},
		{
			name:        "tx signed by a clean address is NOT rejected as blacklisted",
			from:        cleanAddr,
			priv:        testPrivKeys[1],
			expectBlock: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			msg := banktypes.NewMsgSend(tc.from, blacklistedAddr, sdk.NewCoins(sdk.NewInt64Coin(suite.denom, 1)))
			tx, err := createTx(tc.priv, msg)
			suite.Require().NoError(err)

			txEncoder := suite.clientCtx.TxConfig.TxEncoder()
			bz, err := txEncoder(tx)
			suite.Require().NoError(err)

			resCheckTx, err := suite.app.CheckTx(&abci.RequestCheckTx{
				Tx:   bz,
				Type: abci.CheckTxType_New,
			})
			suite.Require().NoError(err)

			if tc.expectBlock {
				suite.Require().Equal(sdkerrors.ErrUnauthorized.ABCICode(), resCheckTx.Code, resCheckTx.Log)
				suite.Require().Contains(resCheckTx.Log, "blacklisted")
			} else {
				suite.Require().NotEqual(sdkerrors.ErrUnauthorized.ABCICode(), resCheckTx.Code, resCheckTx.Log)
			}
		})
	}
}

// TestBlacklistDecoratorAuthzExec verifies that a clean grantee cannot use an
// existing authz grant to execute a message on behalf of a blacklisted granter.
func (suite *AnteTestSuite) TestBlacklistDecoratorAuthzExec() {
	suite.SetupTest()

	testPrivKeys, testAddresses, err := generatePrivKeyAddressPairs(3)
	suite.Require().NoError(err)
	blacklistedGranter := testAddresses[0]
	cleanGrantee := testAddresses[1]
	recipient := testAddresses[2]

	checkCtx := suite.app.BaseApp.NewContextLegacy(true, suite.ctx.BlockHeader())
	err = suite.app.BlacklistKeeper.SetBlacklisted(checkCtx, blacklistedGranter)
	suite.Require().NoError(err)

	testcases := []struct {
		name        string
		msg         sdk.Msg
		expectBlock bool
	}{
		{
			name: "clean grantee executing for blacklisted granter is rejected",
			msg: newMsgExec(cleanGrantee, []sdk.Msg{
				banktypes.NewMsgSend(blacklistedGranter, recipient, sdk.NewCoins(sdk.NewInt64Coin(suite.denom, 1))),
			}),
			expectBlock: true,
		},
		{
			name: "nested MsgExec for blacklisted granter is rejected",
			msg: createNestedMsgExec(cleanGrantee, 2, []sdk.Msg{
				banktypes.NewMsgSend(blacklistedGranter, recipient, sdk.NewCoins(sdk.NewInt64Coin(suite.denom, 1))),
			}),
			expectBlock: true,
		},
		{
			name: "clean grantee executing for clean granter is not rejected as blacklisted",
			msg: newMsgExec(cleanGrantee, []sdk.Msg{
				banktypes.NewMsgSend(cleanGrantee, recipient, sdk.NewCoins(sdk.NewInt64Coin(suite.denom, 1))),
			}),
			expectBlock: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			tx, err := createTx(testPrivKeys[1], tc.msg)
			suite.Require().NoError(err)

			txEncoder := suite.clientCtx.TxConfig.TxEncoder()
			bz, err := txEncoder(tx)
			suite.Require().NoError(err)

			resCheckTx, err := suite.app.CheckTx(&abci.RequestCheckTx{
				Tx:   bz,
				Type: abci.CheckTxType_New,
			})
			suite.Require().NoError(err)

			if tc.expectBlock {
				suite.Require().Equal(sdkerrors.ErrUnauthorized.ABCICode(), resCheckTx.Code, resCheckTx.Log)
				suite.Require().Contains(resCheckTx.Log, blacklistedGranter.String())
				suite.Require().Contains(resCheckTx.Log, "blacklisted")
			} else {
				suite.Require().NotContains(resCheckTx.Log, "blacklisted")
			}
		})
	}
}

// TestEVMBlacklistDecoratorSetCodeAuthority verifies that a clean EVM sender
// cannot submit an EIP-7702 authorization signed by a blacklisted authority.
func (suite *AnteTestSuite) TestEVMBlacklistDecoratorSetCodeAuthority() {
	suite.SetupTest()

	privKeys, addresses, err := generatePrivKeyAddressPairs(3)
	suite.Require().NoError(err)
	blacklistedAuthority := addresses[0]
	cleanAuthority := addresses[1]
	cleanSender := addresses[2]

	err = suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, blacklistedAuthority)
	suite.Require().NoError(err)

	for _, tc := range []struct {
		name         string
		authority    sdk.AccAddress
		authorityKey *osecp256k1.PrivKey
		expectBlock  bool
	}{
		{"blacklisted SetCode authority is rejected", blacklistedAuthority, privKeys[0], true},
		{"clean SetCode authority is accepted", cleanAuthority, privKeys[1], false},
	} {
		suite.Run(tc.name, func() {
			chainID := evmtypes.GetEthChainConfig().ChainID
			authorization := ethtypes.SetCodeAuthorization{
				ChainID: *uint256.MustFromBig(chainID),
				Address: common.HexToAddress("0x0000000000000000000000000000000000001234"),
				Nonce:   0,
			}
			ecdsaKey, err := tc.authorityKey.ToECDSA()
			suite.Require().NoError(err)
			signedAuthorization, err := ethtypes.SignSetCode(ecdsaKey, authorization)
			suite.Require().NoError(err)

			to := common.Address{}
			msg := evmtypes.NewTx(&evmtypes.EvmTxArgs{
				GasLimit:          100_000,
				GasFeeCap:         big.NewInt(1),
				GasTipCap:         big.NewInt(1),
				ChainID:           chainID,
				To:                &to,
				AuthorizationList: []ethtypes.SetCodeAuthorization{signedAuthorization},
			})
			msg.From = cleanSender.Bytes()
			builder := suite.CreateEthTestTxBuilder(msg)

			next := &MockAnteHandler{}
			decorator := realioante.NewEVMBlacklistDecorator(suite.app.BlacklistKeeper)
			_, err = decorator.AnteHandle(suite.ctx, builder.GetTx(), false, next.AnteHandle)

			if tc.expectBlock {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.authority.String())
				suite.Require().False(next.WasCalled)
			} else {
				suite.Require().NoError(err)
				suite.Require().True(next.WasCalled)
			}
		})
	}
}

// TestBlacklistDecoratorFeeGranter verifies that a tx whose AuthInfo.Fee.Granter
// is a blacklisted address is rejected, even though the granter never signs
// this tx — the msg signer (grantee) is a clean, unrelated address. This is
// the gap BlacklistSendRestriction deliberately leaves open (it exempts
// module-account destinations, and the fee-collector is one), closed here
// instead: DeductFeeDecorator (x/auth/ante/fee.go) reads this same field and
// debits the granter regardless of whether the granter signed anything.
func (suite *AnteTestSuite) TestBlacklistDecoratorFeeGranter() {
	suite.SetupTest()

	testPrivKeys, testAddresses, err := generatePrivKeyAddressPairs(3)
	suite.Require().NoError(err)
	blacklistedGranter := testAddresses[0]
	cleanGrantee := testAddresses[1]
	recipient := testAddresses[2]

	checkCtx := suite.app.BaseApp.NewContextLegacy(true, suite.ctx.BlockHeader())
	err = suite.app.BlacklistKeeper.SetBlacklisted(checkCtx, blacklistedGranter)
	suite.Require().NoError(err)

	testcases := []struct {
		name        string
		granter     sdk.AccAddress
		expectBlock bool
	}{
		{
			name:        "tx fee-granted by a blacklisted address is rejected",
			granter:     blacklistedGranter,
			expectBlock: true,
		},
		{
			name:        "tx fee-granted by a clean address is not rejected as blacklisted",
			granter:     cleanGrantee,
			expectBlock: false,
		},
		{
			name:        "tx with no fee granter at all is not rejected as blacklisted",
			granter:     nil,
			expectBlock: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			msg := banktypes.NewMsgSend(cleanGrantee, recipient, sdk.NewCoins(sdk.NewInt64Coin(suite.denom, 1)))
			tx, err := createTxWithFeeGranter(testPrivKeys[1], tc.granter, msg)
			suite.Require().NoError(err)

			txEncoder := suite.clientCtx.TxConfig.TxEncoder()
			bz, err := txEncoder(tx)
			suite.Require().NoError(err)

			resCheckTx, err := suite.app.CheckTx(&abci.RequestCheckTx{
				Tx:   bz,
				Type: abci.CheckTxType_New,
			})
			suite.Require().NoError(err)

			if tc.expectBlock {
				suite.Require().Equal(sdkerrors.ErrUnauthorized.ABCICode(), resCheckTx.Code, resCheckTx.Log)
				suite.Require().Contains(resCheckTx.Log, blacklistedGranter.String())
				suite.Require().Contains(resCheckTx.Log, "blacklisted")
			} else {
				suite.Require().NotContains(resCheckTx.Log, "blacklisted")
			}
		})
	}
}

// TestEVMBlacklistDecoratorFeeSponsor verifies the EVM-specific feegrant path:
// EVMMonoDecorator records the granter it actually used in an event before the
// blacklist decorator runs.
func (suite *AnteTestSuite) TestEVMBlacklistDecoratorFeeSponsor() {
	suite.SetupTest()

	_, addresses, err := generatePrivKeyAddressPairs(2)
	suite.Require().NoError(err)
	blacklistedSponsor := addresses[0]
	cleanSender := addresses[1]
	err = suite.app.BlacklistKeeper.SetBlacklisted(suite.ctx, blacklistedSponsor)
	suite.Require().NoError(err)

	ctx := suite.ctx.WithEventManager(sdk.NewEventManager())
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		feegrant.EventTypeUseFeeGrant,
		sdk.NewAttribute(feegrant.AttributeKeyGranter, blacklistedSponsor.String()),
		sdk.NewAttribute(feegrant.AttributeKeyGrantee, cleanSender.String()),
	))

	to := common.Address{}
	msg := suite.BuildTestEthTx(common.BytesToAddress(cleanSender), to, big.NewInt(1), nil, nil, nil)
	builder := suite.CreateEthTestTxBuilder(msg)
	next := &MockAnteHandler{}
	decorator := realioante.NewEVMBlacklistDecorator(suite.app.BlacklistKeeper)
	_, err = decorator.AnteHandle(ctx, builder.GetTx(), false, next.AnteHandle)

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), blacklistedSponsor.String())
	suite.Require().False(next.WasCalled)
}
