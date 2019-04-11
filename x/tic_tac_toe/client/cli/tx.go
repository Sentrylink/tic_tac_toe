package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/spf13/cobra"
	"tic_tac_toe/x/tic_tac_toe"
	"strconv"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetCmdStartGame(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "start [opponent_address]",
		Short: "starts a new game",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			opponent, err := sdkTypes.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()

			msg := tic_tac_toe.NewMsgStartGame(sender, opponent)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return SendTx(txBldr, cliCtx, []sdkTypes.Msg{msg})
		},
	}
}

func GetCmdPlay(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "play [game_id] [field_id]",
		Short: "plays a move in the game",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			gameIdStr := args[0]
			fieldStr := args[1]

			gameId, err := strconv.Atoi(gameIdStr)
			if err != nil {
				return err
			}

			field, err := strconv.Atoi(fieldStr)
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()

			msg := tic_tac_toe.NewMsgPlay(uint(gameId), sender, uint(field))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return SendTx(txBldr, cliCtx, []sdkTypes.Msg{msg})
		},
	}
}

func SendTx(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg) error {
	if err := cliCtx.EnsureAccountExists(); err != nil {
		txBldr = txBldr.WithAccountNumber(0)
		txBldr = txBldr.WithSequence(0)
	} else {
		from := cliCtx.GetFromAddress()

		accNum, err := cliCtx.GetAccountNumber(from)
		if err != nil {
			return err
		}
		txBldr = txBldr.WithAccountNumber(accNum)

		accSeq, err := cliCtx.GetAccountSequence(from)
		if err != nil {
			return err
		}
		txBldr = txBldr.WithSequence(accSeq)
	}

	fromName := cliCtx.GetFromName()

	passphrase, err := keys.GetPassphrase(fromName)
	if err != nil {
		return err
	}

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := cliCtx.BroadcastTx(txBytes)
	cliCtx.PrintOutput(res)
	return err
}