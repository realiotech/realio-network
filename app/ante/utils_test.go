package ante_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
	client "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	protov2 "google.golang.org/protobuf/proto"

	cryptocodec "github.com/cosmos/evm/crypto/codec"
	osecp256k1 "github.com/cosmos/evm/crypto/ethsecp256k1"
	ethcryptocodec "github.com/realiotech/realio-network/crypto/codec"

	"github.com/cosmos/evm/encoding"
	"github.com/cosmos/evm/ethereum/eip712"
	tests "github.com/cosmos/evm/testutil/tx"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"

	"github.com/realiotech/realio-network/app"
	realionetworktypes "github.com/realiotech/realio-network/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	enccodec "github.com/cosmos/evm/encoding/codec"
)

var (
	s *AnteTestSuite
	_ sdk.AnteHandler = (&MockAnteHandler{}).AnteHandle
)

type AnteTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	clientCtx client.Context
	app       *app.RealioNetwork
	denom     string
}

type MockAnteHandler struct {
	WasCalled bool
	CalledCtx sdk.Context
}

func (mah *MockAnteHandler) AnteHandle(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	mah.WasCalled = true
	mah.CalledCtx = ctx
	return ctx, nil
}

func (suite *AnteTestSuite) SetupTest() {
	t := suite.T()
	privCons, err := osecp256k1.GenerateKey()
	require.NoError(t, err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	isCheckTx := false
	suite.app = app.Setup(isCheckTx, feemarkettypes.DefaultGenesisState(), 1)
	suite.Require().NotNil(suite.app.AppCodec())

	suite.ctx = suite.app.BaseApp.NewContextLegacy(isCheckTx, tmproto.Header{
		Height:          1,
		ChainID:         realionetworktypes.MainnetChainID + "-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	suite.denom = realionetworktypes.BaseDenom
	evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
	evmParams.EvmDenom = suite.denom
	_ = suite.app.EvmKeeper.SetParams(suite.ctx, evmParams)

	suite.clientCtx = client.Context{}.WithTxConfig(suite.app.GetTxConfig())
}

func TestAnteTestSuite(t *testing.T) {
	s = new(AnteTestSuite)
	suite.Run(t, s)
}

// Commit commits and starts a new block with an updated context.
func (suite *AnteTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *AnteTestSuite) CommitAfter(t time.Duration) {
	_, err := suite.app.EndBlocker(suite.ctx)
	suite.Require().NoError(err)
	_, err = suite.app.Commit()
	suite.Require().NoError(err)

	header := suite.ctx.BlockHeader()
	header.Height++
	header.Time = header.Time.Add(t)

	suite.ctx = suite.app.BaseApp.NewContextLegacy(false, header)
	_, err = suite.app.BeginBlocker(suite.ctx)
	suite.Require().NoError(err)
}

func (suite *AnteTestSuite) CreateTestTxBuilder(gasPrice sdkmath.Int, denom string, msgs ...sdk.Msg) client.TxBuilder {
	encodingConfig := encoding.MakeConfig(app.MainnetChainID)
	gasLimit := uint64(1000000)

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(gasLimit)
	fees := &sdk.Coins{{Denom: denom, Amount: gasPrice.Mul(sdkmath.NewIntFromUint64(gasLimit))}}
	txBuilder.SetFeeAmount(*fees)
	err := txBuilder.SetMsgs(msgs...)
	suite.Require().NoError(err)
	return txBuilder
}

func (suite *AnteTestSuite) CreateEthTestTxBuilder(msgEthereumTx *evmtypes.MsgEthereumTx) client.TxBuilder {
	encodingConfig := encoding.MakeConfig(app.MainnetChainID)
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	suite.Require().NoError(err)

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)
	builder.SetExtensionOptions(option)

	err = txBuilder.SetMsgs(msgEthereumTx)
	suite.Require().NoError(err)

	fees := sdk.Coins{{Denom: s.denom, Amount: sdkmath.NewInt(100_000)}}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msgEthereumTx.GetGas())

	return txBuilder
}

func (suite *AnteTestSuite) BuildTestEthTx(
	from common.Address,
	to common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) *evmtypes.MsgEthereumTx {
	chainID := evmtypes.GetEthChainConfig().ChainID
	nonce := suite.app.EvmKeeper.GetNonce(
		suite.ctx,
		common.BytesToAddress(from.Bytes()),
	)
	data := make([]byte, 0)
	gasLimit := uint64(100000)

	args := evmtypes.EvmTxArgs{
		Nonce:     nonce,
		GasLimit:  gasLimit,
		Input:     data,
		GasFeeCap: gasFeeCap,
		GasPrice:  gasPrice,
		ChainID:   chainID,
		GasTipCap: gasTipCap,
		Amount:    nil,
		To:        &to,
		Accesses:  accesses,
	}
	msgEthereumTx := evmtypes.NewTx(&args)
	msgEthereumTx.From = from.Bytes()
	return msgEthereumTx
}

var _ sdk.Tx = &invalidTx{}

type invalidTx struct{}

func (invalidTx) GetMsgs() []sdk.Msg { return []sdk.Msg{nil} }
func (invalidTx) GetMsgsV2() ([]protov2.Message, error) {
	return []protov2.Message{nil}, nil
}
func (invalidTx) ValidateBasic() error { return nil }

func newMsgGrant(granter sdk.AccAddress, grantee sdk.AccAddress, a authz.Authorization, expiration *time.Time) *authz.MsgGrant {
	msg, err := authz.NewMsgGrant(granter, grantee, a, expiration)
	if err != nil {
		panic(err)
	}
	return msg
}

func newMsgExec(grantee sdk.AccAddress, msgs []sdk.Msg) *authz.MsgExec {
	msg := authz.NewMsgExec(grantee, msgs)
	return &msg
}

func createNestedMsgExec(a sdk.AccAddress, nestedLvl int, lastLvlMsgs []sdk.Msg) *authz.MsgExec {
	msgs := make([]*authz.MsgExec, nestedLvl)
	for i := range msgs {
		if i == 0 {
			msgs[i] = newMsgExec(a, lastLvlMsgs)
			continue
		}
		msgs[i] = newMsgExec(a, []sdk.Msg{msgs[i-1]})
	}
	return msgs[nestedLvl-1]
}

func generatePrivKeyAddressPairs(accCount int) ([]*osecp256k1.PrivKey, []sdk.AccAddress, error) {
	var (
		err           error
		testPrivKeys  = make([]*osecp256k1.PrivKey, accCount)
		testAddresses = make([]sdk.AccAddress, accCount)
	)

	for i := range testPrivKeys {
		testPrivKeys[i], err = osecp256k1.GenerateKey()
		if err != nil {
			return nil, nil, err
		}
		testAddresses[i] = testPrivKeys[i].PubKey().Address().Bytes()
	}
	return testPrivKeys, testAddresses, nil
}

func createTx(priv *osecp256k1.PrivKey, msgs ...sdk.Msg) (sdk.Tx, error) {
	encodingConfig := encoding.MakeConfig(app.MainnetChainID)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	defaultSignMode, err := authsigning.APISignModeToInternal(encodingConfig.TxConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, err
	}
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  defaultSignMode,
			Signature: nil,
		},
		Sequence: 0,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		ChainID:       realionetworktypes.MainnetChainID + "-1",
		AccountNumber: 0,
		Sequence:      0,
	}
	sigV2, err = tx.SignWithPrivKey(
		context.TODO(), defaultSignMode, signerData,
		txBuilder, priv, encodingConfig.TxConfig, 0,
	)
	if err != nil {
		return nil, err
	}

	sigsV2 = []signing.SignatureV2{sigV2}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return txBuilder.GetTx(), nil
}

func createEIP712CosmosTx(
	from sdk.AccAddress, priv cryptotypes.PrivKey, msgs []sdk.Msg,
) (sdk.Tx, error) {
	var err error

	encodingConfig := app.MakeEncodingConfig(app.MainnetChainID)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	// GenerateTypedData TypedData
	registry := codectypes.NewInterfaceRegistry()
	enccodec.RegisterInterfaces(registry)
	ethermintCodec := codec.NewProtoCodec(registry)
	cryptocodec.RegisterInterfaces(registry)
	ethcryptocodec.RegisterInterfaces(registry)

	coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(20))
	amount := sdk.NewCoins(coinAmount)
	gas := uint64(200000)

	fee := legacytx.NewStdFee(gas, amount) //nolint

	data := legacytx.StdSignBytes(realionetworktypes.MainnetChainID+"-1", 0, 0, 0, fee, msgs, "")

	typedData, err := eip712.LegacyWrapTxToTypedData(ethermintCodec, app.MainnetChainID, msgs[0], data, &eip712.FeeDelegationOptions{
		FeePayer: from,
	})
	if err != nil {
		return nil, err
	}

	sigHash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	// Sign typedData
	keyringSigner := tests.NewSigner(priv)
	signature, pubKey, err := keyringSigner.SignByAddress(from, sigHash, signing.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		return nil, err
	}
	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper

	builder, _ := txBuilder.(authtx.ExtensionOptionsTxBuilder)

	builder.SetFeeAmount(amount)
	builder.SetGasLimit(gas)

	sigsV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		},
		Sequence: 0,
	}

	if err = builder.SetSignatures(sigsV2); err != nil {
		return nil, err
	}

	if err = builder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	return builder.GetTx(), err
}
