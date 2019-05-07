package tic_tac_toe

import (
	"encoding/json"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
)

type MsgStartGame struct {
	Opponent sdkTypes.AccAddress `json:"opponent"`
	Inviter  sdkTypes.AccAddress `json:"inviter"`
	Amount sdkTypes.Coin `json:"amount"`
}

func NewMsgStartGame(inviter, opponent sdkTypes.AccAddress, amount sdkTypes.Coin) MsgStartGame {
	return MsgStartGame{
		Inviter:  inviter,
		Opponent: opponent,
		Amount: amount,
	}
}

func (msg MsgStartGame) Route() string {
	return "tictactoe"
}

func (msg MsgStartGame) Type() string {
	return "startgame"
}

func (msg MsgStartGame) ValidateBasic() sdkTypes.Error {
	if msg.Inviter.Empty() {
		return sdkTypes.ErrInvalidAddress("Inviter is empty")
	}

	if msg.Opponent.Empty() {
		return sdkTypes.ErrInvalidAddress("Opponent is empty")
	}

	return nil
}

func (msg MsgStartGame) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdkTypes.MustSortJSON(b)
}

func (msg MsgStartGame) GetSigners() []sdkTypes.AccAddress {
	return []sdkTypes.AccAddress{msg.Inviter}
}

//

type MsgPlay struct {
	GameId uint                `json:game_id"`
	Player sdkTypes.AccAddress `json:"player"`
	Field  uint                `json:"field"`
}

func NewMsgPlay(gameId uint, player sdkTypes.AccAddress, field uint) MsgPlay {
	return MsgPlay{
		GameId: gameId,
		Player: player,
		Field:  field,
	}
}

func (msg MsgPlay) Route() string {
	return "tictactoe"
}

func (msg MsgPlay) Type() string {
	return "play"
}

func (msg MsgPlay) ValidateBasic() sdkTypes.Error {
	if msg.Player.Empty() {
		return sdkTypes.ErrInvalidAddress("Inviter is empty")
	}

	if msg.Field > 8 {
		return sdkTypes.ErrUnknownRequest("Field has to be from 0 to 8")
	}

	return nil
}

func (msg MsgPlay) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdkTypes.MustSortJSON(b)
}

func (msg MsgPlay) GetSigners() []sdkTypes.AccAddress {
	return []sdkTypes.AccAddress{msg.Player}
}
