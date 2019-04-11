package tic_tac_toe

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Game struct {
	Id      uint            `json:"id"`
	Player1 sdk.AccAddress  `json:"player_1"`
	Player2 sdk.AccAddress  `json:"player_2"`
	Fields  map[string]uint `json:"fields"`
	Winner  uint            `json:"winner"`
}
