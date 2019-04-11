package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	cli "tic_tac_toe/x/tic_tac_toe/client/cli"
)

type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "tic_tac_toe",
		Short: "Tic tac toe querying subcommands",
	}

	queryCmd.AddCommand(client.GetCommands(
		cli.GetCmdQueryGame(mc.storeKey, mc.cdc),
	)...)

	return queryCmd
}

func (mc ModuleClient) GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tic_tac_toe",
		Short: "Tic tac toe transaction subcommands",
	}

	txCmd.AddCommand(client.PostCommands(
		cli.GetCmdStartGame(mc.cdc),
		cli.GetCmdPlay(mc.cdc),
	)...)

	return txCmd
}
