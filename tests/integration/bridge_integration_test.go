package integration

import (
	// "fmt"
	"fmt"
	"testing"

	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	"github.com/realiotech/realio-network/testutil/integration/network"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	integrationutils "github.com/realiotech/realio-network/testutil/integration/utils"
	bridgetypes "github.com/realiotech/realio-network/x/bridge/types"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type BridgeTestSuite struct {
	suite.Suite
	network     network.Network
	grpcHandler grpc.Handler
	factory     factory.TxFactory
	keyring     testkeyring.Keyring
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *BridgeTestSuite) SetupTest() {
	keyring := testkeyring.New(4)
	integrationNetwork := network.New(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	factory := factory.New(integrationNetwork, grpcHandler)

	suite.grpcHandler = grpcHandler
	suite.factory = factory
	suite.network = integrationNetwork
	suite.keyring = keyring
}

func (suite *BridgeTestSuite) TestRegisterNewCoins() {
	testCases := []struct {
		name        string
		msg         *bridgetypes.MsgRegisterNewCoins
		buildGovErr bool
		expectErr   bool
	}{
		{
			name: "valid MsgRegisterNewCoins",
			msg: &bridgetypes.MsgRegisterNewCoins{
				Authority: authtypes.NewModuleAddress("gov").String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin("ario", 1000000),
				),
			},
			expectErr:   false,
			buildGovErr: false,
		},
		{
			name: "invalid MsgRegisterNewCoins; duplicated denom ario",
			msg: &bridgetypes.MsgRegisterNewCoins{
				Authority: authtypes.NewModuleAddress("gov").String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin("ario", 1000000),
				),
			},
			buildGovErr: false,
			expectErr:   true,
		},
		{
			name: "invalid MsgRegisterNewCoins; unauthorized",
			msg: &bridgetypes.MsgRegisterNewCoins{
				Authority: suite.keyring.GetKey(0).AccAddr.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin("eth", 1000000),
				),
			},
			buildGovErr: true,
			expectErr:   true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			senderKey := suite.keyring.GetKey(0)
			proposalID, err := integrationutils.SubmitProposal(
				suite.factory,
				suite.network,
				senderKey.Priv,
				"Register new coins",
				tc.msg,
			)
			if tc.buildGovErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			err = integrationutils.ApproveProposal(
				suite.factory,
				suite.network,
				senderKey.Priv,
				proposalID,
			)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NoError(suite.network.NextBlock())
				client := suite.network.GetBridgeClient()

				res, err := client.RateLimit(suite.network.GetContext(), &bridgetypes.QueryRateLimitRequest{
					Denom: "ario",
				})
				suite.Require().NoError(err)
				suite.Require().Equal(res.Ratelimit.Ratelimit, tc.msg.Coins[0].Amount)
			}
		})
	}
}

func (suite *BridgeTestSuite) TestDeregisterCoins() {
	testCases := []struct {
		name        string
		msg         *bridgetypes.MsgDeregisterCoins
		expectErr   bool
		buildGovErr bool
		errString   string
	}{
		{
			name: "valid MsgDeregisterCoins",
			msg: &bridgetypes.MsgDeregisterCoins{
				Authority: authtypes.NewModuleAddress("gov").String(),
				Denoms:    []string{"ario"},
			},
			buildGovErr: false,
			expectErr:   false,
		},
		{
			name: "invalid MsgDeregisterCoins; coin not in register list",
			msg: &bridgetypes.MsgDeregisterCoins{
				Authority: authtypes.NewModuleAddress("gov").String(),
				Denoms:    []string{"eth"},
			},
			expectErr:   true,
			buildGovErr: false,
			errString:   "denom: eth: coin not in register list",
		},
		{
			name: "invalid MsgDeregisterCoins; unauthorized",
			msg: &bridgetypes.MsgDeregisterCoins{
				Authority: suite.keyring.GetKey(0).AccAddr.String(),
				Denoms:    []string{"ario"},
			},
			buildGovErr: true,
			expectErr:   true,
			errString:   "invalid authority",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			senderKey := suite.keyring.GetKey(0)

			// Register ario first
			proposalID, err := integrationutils.SubmitProposal(
				suite.factory,
				suite.network,
				senderKey.Priv,
				"Register new coins",
				&bridgetypes.MsgRegisterNewCoins{
					Authority: authtypes.NewModuleAddress("gov").String(),
					Coins: sdk.NewCoins(
						sdk.NewInt64Coin("ario", 1000000),
					),
				},
			)
			suite.Require().NoError(err)
			err = integrationutils.ApproveProposal(
				suite.factory,
				suite.network,
				senderKey.Priv,
				proposalID,
			)
			suite.Require().NoError(err)

			proposalID, err = integrationutils.SubmitProposal(
				suite.factory,
				suite.network,
				senderKey.Priv,
				"Deregister coins",
				tc.msg,
			)
			if tc.buildGovErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			err = integrationutils.ApproveProposal(
				suite.factory,
				suite.network,
				senderKey.Priv,
				proposalID,
			)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NoError(suite.network.NextBlock())
				client := suite.network.GetBridgeClient()

				res, err := client.RateLimit(suite.network.GetContext(), &bridgetypes.QueryRateLimitRequest{
					Denom: "ario",
				})
				fmt.Println("res: ", res, err)
				// suite.Require().NoError(err)
				// suite.Require().Equal(res.Ratelimit.Ratelimit, tc.msg.Coins[0].Amount)
			}
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestBridgeTestSuite(t *testing.T) {
	suite.Run(t, new(BridgeTestSuite))
}
