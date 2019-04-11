package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"tic_tac_toe/x/tic_tac_toe"
)

func GetCmdQueryGame(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "game [game_id]",
		Short: "checks the game state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			gameStr := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, tic_tac_toe.QueryGame, gameStr), nil)
			if err != nil {
				fmt.Printf("Could not check %s: %s\n", gameStr, err)
				return nil
			}

			fmt.Println(string(res))

			return nil
		},
	}
}
