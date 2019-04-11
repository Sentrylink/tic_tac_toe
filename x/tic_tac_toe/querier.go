package tic_tac_toe

import (
	"encoding/json"
	"fmt"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"strconv"
)

const (
	QueryGame = "game"
)

func NewQuerier(keeper Keeper) sdkTypes.Querier {
	return func(ctx sdkTypes.Context, path []string, req abci.RequestQuery) ([]byte, sdkTypes.Error) {
		switch path[0] {
		case QueryGame:
			return queryGame(ctx, path[1:], req, keeper)
		default:
			return nil, sdkTypes.ErrUnknownRequest("unknown kyc query endpoint")
		}
	}
}

func queryGame(ctx sdkTypes.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdkTypes.Error) {
	idStr := path[0]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, sdkTypes.ErrUnknownRequest(fmt.Sprintf("Bad game id %s", err))
	}

	game := keeper.getGame(ctx, uint(id))
	if game == nil {
		return nil, sdkTypes.ErrUnknownRequest("No such game")
	}

	gameJson, err := json.Marshal(game)
	if err != nil {
		panic(fmt.Sprintf("Failed to encode game"))
	}

	return gameJson, nil
}
