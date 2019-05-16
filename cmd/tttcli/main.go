package main

import (
	"github.com/spf13/viper"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"tic_tac_toe"
	ticTacToeClient "tic_tac_toe/x/tic_tac_toe/client"
)

const (
	storeTicTacToe = "tictactoe"
)

var (
	defaultCLIHome = os.ExpandEnv("$HOME/.tttcli")
)

func main() {
	cobra.EnableCommandSorting = false
	cdc := app.MakeDefaultCodec()

	config := sdkTypes.GetConfig()
	config.SetBech32PrefixForAccount(sdkTypes.Bech32PrefixAccAddr, sdkTypes.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdkTypes.Bech32PrefixValAddr, sdkTypes.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdkTypes.Bech32PrefixConsAddr, sdkTypes.Bech32PrefixConsPub)

	config.Seal()

	tttModuleClient := ticTacToeClient.NewModuleClient(storeTicTacToe, cdc)

	mc := []sdkTypes.ModuleClients{
		tttModuleClient,
	}

	rootCmd := &cobra.Command{
		Use:   "tttcli",
		Short: "Tic tac toe client",
	}

	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(defaultCLIHome),
		queryCmd(cdc, mc),
		txCmd(cdc, mc),
		client.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
		client.LineBreak,
	)

	executor := cli.PrepareMainCmd(rootCmd, "ttt", defaultCLIHome)
	if err := executor.Execute(); err != nil {
		panic(err)
	}
}

func registerRoutes(rs *lcd.RestServer) {
	rs.CliCtx = rs.CliCtx.WithAccountDecoder(rs.Cdc)
	rpc.RegisterRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
}

func queryCmd(cdc *codec.Codec, mc []sdkTypes.ModuleClients) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		tx.SearchTxCmd(cdc),
		authcli.GetAccountCmd("acc", cdc),
	)

	for _, m := range mc {
		queryCmd.AddCommand(m.GetQueryCmd())
	}

	return queryCmd
}

func txCmd(cdc *amino.Codec, mc []sdkTypes.ModuleClients) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcli.SendTxCmd(cdc),
		client.LineBreak,
		authcli.GetSignCommand(cdc),
		client.LineBreak,
	)

	for _, m := range mc {
		txCmd.AddCommand(m.GetTxCmd())
	}

	return txCmd
}

func prepareTxBuilder(addr sdkTypes.AccAddress, txBldr authtxb.TxBuilder, cliCtx context.CLIContext) (authtxb.TxBuilder, error) {
	if txBldr.AccountNumber() != 0 && txBldr.Sequence() != 0 {
		return txBldr, nil
	}

	acc, err := cliCtx.GetAccount(addr)
	if err != nil {
		return txBldr, err
	}

	if txBldr.AccountNumber() == 0 {
		txBldr = txBldr.WithAccountNumber(acc.GetAccountNumber())
	}

	if txBldr.Sequence() == 0 {
		txBldr = txBldr.WithSequence(acc.GetSequence())
	}

	return txBldr, nil
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
