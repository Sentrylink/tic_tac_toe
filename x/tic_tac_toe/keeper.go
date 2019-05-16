package tic_tac_toe

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
)

type Keeper struct {
	key sdk.StoreKey
	cdc *codec.Codec
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey) Keeper {
	return Keeper{
		cdc: cdc,
		key: key,
	}
}

func (k Keeper) setGameId(ctx sdk.Context, id uint) {
	idBytes := []byte(strconv.Itoa(int(id)))
	store := ctx.KVStore(k.key)
	store.Set([]byte("id"), idBytes)
}

func (k Keeper) getGameId(ctx sdk.Context) int {
	store := ctx.KVStore(k.key)
	idBytes := store.Get([]byte("id"))
	if idBytes == nil {
		return -1
	}

	id, err := strconv.Atoi(string(idBytes))
	if err != nil {
		panic(fmt.Sprintf("Invalid game id: %v", idBytes))
	}

	return id
}

func (k Keeper) storeGame(ctx sdk.Context, game *Game) {
	store := ctx.KVStore(k.key)
	key := strconv.Itoa(int(game.Id))
	value := k.cdc.MustMarshalJSON(game)
	store.Set([]byte(key), value)
}

func (k Keeper) getGame(ctx sdk.Context, id uint) *Game {
	store := ctx.KVStore(k.key)
	key := strconv.Itoa(int(id))
	value := store.Get([]byte(key))
	if value == nil {
		return nil
	}

	game := new(Game)
	err := k.cdc.UnmarshalJSON(value, game)
	if err != nil {
		panic(fmt.Sprintf("Invalid game stored: %s", err))
	}

	return game
}

func (k Keeper) StartGame(ctx sdk.Context, player1, player2 sdk.AccAddress) *Game {
	nextGameID := k.getGameId(ctx) + 1
	k.setGameId(ctx, uint(nextGameID))
	game := &Game{
		Id:      uint(nextGameID),
		Player1: player1,
		Player2: player2,
		Fields:  emptyFields(),
		Winner:  0,
	}

	k.storeGame(ctx, game)

	return game
}

func (k Keeper) Play(ctx sdk.Context, gameID uint, player sdk.AccAddress, field uint) sdk.Result {
	game := k.getGame(ctx, gameID)
	if game == nil {
		return sdk.ErrUnknownRequest("No such game").Result()
	}

	if !game.Player1.Equals(player) && !game.Player2.Equals(player) {
		return sdk.ErrUnauthorized("Not playing in this game").Result()
	}

	if game.Winner != 0 {
		return sdk.ErrUnknownRequest("Game already finished").Result()
	}

	var player1 bool
	if game.Player1.Equals(player) {
		player1 = true
	}

	var player1ShouldPlay bool
	movesPlayed := totalMoves(game.Fields)
	if movesPlayed%2 == 0 {
		player1ShouldPlay = true
	}

	if (player1 && !player1ShouldPlay) || (!player1 && player1ShouldPlay) {
		return sdk.ErrUnknownRequest("Not your turn").Result()
	}

	fieldStr := strconv.Itoa(int(field))

	if isFieldTaken(game.Fields, fieldStr) {
		return sdk.ErrUnknownRequest("Field is already taken").Result()
	}

	var mark uint
	if player1 {
		mark = 1
	} else {
		mark = 2
	}

	game.Fields[fieldStr] = mark

	checkWinner(game)

	k.storeGame(ctx, game)

	return sdk.Result{}
}

func emptyFields() map[string]uint {
	return map[string]uint{
		"0": 0,
		"1": 0,
		"2": 0,
		"3": 0,
		"4": 0,
		"5": 0,
		"6": 0,
		"7": 0,
		"8": 0,
	}
}

func totalMoves(fields map[string]uint) int {
	var movesPlayed int

	for _, player := range fields {
		if player != 0 {
			movesPlayed++
		}
	}

	return movesPlayed
}

func isFieldTaken(fields map[string]uint, field string) bool {
	player := fields[field]
	return player != 0
}

func checkWinner(game *Game) {
	fields := game.Fields

	if fields["0"] != 0 {
		if fields["0"] == fields["1"] && fields["1"] == fields["2"] {
			game.Winner = fields["0"]
			return
		}

		if fields["0"] == fields["4"] && fields["0"] == fields["8"] {
			game.Winner = fields["0"]
			return
		}

		if fields["0"] == fields["3"] && fields["0"] == fields["6"] {
			game.Winner = fields["0"]
			return
		}
	}

	if fields["8"] != 0 {
		if fields["8"] == fields["7"] && fields["8"] == fields["6"] {
			game.Winner = fields["8"]
			return
		}

		if fields["8"] == fields["5"] && fields["8"] == fields["2"] {
			game.Winner = fields["8"]
			return
		}
	}

	if fields["2"] != 0 && (fields["2"] == fields["4"] && fields["4"] == fields["6"]) {
		game.Winner = fields["2"]
		return
	}
}
