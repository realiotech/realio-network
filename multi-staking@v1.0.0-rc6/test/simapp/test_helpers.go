package simapp

import (
	"encoding/json"
	"time"

	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DefaultConsensusParams defines the default Tendermint consensus params used in
// SimApp testing.
var (
	DefaultConsensusParams = &abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxBytes: 200000,
			MaxGas:   2000000,
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
		Amount:     sdk.NewIntFromUint64(100000000),
		BondWeight: sdk.MustNewDecFromStr("1.23"),
	}
	MultiStakingCoinB = multistakingtypes.MultiStakingCoin{
		Denom:      "arst",
		Amount:     sdk.NewIntFromUint64(100000000),
		BondWeight: sdk.MustNewDecFromStr("0.12"),
	}
)

// Setup initializes a new SimApp. A Nop logger is set in SimApp.
func Setup(isCheckTx bool) *SimApp {
	valSet := GenValSet()

	app := SetupWithGenesisValSet(valSet)
	return app
}

// SetupWithGenesisValSet initializes a new SimApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func SetupWithGenesisValSet(valSet *tmtypes.ValidatorSet) *SimApp {
	app, genesisState := setup(true, 5)
	genesisState = genesisStateWithValSet(app, genesisState, valSet)

	stateBytes, _ := json.MarshalIndent(genesisState, "", " ")

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// commit genesis changes
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		Height:             app.LastBlockHeight() + 1,
		AppHash:            app.LastCommitID().Hash,
		ValidatorsHash:     valSet.Hash(),
		NextValidatorsHash: valSet.Hash(),
	}})

	return app
}

func setup(withGenesis bool, invCheckPeriod uint) (*SimApp, GenesisState) {
	db := dbm.NewMemDB()
	encCdc := MakeTestEncodingConfig()
	app := NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, invCheckPeriod, encCdc, EmptyAppOptions{})
	if withGenesis {
		return app, NewDefaultGenesisState(encCdc.Codec)
	}
	return app, GenesisState{}
}

func genesisStateWithValSet(app *SimApp, genesisState GenesisState, valSet *tmtypes.ValidatorSet) GenesisState {
	genAcc := GenAcc()
	genAccs := []authtypes.GenesisAccount{genAcc}
	balances := []banktypes.Balance{}

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

	// staking genesis state
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

		lockId := multistakingtypes.MultiStakingLockID(genAcc.GetAddress().String(), sdk.ValAddress(val.Address).String())
		lockRecord := multistakingtypes.NewMultiStakingLock(lockId, valMsCoin)

		locks = append(locks, lockRecord)
		lockCoins = lockCoins.Add(valMsCoin.ToCoin())

		pk, _ := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		pkAny, _ := codectypes.NewAnyWithValue(pk)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            valMsCoin.BondValue(),
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAcc.GetAddress(), val.Address.Bytes(), sdk.OneDec()))

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
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	return genesisState
}

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}

func GenValSet() *tmtypes.ValidatorSet {
	privVal0 := mock.NewPV()
	privVal1 := mock.NewPV()

	pubKey0, _ := privVal0.GetPubKey()
	pubKey1, _ := privVal1.GetPubKey()

	// create validator set with single validator
	val0 := tmtypes.NewValidator(pubKey0, 1)
	val1 := tmtypes.NewValidator(pubKey1, 1)

	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{val0, val1})

	return valSet
}

func GenAcc() authtypes.GenesisAccount {
	senderPrivKey := secp256k1.GenPrivKey()
	return authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
}
