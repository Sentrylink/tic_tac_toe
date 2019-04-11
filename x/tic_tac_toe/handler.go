package tic_tac_toe

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgStartGame:
			return handleMsgStartGame(ctx, keeper, msg)
		case MsgPlay:
			return handleMsgPlay(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized tic tac toe Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgStartGame(ctx sdk.Context, keeper Keeper, msg MsgStartGame) sdk.Result {
	game := keeper.StartGame(ctx, msg.Inviter, msg.Opponent)
	gameData, err := json.Marshal(game)
	if err != nil {
		panic(err)
	}

	return sdk.Result{Data: gameData}
}

func handleMsgPlay(ctx sdk.Context, keeper Keeper, msg MsgPlay) sdk.Result {
	return keeper.Play(ctx, msg.GameId, msg.Player, msg.Field)
}
