package tic_tac_toe

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgStartGame{}, "tictactoe/StartGame", nil)
	cdc.RegisterConcrete(MsgPlay{}, "tictactoe/Play", nil)
	cdc.RegisterConcrete(Game{}, "tictactoe/Game", nil)
}
