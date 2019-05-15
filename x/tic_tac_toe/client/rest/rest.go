package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"tic_tac_toe/x/tic_tac_toe"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/cosmos/cosmos-sdk/x/auth"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
)

// register REST routes
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc("/tictactoe/game/{gameID}", QueryGame(cdc, context.GetAccountDecoder(cdc), cliCtx)).Methods("GET")
	r.HandleFunc("/tictactoe/game", startGameHandler(cdc, cliCtx)).Methods("POST")
	r.HandleFunc("/tictactoe/game/{gameID}/play", playHandler(cdc, cliCtx)).Methods("POST")
}

// query accountREST Handler
func QueryGame(
	cdc *codec.Codec,
	decoder auth.AccountDecoder, cliCtx context.CLIContext,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		gameIDStr := vars["gameID"]
		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/tictactoe/game/%d", gameID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		fmt.Printf("Res: %s\n\n", res)

		// the query will return empty account if there is no data
		if len(res) == 0 {
			rest.PostProcessResponse(w, cdc, auth.BaseAccount{}, cliCtx.Indent)
			return
		}

		// decode the value
		game := new(tic_tac_toe.Game)
		err = json.Unmarshal(res, game)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, game, cliCtx.Indent)
	}
}


type startGameRequest struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Opponent sdk.AccAddress `json:"opponent"`
	Inviter  sdk.AccAddress `json:"inviter"`
	Amount sdk.Coin `json:"amount"`
}

func startGameHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req startGameRequest

		if !rest.ReadRESTReq(w, r, cdc, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		// create the message
		msg := tic_tac_toe.NewMsgStartGame(req.Inviter, req.Opponent, req.Amount)
		err := msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})
	}
}


type playRequest struct {
	BaseReq rest.BaseReq `json:"base_req"`
	GameId uint                `json:game_id"`
	Player sdk.AccAddress `json:"player"`
	Field  uint                `json:"field"`
}

func playHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req playRequest

		if !rest.ReadRESTReq(w, r, cdc, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		// create the message
		msg := tic_tac_toe.NewMsgPlay(req.GameId, req.Player, req.Field)
		err := msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
