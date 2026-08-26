package ante_test

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	osecp256k1 "github.com/cosmos/evm/crypto/ethsecp256k1"
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
