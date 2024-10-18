package app

import (
	"encoding/json"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	evmtypes "github.com/evmos/os/x/evm/types"
	feemarkettypes "github.com/evmos/os/x/feemarket/types"

	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	"github.com/realiotech/realio-network/cmd/config"
	"github.com/realiotech/realio-network/types"
	bridgetypes "github.com/realiotech/realio-network/x/bridge/types"
	minttypes "github.com/realiotech/realio-network/x/mint/types"
)

func init() {
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)
	config.SetBip44CoinType(cfg)
}

// DefaultTestingAppInit defines the IBC application used for testing
var DefaultTestingAppInit func() (ibctesting.TestingApp, map[string]json.RawMessage) = SetupTestingApp

// DefaultConsensusParams defines the default Tendermint consensus params used in
// Evmos testing.
var (
	DefaultConsensusParams = &tmproto.ConsensusParams{
		Block: &tmproto.BlockParams{
			MaxBytes: 200000,
			MaxGas:   -1, // no limit
		},
		Evidence: &tmproto.EvidenceParams{
			MaxAgeNumBlocks: 302400,
			MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
			MaxBytes:        10000,
		},
		Validator: &tmproto.ValidatorParams{
			PubKeyTypes: []string{
				tmtypes.ABCIPubKeyTypeEd25519,
			},
		},
	}
	MultiStakingCoinA = multistakingtypes.MultiStakingCoin{
		Denom:      "ario",
		Amount:     math.NewIntFromUint64(1000000000000000000),
		BondWeight: math.LegacyMustNewDecFromStr("1.23"),
	}
	MultiStakingCoinB = multistakingtypes.MultiStakingCoin{
		Denom:      "arst",
		Amount:     math.NewIntFromUint64(1000000000000000000),
		BondWeight: math.LegacyMustNewDecFromStr("0.12"),
	}
)

func init() {
	feemarkettypes.DefaultMinGasPrice = math.LegacyZeroDec()
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)
	config.SetBip44CoinType(cfg)
}

// Setup initializes a new App. A Nop logger is set in App.
func Setup(
	isCheckTx bool,
	feemarketGenesis *feemarkettypes.GenesisState,
	numberVals int,
) *RealioNetwork {
	encCdc := MakeEncodingConfig()

	valSet := GenValSet(numberVals)

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(types.BaseDenom, math.NewInt(100000000000000))),
	}

	db := dbm.NewMemDB()
	opt := baseapp.SetChainID(types.MainnetChainID)
	app := New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, encCdc, simtestutil.EmptyAppOptions{}, opt)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewDefaultGenesisState(encCdc.Codec)

		genesisState = GenesisStateWithValSet(app, genesisState, valSet, []authtypes.GenesisAccount{acc}, balance)

		// Verify feeMarket genesis
		if feemarketGenesis != nil {
			if err := feemarketGenesis.Validate(); err != nil {
				panic(err)
			}
			genesisState[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
		}

		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		if _, err = app.InitChain(
			&abci.RequestInitChain{
				ChainId:         types.MainnetChainID,
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		); err != nil {
			panic(err)
		}

		_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{}})
		if err != nil {
			panic(err)
		}
	}

	return app
}

func GenesisStateWithValSet(app *RealioNetwork, genesisState simapp.GenesisState,
	valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) simapp.GenesisState {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	// set multi staking genesis state
	msCoinAInfo := multistakingtypes.MultiStakingCoinInfo{
		Denom:      MultiStakingCoinA.Denom,
		BondWeight: MultiStakingCoinA.BondWeight,
	}
	msCoinBInfo := multistakingtypes.MultiStakingCoinInfo{
		Denom:      MultiStakingCoinB.Denom,
		BondWeight: MultiStakingCoinB.BondWeight,
	}
	msCoinInfos := []multistakingtypes.MultiStakingCoinInfo{msCoinAInfo, msCoinBInfo}
	validatorMsCoins := make([]multistakingtypes.ValidatorMultiStakingCoin, 0, len(valSet.Validators))
	locks := make([]multistakingtypes.MultiStakingLock, 0, len(valSet.Validators))
	lockCoins := sdk.NewCoins()

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))
	bondCoins := sdk.NewCoins()

	for i, val := range valSet.Validators {
		valMsCoin := MultiStakingCoinA
		if i%2 == 1 {
			valMsCoin = MultiStakingCoinB
		}
		validatorMsCoins = append(validatorMsCoins, multistakingtypes.ValidatorMultiStakingCoin{
			ValAddr:   sdk.ValAddress(val.Address).String(),
			CoinDenom: valMsCoin.Denom,
		})
		lockID := multistakingtypes.MultiStakingLockID(genAccs[0].GetAddress().String(), sdk.ValAddress(val.Address).String())
		lockRecord := multistakingtypes.NewMultiStakingLock(lockID, valMsCoin)
		locks = append(locks, lockRecord)
		lockCoins = lockCoins.Add(valMsCoin.ToCoin())

		pk, _ := cryptocodec.FromTmPubKeyInterface(val.PubKey)

		comm := stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
		pkAny, _ := codectypes.NewAnyWithValue(pk)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            valMsCoin.BondValue(),
			DelegatorShares:   math.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        comm,
			MinSelfDelegation: math.ZeroInt(),
		}

		valAddr := sdk.ValAddress(pk.Address())
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(sdk.AccAddress(valAddr).String(), valAddr.String(), math.LegacyOneDec()))

		bondCoins = bondCoins.Add(sdk.NewCoin(sdk.DefaultBondDenom, valMsCoin.BondValue()))
	}
	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	multistakingGenesis := multistakingtypes.GenesisState{
		MultiStakingLocks:          locks,
		MultiStakingUnlocks:        []multistakingtypes.MultiStakingUnlock{},
		MultiStakingCoinInfo:       msCoinInfos,
		ValidatorMultiStakingCoins: validatorMsCoins,
		StakingGenesisState:        *stakingGenesis,
	}
	genesisState[multistakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(&multistakingGenesis)

	// set mint genesis
	mintGenesis := minttypes.DefaultGenesisState()
	genesisState[minttypes.ModuleName] = app.AppCodec().MustMarshalJSON(mintGenesis)

	// set evm genesis
	evmGenesis := evmtypes.DefaultGenesisState()
	evmGenesis.Params.EvmDenom = types.BaseDenom
	genesisState[evmtypes.ModuleName] = app.AppCodec().MustMarshalJSON(evmGenesis)

	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   bondCoins,
	})
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(multistakingtypes.ModuleName).String(),
		Coins:   lockCoins,
	})

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	bridgeGenesis := bridgetypes.DefaultGenesis()
	bridgeGenesis.RatelimitEpochInfo.CurrentEpochStartHeight = 5
	bridgeGenesis.RegisteredCoins = []sdk.Coin{
		sdk.NewCoin(MultiStakingCoinA.Denom, math.NewInt(1000000000)),
		sdk.NewCoin(MultiStakingCoinB.Denom, math.NewInt(1000000000)),
	}
	bridgeGenesis.Params.Authority = genAccs[0].GetAddress().String()
	genesisState[bridgetypes.ModuleName] = app.AppCodec().MustMarshalJSON(bridgeGenesis)

	return genesisState
}

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) simapp.GenesisState {
	return ModuleBasics.DefaultGenesis(cdc)
}

// SetupTestingApp initializes the IBC-go testing application
func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	cfg := MakeEncodingConfig()
	app := New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, cfg, simtestutil.EmptyAppOptions{})
	return app, NewDefaultGenesisState(cfg.Codec)
}

func GenValSet(nums int) *tmtypes.ValidatorSet {
	vals := []*tmtypes.Validator{}

	for i := 0; i < nums; i++ {
		privVal := mock.NewPV()
		pubKey, _ := privVal.GetPubKey()
		vals = append(vals, tmtypes.NewValidator(pubKey, 1))
	}
	valSet := tmtypes.NewValidatorSet(vals)

	return valSet
}
