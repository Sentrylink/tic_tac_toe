package app

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"tic_tac_toe/x/tic_tac_toe"
)

const (
	appName = "tic_tac_toe"
)

type App struct {
	*baseapp.BaseApp
	cdc    *codec.Codec
	logger log.Logger

	// Storage keys
	keyMain *sdk.KVStoreKey

	paramsKeeper  params.Keeper
	accountKeeper auth.AccountKeeper
	feeCollectionKeeper auth.FeeCollectionKeeper

	keeper tic_tac_toe.Keeper
}

func NewApp(logger log.Logger, db db.DB) *App {
	cdc := MakeDefaultCodec()

	base := baseapp.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc))

	app := &App{
		BaseApp: base,
		cdc:     cdc,
		logger:  logger,
		keyMain: sdk.NewKVStoreKey("main"),
	}

	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	app.paramsKeeper = params.NewKeeper(
		app.cdc,
		keyParams,
		tkeyParams,
	)

	keyAccount := sdk.NewKVStoreKey("acc")
	// Uses default account struct
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		keyAccount,
		app.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	keyFeeCollection := sdk.NewKVStoreKey("fee_collection")
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(cdc, keyFeeCollection)

	keyTicTacToe := sdk.NewKVStoreKey("tictactoe")
	app.keeper = tic_tac_toe.NewKeeper(cdc, keyTicTacToe, app.accountKeeper)

	app.Router().
		AddRoute("tictactoe", tic_tac_toe.NewHandler(app.keeper))

	app.QueryRouter().
		AddRoute(auth.QuerierRoute, auth.NewQuerier(app.accountKeeper)).
		AddRoute("tictactoe", tic_tac_toe.NewQuerier(app.keeper))

	app.MountStores(
		app.keyMain,
		keyParams,
		tkeyParams,
		keyAccount,
		keyTicTacToe,
		keyFeeCollection,
	)

	app.SetInitChainer(app.initChainer)

	if err := app.LoadLatestVersion(app.keyMain); err != nil {
		common.Exit(err.Error())
	}

	return app
}

func (app *App) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	// The initial state as stored in genesis file
	genesisStateJSON := req.AppStateBytes

	genesisState := new(GenesisState)

	if err := app.cdc.UnmarshalJSON(genesisStateJSON, genesisState); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal genesis state: %s", err))
	}

	if err := auth.ValidateGenesis(genesisState.AuthState); err != nil {
		panic(fmt.Sprintf("Invalid genesis auth state: %s", err))
	}

	auth.InitGenesis(ctx, app.accountKeeper, app.feeCollectionKeeper, genesisState.AuthState)

	// Setting up initial accounts
	accounts := make(map[string]bool)

	for _, initialAccount := range genesisState.Accounts {
		addrStr := initialAccount.Address.String()
		if _, exists := accounts[addrStr]; exists {
			panic(fmt.Sprintf("Duplicate account %s in genesis", addrStr))
		}

		accounts[addrStr] = true

		initialAccount.AccountNumber = app.accountKeeper.GetNextAccountNumber(ctx)
		app.accountKeeper.SetAccount(ctx, initialAccount)
	}

	initResponse := abci.ResponseInitChain{
		Validators: genesisState.Validators,
	}

	return initResponse
}


// Uses go-amino which is a fork of protobuf3
// Here the codec implementation is injected into different modules
func MakeDefaultCodec() *codec.Codec {
	var cdc = codec.New()
	tic_tac_toe.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

type GenesisState struct {
	AuthState     auth.GenesisState         `json:"auth"`
	Accounts      []*auth.BaseAccount       `json:"accounts"`
	Validators []abci.ValidatorUpdate `json:"validators"`
}