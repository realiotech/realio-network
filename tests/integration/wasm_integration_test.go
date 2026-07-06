package integration

import (
	"errors"
	"fmt"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commonfactory "github.com/cosmos/evm/testutil/integration/base/factory"
	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/suite"

	contracts "github.com/realiotech/realio-network/test-contracts"
	"github.com/realiotech/realio-network/testutil/integration/network"
)

// WasmTestSuite exercises the CosmWasm module end to end: a user stores a
// contract's bytecode, instantiates it, and then interacts with it via
// execute and smart query.
type WasmTestSuite struct {
	suite.Suite
	network     network.Network
	grpcHandler grpc.Handler
	factory     factory.TxFactory
	keyring     testkeyring.Keyring
}

func (suite *WasmTestSuite) SetupTest() {
	configurator := evmtypes.NewEVMConfigurator()
	configurator.ResetTestConfig()
	keyring := testkeyring.New(2)
	integrationNetwork := network.New(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	suite.grpcHandler = grpcHandler
	suite.factory = txFactory
	suite.network = integrationNetwork
	suite.keyring = keyring
}

func TestWasmTestSuite(t *testing.T) {
	suite.Run(t, new(WasmTestSuite))
}

// TestStoreInstantiateExecuteContract walks through the full lifecycle of a
// CosmWasm contract:
//  1. the deployer uploads the "hackatom" contract's bytecode (store code)
//  2. the deployer instantiates it, designating itself as verifier and a
//     second account as beneficiary, and funds it
//  3. anyone can query the contract's state (verifier)
//  4. only the verifier can execute "release", which pays out the
//     contract's balance to the beneficiary
func (suite *WasmTestSuite) TestStoreInstantiateExecuteContract() {
	deployerPriv := suite.keyring.GetPrivKey(0)
	deployerAddr := suite.keyring.GetAccAddr(0)
	beneficiaryAddr := suite.keyring.GetAccAddr(1)
	denom := suite.network.GetBaseDenom()

	// 1. store the contract's wasm bytecode
	storeMsg := &wasmtypes.MsgStoreCode{
		Sender:       deployerAddr.String(),
		WASMByteCode: contracts.HackatomWasm,
	}
	storeRes, err := suite.factory.ExecuteCosmosTx(deployerPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{storeMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(storeRes.IsOK(), "store code should have succeeded: %s", storeRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	codeID, err := getCodeIDFromEvents(storeRes.Events)
	suite.Require().NoError(err)
	suite.Require().NotZero(codeID)

	// 2. instantiate the contract, funding it and setting up verifier/beneficiary
	initMsg := []byte(fmt.Sprintf(`{"verifier":"%s","beneficiary":"%s"}`, deployerAddr.String(), beneficiaryAddr.String()))
	contractFunds := sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000))
	instantiateMsg := &wasmtypes.MsgInstantiateContract{
		Sender: deployerAddr.String(),
		Admin:  deployerAddr.String(),
		CodeID: codeID,
		Label:  "hackatom integration test",
		Msg:    wasmtypes.RawContractMessage(initMsg),
		Funds:  contractFunds,
	}
	instantiateRes, err := suite.factory.ExecuteCosmosTx(deployerPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{instantiateMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(instantiateRes.IsOK(), "instantiate should have succeeded: %s", instantiateRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	contractAddr, err := getContractAddrFromEvents(instantiateRes.Events)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(contractAddr)

	// the contract should hold the funds sent on instantiation
	contractBalance, err := suite.grpcHandler.GetBalanceFromBank(sdk.MustAccAddressFromBech32(contractAddr), denom)
	suite.Require().NoError(err)
	suite.Require().Equal(contractFunds.AmountOf(denom), contractBalance.Balance.Amount)

	// 3. smart query the contract's state
	wasmClient := suite.network.GetWasmClient()
	queryRes, err := wasmClient.SmartContractState(suite.network.GetContext(), &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: []byte(`{"verifier":{}}`),
	})
	suite.Require().NoError(err)
	suite.Require().JSONEq(fmt.Sprintf(`{"verifier":"%s"}`, deployerAddr.String()), string(queryRes.Data))

	// a non-verifier account cannot release the funds: the contract rejects it
	// during gas estimation, so the tx never even makes it to broadcast
	beneficiaryPriv := suite.keyring.GetPrivKey(1)
	unauthorizedExecuteMsg := &wasmtypes.MsgExecuteContract{
		Sender:   beneficiaryAddr.String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(`{"release":{}}`),
	}
	_, err = suite.factory.ExecuteCosmosTx(beneficiaryPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{unauthorizedExecuteMsg},
	})
	suite.Require().Error(err, "release from a non-verifier account should have failed")
	suite.Require().Contains(err.Error(), "Unauthorized")

	beneficiaryBalanceBefore, err := suite.grpcHandler.GetBalanceFromBank(beneficiaryAddr, denom)
	suite.Require().NoError(err)

	// 4. the verifier executes "release", paying out the contract's balance to the beneficiary
	releaseMsg := &wasmtypes.MsgExecuteContract{
		Sender:   deployerAddr.String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(`{"release":{}}`),
	}
	releaseRes, err := suite.factory.ExecuteCosmosTx(deployerPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{releaseMsg},
	})
	suite.Require().NoError(err)
	suite.Require().True(releaseRes.IsOK(), "release from the verifier should have succeeded: %s", releaseRes.GetLog())
	suite.Require().NoError(suite.network.NextBlock())

	beneficiaryBalanceAfter, err := suite.grpcHandler.GetBalanceFromBank(beneficiaryAddr, denom)
	suite.Require().NoError(err)
	suite.Require().Equal(contractFunds.AmountOf(denom), beneficiaryBalanceAfter.Balance.Amount.Sub(beneficiaryBalanceBefore.Balance.Amount))

	contractBalanceAfter, err := suite.grpcHandler.GetBalanceFromBank(sdk.MustAccAddressFromBech32(contractAddr), denom)
	suite.Require().NoError(err)
	suite.Require().True(contractBalanceAfter.Balance.Amount.IsZero(), "contract balance should be drained after release")
}

// getCodeIDFromEvents extracts the code ID assigned to an uploaded contract
// from a MsgStoreCode tx's events.
func getCodeIDFromEvents(events []abcitypes.Event) (uint64, error) {
	for _, event := range events {
		if event.Type != wasmtypes.EventTypeStoreCode {
			continue
		}
		for _, attr := range event.Attributes {
			if attr.Key == wasmtypes.AttributeKeyCodeID {
				var codeID uint64
				if _, err := fmt.Sscanf(attr.Value, "%d", &codeID); err != nil {
					return 0, err
				}
				return codeID, nil
			}
		}
	}
	return 0, errors.New("store_code event not found")
}

// getContractAddrFromEvents extracts the contract's address from a
// MsgInstantiateContract tx's events.
func getContractAddrFromEvents(events []abcitypes.Event) (string, error) {
	for _, event := range events {
		if event.Type != wasmtypes.EventTypeInstantiate {
			continue
		}
		for _, attr := range event.Attributes {
			if attr.Key == wasmtypes.AttributeKeyContractAddr {
				return attr.Value, nil
			}
		}
	}
	return "", errors.New("instantiate event not found")
}
