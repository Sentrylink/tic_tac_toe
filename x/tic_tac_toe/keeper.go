package tic_tac_toe

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"strconv"
)

type Keeper struct {
	accountKeeper auth.AccountKeeper
	key sdk.StoreKey
	cdc *codec.Codec
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, accountKeeper auth.AccountKeeper) Keeper {
	return Keeper{
		cdc: cdc,
		key: key,
		accountKeeper: accountKeeper,
	}
}

func (k Keeper) setGameId(ctx sdk.Context, id uint) {
	idBytes := []byte(strconv.Itoa(int(id)))
	store := ctx.KVStore(k.key)
	store.Set([]byte("id"), idBytes)
}

func (k Keeper) getGameId(ctx sdk.Context) uint {
	store := ctx.KVStore(k.key)
	idBytes := store.Get([]byte("id"))
	if idBytes == nil {
		k.setGameId(ctx, 0)
		return 0
	}

	id, err := strconv.Atoi(string(idBytes))
	if err != nil {
		panic(fmt.Sprintf("Invalid game id: %v", idBytes))
	}

	return uint(id)
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

	fmt.Printf("Value in keeper: %s\n\n", value)

	game := new(Game)
	err := k.cdc.UnmarshalJSON(value, game)
	if err != nil {
		panic(fmt.Sprintf("Invalid game stored: %s", err))
	}

	return game
}

func (k Keeper) StartGame(ctx sdk.Context, player1, player2 sdk.AccAddress, amount sdk.Coin) (*Game, sdk.Result) {
	if !amount.IsZero() {
		acc1 := k.accountKeeper.GetAccount(ctx, player1)
		if acc1 == nil {
			return nil, sdk.ErrInvalidAddress("No player 1 account").Result()
		}

		acc2 := k.accountKeeper.GetAccount(ctx, player2)
		if acc2 == nil {
			return nil, sdk.ErrInvalidAddress("No player 2 account").Result()
		}

		amountCoins := sdk.Coins{amount}
		coins1 := acc1.GetCoins()
		if !coins1.IsAllGTE(amountCoins) {
			return nil, sdk.ErrInsufficientCoins("Player 1 has not enough tokens").Result()
		}

		coins2 := acc2.GetCoins()
		if !coins2.IsAllGTE(amountCoins) {
			return nil, sdk.ErrInsufficientCoins("Player 2 has not enough tokens").Result()
		}

		newCoins1 := coins1.Sub(amountCoins)
		newCoins2 := coins2.Sub(amountCoins)

		if err := acc1.SetCoins(newCoins1); err != nil {
			panic(err)
		}

		if err := acc2.SetCoins(newCoins2); err != nil {
			panic(err)
		}

		k.accountKeeper.SetAccount(ctx, acc1)
		k.accountKeeper.SetAccount(ctx, acc2)
	}

	nextGameID := k.getGameId(ctx)
	game := &Game{
		Id:      nextGameID,
		Player1: player1,
		Player2: player2,
		Fields:  emptyFields(),
		Amount: amount,
		Winner:  0,
	}

	k.storeGame(ctx, game)

	return game, sdk.Result{}
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
	if game.Winner != 0 && !game.Amount.IsZero() {
		k.distributeReward(ctx, game)
	}

	k.storeGame(ctx, game)

	return sdk.Result{}
}

func (k Keeper) distributeReward(ctx sdk.Context, game *Game) {
	var acc auth.Account
	if game.Winner == 1 {
		acc = k.accountKeeper.GetAccount(ctx, game.Player1)
	} else {
		acc = k.accountKeeper.GetAccount(ctx, game.Player2)
	}

	reward := sdk.Coins{game.Amount}
	reward = reward.Add(reward)
	currentCoins := acc.GetCoins()
	newCoins := currentCoins.Add(reward)

	if err := acc.SetCoins(newCoins); err != nil {
		panic(err)
	}

	k.accountKeeper.SetAccount(ctx, acc)
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
