package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/realiotech/realio-network/app/ante"
	"github.com/realiotech/realio-network/client/docs"
	"github.com/realiotech/realio-network/crypto/ethsecp256k1"
	"github.com/realiotech/realio-network/crypto/ossecp256k1"

	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/runtime"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/gogoproto/proto"

	// staking
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// multi-staking
	multistaking "github.com/realio-tech/multi-staking-module/x/multi-staking"
	multistakingkeeper "github.com/realio-tech/multi-staking-module/x/multi-staking/keeper"
	multistakingtypes "github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	evmosante "github.com/cosmos/evm/ante"
	evmosanteevm "github.com/cosmos/evm/ante/evm"
	evmsecp256k1 "github.com/cosmos/evm/crypto/ethsecp256k1"
	srvflags "github.com/cosmos/evm/server/flags"
	ostypes "github.com/cosmos/evm/types"
	evmosvm "github.com/ethereum/go-ethereum/core/vm"

	"github.com/cosmos/evm/x/feemarket"
	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	"github.com/cosmos/evm/x/vm"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	simappparams "cosmossdk.io/simapp/params"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v10/modules/core"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"

	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/spf13/cast"

	"github.com/realiotech/realio-network/x/mint"
	mintkeeper "github.com/realiotech/realio-network/x/mint/keeper"
	minttypes "github.com/realiotech/realio-network/x/mint/types"

	assetmodule "github.com/realiotech/realio-network/x/asset"
	assetmodulekeeper "github.com/realiotech/realio-network/x/asset/keeper"
	assetmoduletypes "github.com/realiotech/realio-network/x/asset/types"

	bridgemodule "github.com/realiotech/realio-network/x/bridge"
	bridgemodulekeeper "github.com/realiotech/realio-network/x/bridge/keeper"
	bridgemoduletypes "github.com/realiotech/realio-network/x/bridge/types"

	realionetworktypes "github.com/realiotech/realio-network/types"
	// this line is used by starport scaffolding # stargate/app/moduleImport

	// Force-load the tracer engines to trigger registration due to Go-Ethereum v1.10.15 changes
	_ "github.com/ethereum/go-ethereum/eth/tracers/js"
	_ "github.com/ethereum/go-ethereum/eth/tracers/native"
)

const (
	Name                    = "realio-network"
	bech32PrecompileBaseGas = 6_000
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{
			GenTxValidator: genutiltypes.DefaultMessageValidator,
		},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		multistaking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			[]govclient.ProposalHandler{
				paramsclient.ProposalHandler,
				multistaking.AddMultiStakingProposalHandler,
				multistaking.UpdateBondWeightProposalHandler,
			},
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctm.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		vm.AppModuleBasic{},
		feemarket.AppModuleBasic{},
		assetmodule.AppModuleBasic{},
		bridgemodule.AppModuleBasic{},
		consensus.AppModuleBasic{},
		// this line is used by starport scaffolding # stargate/app/moduleBasic
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter, authtypes.Burner},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		evmtypes.ModuleName:            {authtypes.Minter, authtypes.Burner}, // used for secure addition and subtraction of balance using module account
		assetmoduletypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		bridgemoduletypes.ModuleName:   {authtypes.Minter, authtypes.Burner},
		multistakingtypes.ModuleName:   {authtypes.Minter, authtypes.Burner},
		// this line is used by starport scaffolding # stargate/app/maccPerms
	}
)

var (
	_ runtime.AppI            = (*RealioNetwork)(nil)
	_ servertypes.Application = (*RealioNetwork)(nil)
	_ ibctesting.TestingApp   = (*RealioNetwork)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)

	// manually update the power reduction by replacing micro (u) -> atto (a) rio
	sdk.DefaultPowerReduction = realionetworktypes.PowerReduction
}

// RealioNetwork extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type RealioNetwork struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry
	txConfig          client.TxConfig

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	MultiStakingKeeper    multistakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper        evidencekeeper.Keeper
	TransferKeeper        ibctransferkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	// Ethermint keepers
	EvmKeeper       *evmkeeper.Keeper
	FeeMarketKeeper feemarketkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper

	// Realio Network keepers
	AssetKeeper  assetmodulekeeper.Keeper
	BridgeKeeper bridgemodulekeeper.Keeper

	// mm is the module manager
	mm *module.Manager

	// sm is the simulation manager
	sm *module.SimulationManager

	// the configurator
	configurator module.Configurator
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig simappparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	evmAppOptions EVMOptionsFn,
	baseAppOptions ...func(*baseapp.BaseApp),
) *RealioNetwork {
	appCodec := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetTxEncoder(encodingConfig.TxConfig.TxEncoder())
	bApp.SetInterfaceRegistry(interfaceRegistry)

	// initialize the Cosmos EVM application configuration
	if err := evmAppOptions(bApp.ChainID()); err != nil {
		panic(err)
	}

	keys := storetypes.NewKVStoreKeys(
		// sdk keys
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, capabilitytypes.StoreKey, authzkeeper.StoreKey,
		crisistypes.StoreKey, consensusparamtypes.StoreKey,
		// ibc keys
		ibcexported.StoreKey, ibctransfertypes.StoreKey,
		// realio network keys
		assetmoduletypes.StoreKey, bridgemoduletypes.StoreKey,
		// ethermint keys
		evmtypes.StoreKey, feemarkettypes.StoreKey,
		// multi-staking keys
		multistakingtypes.StoreKey,
		// this line is used by starport scaffolding # stargate/app/storeKey
	)

	// register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	// Add the EVM transient store key
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey, evmtypes.TransientKey, feemarkettypes.TransientKey)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &RealioNetwork{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		txConfig:          encodingConfig.TxConfig,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]), authtypes.NewModuleAddress(govtypes.ModuleName).String(), runtime.EventService{})
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)

	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	app.CapabilityKeeper.Seal()
	// this line is used by starport scaffolding # stargate/app/scopedKeeper

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(appCodec, runtime.NewKVStoreService(keys[authtypes.StoreKey]), authtypes.ProtoBaseAccount, maccPerms, authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()), sdk.GetConfig().GetBech32AccountAddrPrefix(), authtypes.NewModuleAddress(govtypes.ModuleName).String())

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		app.ModuleAccountAddrs(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
	// multi-staking keeper
	app.MultiStakingKeeper = *multistakingkeeper.NewKeeper(
		appCodec,
		app.AccountKeeper,
		app.StakingKeeper,
		app.BankKeeper,
		keys[multistakingtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		app.StakingKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AccountKeeper.AddressCodec(),
	)
	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feegrant.StoreKey]),
		app.AccountKeeper,
	)

	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	// Ethermint keepers
	tracer := cast.ToString(appOpts.Get(srvflags.EVMTracer))

	app.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		appCodec, authtypes.NewModuleAddress(govtypes.ModuleName),
		keys[feemarkettypes.StoreKey], tkeys[feemarkettypes.TransientKey],
	)

	app.EvmKeeper = evmkeeper.NewKeeper(
		appCodec, keys[evmtypes.StoreKey], tkeys[evmtypes.TransientKey], authtypes.NewModuleAddress(govtypes.ModuleName),
		app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.FeeMarketKeeper, MockErc20Keeper{},
		tracer,
	)

	app.EvmKeeper.WithStaticPrecompiles(evmosvm.PrecompiledContractsBerlin)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	// realio keeper
	app.AssetKeeper = *assetmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[assetmoduletypes.StoreKey]),
		app.BankKeeper,
		app.AccountKeeper,
		app.ModuleAccountAddrs(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.BridgeKeeper = bridgemodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[bridgemoduletypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Add transfer restriction
	app.BankKeeper.AppendSendRestriction(app.AssetKeeper.AssetSendRestriction)

	// IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[ibcexported.StoreKey]),
		app.GetSubspace(ibcexported.ModuleName),
		app.UpgradeKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(keys[evidencetypes.StoreKey]), app.StakingKeeper, app.SlashingKeeper, app.AccountKeeper.AddressCodec(), runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	// register the proposal types
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(multistakingtypes.RouterKey, multistaking.NewMultiStakingProposalHandler(&app.MultiStakingKeeper))

	govConfig := govtypes.DefaultConfig()
	/*
		Example of setting gov params:
		govConfig.MaxMetadataLen = 10000
	*/
	app.GovKeeper = *govkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(keys[govtypes.StoreKey]), app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, app.DistrKeeper, app.MsgServiceRouter(), govConfig, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	// Set legacy router for backwards compatibility with gov v1beta1
	app.GovKeeper.SetLegacyRouter(govRouter)

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(keys[ibctransfertypes.StoreKey]), app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper, // ISC4 Wrapper: fee IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		app.AccountKeeper, app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	transferModule := transfer.NewAppModule(app.TransferKeeper)

	// create IBC module from top to bottom of stack
	transferStack := transfer.NewIBCModule(app.TransferKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	clientKeeper := app.IBCKeeper.ClientKeeper
	storeProvider := clientKeeper.GetStoreProvider()

	tmLightClientModule := ibctm.NewLightClientModule(appCodec, storeProvider)
	clientKeeper.AddRoute(ibctm.ModuleName, &tmLightClientModule)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		// sdk modules
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		multistaking.NewAppModule(appCodec, app.MultiStakingKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),

		// ibc
		ibc.NewAppModule(app.IBCKeeper),
		transferModule,
		ibctm.NewAppModule(tmLightClientModule),

		// ethermint
		vm.NewAppModule(app.EvmKeeper, app.AccountKeeper),
		feemarket.NewAppModule(app.FeeMarketKeeper),

		// realio network
		assetmodule.NewAppModule(appCodec, app.AssetKeeper, app.BankKeeper, app.GetSubspace(assetmoduletypes.ModuleName)),
		bridgemodule.NewAppModule(appCodec, app.BridgeKeeper),
	)

	// NOTE: upgrade module is required to be prioritized
	app.mm.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: upgrade module must go first to handle software upgrades.
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
	app.mm.SetOrderBeginBlockers(
		capabilitytypes.ModuleName,
		feemarkettypes.ModuleName,
		evmtypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		multistakingtypes.ModuleName,
		ibcexported.ModuleName,
		// no-op modules
		ibctransfertypes.ModuleName,
		authtypes.ModuleName,
		minttypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		assetmoduletypes.ModuleName,
		bridgemoduletypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/beginBlockers
	)

	// NOTE: fee market module must go last in order to retrieve the block gas used.
	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		multistakingtypes.ModuleName,
		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		// no-op modules
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		vestingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		// realio modules
		assetmoduletypes.ModuleName,
		bridgemoduletypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		// sdk modules
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		multistakingtypes.ModuleName,
		vestingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		ibcexported.ModuleName,
		// ethermint
		// evm module denomination is used by the feemarket module, in AnteHandle
		evmtypes.ModuleName,
		// NOTE: feemarket need to be initialized before genutil module:
		// gentx transactions use MinGasPriceDecorator.AnteHandle
		feemarkettypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		// realio modules
		assetmoduletypes.ModuleName,
		bridgemoduletypes.ModuleName,
		// NOTE: crisis module must go at the end to check for invariants on each module
		crisistypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	app.mm.SetOrderExportGenesis(
		// sdk modules
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		multistakingtypes.ModuleName,
		vestingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		ibcexported.ModuleName,
		// ethermint
		// evm module denomination is used by the feemarket module, in AnteHandle
		evmtypes.ModuleName,
		// NOTE: feemarket need to be initialized before genutil module:
		// gentx transactions use MinGasPriceDecorator.AnteHandle
		feemarkettypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		// realio modules
		assetmoduletypes.ModuleName,
		bridgemoduletypes.ModuleName,
		// NOTE: crisis module must go at the end to check for invariants on each module
		crisistypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err := app.mm.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	app.setupUpgradeHandlers()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	// create the simulation manager and define the order of the modules for deterministic simulations
	// NOTE: this is not required apps that don't use the simulator for fuzz testing transactions
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, app.GetSubspace(minttypes.ModuleName)),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		params.NewAppModule(app.ParamsKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.IBCKeeper),
		transferModule,
		vm.NewAppModule(app.EvmKeeper, app.AccountKeeper),
		feemarket.NewAppModule(app.FeeMarketKeeper),
		assetmodule.NewAppModule(appCodec, app.AssetKeeper, app.BankKeeper, app.GetSubspace(assetmoduletypes.ModuleName)),
		bridgemodule.NewAppModule(appCodec, app.BridgeKeeper),
		// this line is used by starport scaffolding # stargate/app/appModule
	)

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)
	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	maxGasWanted := cast.ToUint64(appOpts.Get(srvflags.EVMMaxTxGasWanted))
	options := ante.HandlerOptions{
		AccountKeeper:          app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		SignModeHandler:        encodingConfig.TxConfig.SignModeHandler(),
		FeegrantKeeper:         app.FeeGrantKeeper,
		SigGasConsumer:         RealioSigVerificationGasConsumer,
		IBCKeeper:              app.IBCKeeper,
		EvmKeeper:              app.EvmKeeper,
		FeeMarketKeeper:        app.FeeMarketKeeper,
		MaxTxGasWanted:         maxGasWanted,
		ExtensionOptionChecker: ostypes.HasDynamicFeeExtensionOption,
		TxFeeChecker:           evmosanteevm.NewDynamicFeeChecker(app.FeeMarketKeeper),
	}

	if err := options.Validate(); err != nil {
		panic(err)
	}

	app.SetAnteHandler(ante.NewAnteHandler(options))
	app.setPostHandler()

	// At startup, after all modules have been registered, check that all prot
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}

	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		fmt.Fprintln(os.Stderr, err.Error())
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	// this line is used by starport scaffolding # stargate/app/beforeInitReturn

	return app
}

func (app *RealioNetwork) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// Name returns the name of the App
func (app *RealioNetwork) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *RealioNetwork) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.mm.PreBlock(ctx)
}

// BeginBlocker runs the Tendermint ABCI BeginBlock logic. It executes state changes at the beginning
// of the new block for every registered module. If there is a registered fork at the current height,
// BeginBlocker will schedule the upgrade plan and perform the state migration (if any).
func (app *RealioNetwork) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	// Perform any scheduled forks before executing the modules logic
	app.ScheduleForkUpgrade(ctx)
	return app.mm.BeginBlock(ctx)
}

// EndBlocker updates every end block
func (app *RealioNetwork) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *RealioNetwork) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState simapp.GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap()); err != nil {
		panic(err)
	}

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *RealioNetwork) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *RealioNetwork) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *RealioNetwork) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *RealioNetwork) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *RealioNetwork) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *RealioNetwork) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *RealioNetwork) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *RealioNetwork) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *RealioNetwork) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *RealioNetwork) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *RealioNetwork) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *RealioNetwork) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// IBC Go TestingApp functions

// GetBaseApp implements the TestingApp interface.
func (app *RealioNetwork) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

type StakingKeeper interface {
	GetHistoricalInfo(ctx context.Context, height int64) (stakingtypes.HistoricalInfo, error)
}

// GetStakingKeeper implements the TestingApp interface.
func (app *RealioNetwork) GetStakingKeeper() StakingKeeper {
	return app.StakingKeeper
}

// GetStakingKeeper implements the TestingApp interface.
func (app *RealioNetwork) GetStakingKeeperSDK() stakingkeeper.Keeper {
	return *app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *RealioNetwork) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *RealioNetwork) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *RealioNetwork) GetTxConfig() client.TxConfig {
	return app.txConfig
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *RealioNetwork) DefaultGenesis() map[string]json.RawMessage {
	return ModuleBasics.DefaultGenesis(app.appCodec)
}

// AutoCliOpts returns the autocli options for the app.
func (app *RealioNetwork) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	modules[stakingtypes.ModuleName] = staking.NewAppModule(app.appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName))

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.mm.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router) {
	root, err := fs.Sub(docs.SwaggerUI, "swagger-ui")
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(http.FS(root))
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

func (app *RealioNetwork) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// SimulationManager implements the SimulationApp interface
func (app *RealioNetwork) SimulationManager() *module.SimulationManager {
	return app.sm
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName).WithKeyTable(authtypes.ParamKeyTable())       //nolint: staticcheck
	paramsKeeper.Subspace(banktypes.ModuleName).WithKeyTable(banktypes.ParamKeyTable())       //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(stakingtypes.ModuleName).WithKeyTable(stakingtypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(minttypes.ModuleName).WithKeyTable(minttypes.ParamKeyTable())
	paramsKeeper.Subspace(distrtypes.ModuleName).WithKeyTable(distrtypes.ParamKeyTable())       //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(slashingtypes.ModuleName).WithKeyTable(slashingtypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())              //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(crisistypes.ModuleName).WithKeyTable(crisistypes.ParamKeyTable())     //nolint: staticcheck // SA1019
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	// realio network subspaces
	paramsKeeper.Subspace(assetmoduletypes.ModuleName).WithKeyTable(assetmoduletypes.ParamKeyTable())
	// ethermint subspaces
	paramsKeeper.Subspace(evmtypes.ModuleName).WithKeyTable(evmtypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(feemarkettypes.ModuleName).WithKeyTable(feemarkettypes.ParamKeyTable())
	// this line is used by starport scaffolding # stargate/app/paramSubspace

	return paramsKeeper
}

func RealioSigVerificationGasConsumer(
	meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params,
) error {
	pubkey := sig.PubKey
	switch pubkey := pubkey.(type) {
	case *evmsecp256k1.PubKey:
		// Ethereum keys
		meter.ConsumeGas(evmosante.Secp256k1VerifyCost, "ante verify: eth_secp256k1")
		return nil
	case *ossecp256k1.PubKey:
		// Ethereum keys
		meter.ConsumeGas(evmosante.Secp256k1VerifyCost, "ante verify: eth_secp256k1")
		return nil
	case *ethsecp256k1.PubKey:
		// Ethereum keys
		meter.ConsumeGas(evmosante.Secp256k1VerifyCost, "ante verify: eth_secp256k1")
		return nil
	case *ed25519.PubKey:
		// Validator keys
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return errorsmod.Wrap(errortypes.ErrInvalidPubKey, "ED25519 public keys are unsupported")
	case multisig.PubKey:
		// Multisig keys
		multisignature, ok := sig.Data.(*signing.MultiSignatureData)
		if !ok {
			return fmt.Errorf("expected %T, got, %T", &signing.MultiSignatureData{}, sig.Data)
		}
		return evmosante.ConsumeMultisignatureVerificationGas(meter, multisignature, pubkey, params, sig.Sequence)
	default:
		return authante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
}
