package main

import (
	"encoding/json"
	"fmt"
	"github.com/tendermint/tendermint/crypto"
	"io"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	tendermintConfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tm "github.com/tendermint/tendermint/types"
	"tic_tac_toe"
)

const (
	flagOverwrite  = "overwrite"
	DefaultChainID = "ttt-chain"
)

var (
	DefaultNodeHome = os.ExpandEnv("$HOME/.ttt")
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeDefaultCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "ttt",
		Short:             "tic tac toe App Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(InitCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.GenTxCmd(ctx, cdc))

	server.AddCommands(ctx, cdc, rootCmd, newApp, nil)


	executor := cli.PrepareBaseCmd(rootCmd, "TTT", DefaultNodeHome)
	if err := executor.Execute(); err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, database db.DB, traceStore io.Writer) abci.Application {
	return app.NewApp(logger, database)
}

// This will set up everything needed and create a genesis file with one validator
func InitCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis configuration, priv-validator file and p2p-node file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			chainID := viper.GetString(client.FlagChainID)

			nodeID, pk, err := gaiaInit.InitializeNodeValidatorFiles(config)
			if err != nil {
				return errors.Wrap(err, "Failed to initialize validator files")
			}

			fmt.Printf("Node ID: %s. Pub key: %s\n", nodeID, pk)

			var appStateJSON json.RawMessage
			genesisFilePath := config.GenesisFile()

			if !viper.GetBool(flagOverwrite) && common.FileExists(genesisFilePath) {
				return fmt.Errorf("genesis.json file already exists at path: %v", genesisFilePath)
			}

			appStateJSON, err = codec.MarshalJSONIndent(cdc, app.GenesisState{})
			if err != nil {
				return err
			}

			_, _, validator, err := SimpleAppGenTx(cdc, pk)
			if err != nil {
				return err
			}

			if err = gaiaInit.ExportGenesisFile(genesisFilePath, chainID, []tm.GenesisValidator{validator}, appStateJSON); err != nil {
				return errors.Wrap(err, "Failed to populate genesis.json")
			}

			// This generates a tendermint configuration. Not sure where it should go
			tendermintConfig.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			fmt.Printf("Initialized tttd configuration and bootstrapping files in %s...\n", viper.GetString(cli.HomeFlag))

			return nil
		},
	}

	cmd.Flags().String(cli.HomeFlag, DefaultNodeHome, "node's home directory")
	cmd.Flags().String(client.FlagChainID, DefaultChainID, "genesis file chain-id")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")

	return cmd
}

// SimpleAppGenTx returns a simple GenTx command that makes the node a valdiator from the start
func SimpleAppGenTx(cdc *codec.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tm.GenesisValidator, err error) {

	addr, secret, err := server.GenerateCoinKey()
	if err != nil {
		return
	}

	bz, err := cdc.MarshalJSON(struct {
		Addr sdk.AccAddress `json:"addr"`
	}{addr})
	if err != nil {
		return
	}

	appGenTx = json.RawMessage(bz)

	bz, err = cdc.MarshalJSON(map[string]string{"secret": secret})
	if err != nil {
		return
	}

	cliPrint = json.RawMessage(bz)

	validator = tm.GenesisValidator{
		PubKey: pk,
		Power:  10,
	}

	return
}
