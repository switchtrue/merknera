package games

import (
	"fmt"
	"log"

	"errors"

	"github.com/mleonard87/merknera/repository"
)

const (
	TICTACTOE_MNEMONIC                  = "TICTACTOE"
	TICTACTOE_NAME                      = "Tic-Tac-Toe"
	TICTACTOE_RPC_NAMESPACE             = "TicTacToe"
	TICTACTOE_RPC_METHOD_NAME_NEXT_MOVE = "NextMove"
)

type TicTacToeGameManager struct{}

func init() {
	err := RegisterGameManager(new(TicTacToeGameManager))
	if err != nil {
		log.Fatal(err)
	}
}

func (tgm TicTacToeGameManager) Mnemonic() string {
	return TICTACTOE_MNEMONIC
}

func (tgm TicTacToeGameManager) Name() string {
	return TICTACTOE_NAME
}

func (tgm TicTacToeGameManager) RPCNamespace() string {
	return TICTACTOE_RPC_NAMESPACE
}

type TicTacToeGameState []string

type TicTacToeNextMoveRequestParams struct {
	GameId    int                `json:"gameid"`
	Mark      string             `json:"mark"`
	GameState TicTacToeGameState `json:"gamestate"`
}

type TicTacToeNextMoveResponseResult struct {
	Position int `json:"position"`
}

type TicTacToeCompleteRequestParams struct {
	GameId    int                `json:"gameid"`
	Mark      string             `json:"mark"`
	GameState TicTacToeGameState `json:"gamestate"`
	Winner    bool               `json:"winner"`
}

type TicTacToeErrorRequestParams struct {
	GameId  int    `json:"gameid"`
	Message string `json:"message"`
}

func (tgm TicTacToeGameManager) Begin(game repository.Game) (rpcMethod string, initialPlayer repository.GameBot, gameState interface{}, err error) {
	var initialGameState TicTacToeGameState
	initialGameState = make([]string, 9, 9)

	players, err := game.Players()
	if err != nil {
		return "", repository.GameBot{}, nil, err
	}
	firstPlayer := players[0]

	return "TicTacToe.NextMove", firstPlayer, initialGameState, nil
}

func (tgm TicTacToeGameManager) Resume(game repository.Game) (rpcMethod string, err error) {
	return fmt.Sprintf("%s.%s", TICTACTOE_RPC_NAMESPACE, TICTACTOE_RPC_METHOD_NAME_NEXT_MOVE), nil
}

func (tgm TicTacToeGameManager) CompleteRequestParams(game *repository.Game, bot *repository.Bot, finalGameState *TicTacToeGameState) (TicTacToeCompleteRequestParams, error) {
	wb, err := game.WinningBot()
	if err != nil {
		return TicTacToeCompleteRequestParams{}, err
	}

	isWinner := bot.Id == wb.Id

	mark, err := gamePlayerMark(game, bot)
	if err != nil {
		return TicTacToeCompleteRequestParams{}, err
	}

	return TicTacToeCompleteRequestParams{
		GameId:    game.Id,
		Mark:      mark,
		GameState: finalGameState,
		Winner:    isWinner,
	}
}

func gamePlayerMark(game repository.Game, bot repository.Bot) (string, error) {
	p, err := game.Players()
	if err != nil {
		return "", err
	}

	for i, player := range p {
		if player.Id == bot.Id {
			switch i {
			case 0:
				return "X"
			case 1:
				return "O"
			}
		}
	}

	return "", errors.New("Unable to determine mark for player.")
}

func (tgm TicTacToeGameManager) ErrorRequestParams(game repository.Game, message string) (TicTacToeErrorRequestParams, error) {
	return TicTacToeErrorRequestParams{
		GameId:  game.Id,
		Message: message,
	}
}

func (tgm TicTacToeGameManager) NextPlayer(currentMove repository.GameMove) (repository.GameBot, error) {
	gb, err := currentMove.GameBot()
	if err != nil {
		return repository.GameBot{}, err
	}

	game, err := gb.Game()
	if err != nil {
		return repository.GameBot{}, err
	}

	gameBots, err := game.Players()
	if err != nil {
		return repository.GameBot{}, err
	}

	for _, b := range gameBots {
		if b.Id != gb.Id {
			return b, nil
		}
	}

	return repository.GameBot{}, errors.New("Could not find GameBot for next move.")
}

func (tgm TicTacToeGameManager) NextMoveRequestParams(gameMove *repository.GameMove, gameState *TicTacToeGameState) (params TicTacToeNextMoveRequestParams, err error) {
	//gb, err := gameMove.GameBot()
	//if err != nil {
	//	return nil, err
	//}
	//
	//mark := getMarkForPlaySequence(gb.PlaySequence)
	//
	//g, err := gb.Game()
	//if err != nil {
	//	return nil, err
	//}
	//
	//gs, err := g.GameState()
	//if err != nil {
	//	return nil, err
	//}
	//
	//var tttGameState TicTacToeGameState
	//err = json.Unmarshal([]byte(gs), &tttGameState)
	//if err != nil {
	//	return nil, err
	//}
	//
	//params := nextMoveParams{
	//	GameId:    g.Id,
	//	Mark:      mark,
	//	GameState: tttGameState,
	//}
	//
	//return params, nil
	fmt.Println("In NextMoveGetRequestParamsBody")
	fmt.Println("gameMove:")
	fmt.Println(gameMove)
	fmt.Println("gameState:")
	fmt.Println(gameState)

	return TicTacToeNextMoveRequestParams{}, nil
}

func (tgm TicTacToeGameManager) NextMoveProcessResponse(gameMove *repository.GameMove, result *TicTacToeNextMoveResponseResult, gameState *TicTacToeGameState) (gameResult GameResult, nextRpcMethodName string, newGameState TicTacToeGameState, err error) {
	fmt.Println("In NextMoveProcessResponse")

	return GAME_RESULT_UNDECIDED, "", TicTacToeGameState{}, nil
}

//type TicTacToeGameState []string

//func (tgs *TicTacToeGameState) MarshalJSON() ([]byte, error) {
//	sa := []string(*tgs)
//	resultString := "{"
//	for i, m := range sa {
//		switch m {
//		case "X":
//			resultString += "\"X\""
//		case "O":
//			resultString += "\"O\""
//		default:
//			resultString += "null"
//		}
//		if i != len(sa) {
//			resultString += ", "
//		}
//	}
//	resultString += "}"
//
//	fmt.Println("Marshalling JSON")
//	fmt.Println(resultString)
//
//	return []byte(resultString), nil
//}
//
func (tgm TicTacToeGameManager) GetGamesForBot(bot repository.Bot, otherBots []repository.Bot) [][]*repository.Bot {
	var gameList [][]*repository.Bot

	// For each other bot schedule a game where the bot is in both the first and second player positions.
	for _, b := range otherBots {
		var game1 []*repository.Bot
		game1 = append(game1, &bot, &b) // Bot goes first.

		var game2 []*repository.Bot
		game2 = append(game2, &b, &bot) // Bot goes second.
	}

	return gameList
}

//func (tgm TicTacToeGameManager) GetNextMoveRPCMethodName() string {
//	return TICTACTOE_RPC_METHOD_NEXT_MOVE
//}
//
//func (tgm TicTacToeGameManager) GetCompleteRPCMethodName() string {
//	return TICTACTOE_RPC_METHOD_COMPLETE
//}
//
//func (tgm TicTacToeGameManager) GetErrorRPCMethodName() string {
//	return TICTACTOE_RPC_METHOD_ERROR
//}

//type nextMoveParams struct {
//	GameId    int                `json:"gameid"`
//	Mark      string             `json:"mark"`
//	GameState TicTacToeGameState `json:"gamestate"`
//}
//
//func (tgm TicTacToeGameManager) GetNextMoveRPCParams(gameMove repository.GameMove) (interface{}, error) {
//	gb, err := gameMove.GameBot()
//	if err != nil {
//		return nil, err
//	}
//
//	mark := getMarkForPlaySequence(gb.PlaySequence)
//
//	g, err := gb.Game()
//	if err != nil {
//		return nil, err
//	}
//
//	gs, err := g.GameState()
//	if err != nil {
//		return nil, err
//	}
//
//	var tttGameState TicTacToeGameState
//	err = json.Unmarshal([]byte(gs), &tttGameState)
//	if err != nil {
//		return nil, err
//	}
//
//	params := nextMoveParams{
//		GameId:    g.Id,
//		Mark:      mark,
//		GameState: tttGameState,
//	}
//
//	return params, nil
//}

//type nextMoveResponse struct {
//	Position int `json:"position"`
//}

//func (tgm TicTacToeGameManager) GetNextMoveRPCResult(gameMove repository.GameMove) interface{} {
//	return nextMoveResponse{}
//}

//func (tgm TicTacToeGameManager) ProcessMove(gameMove repository.GameMove, result map[string]interface{}) (interface{}, GameResult, error) {
//	var position int
//	if pos, ok := result["position"].(float64); ok {
//		position = int(pos)
//	} else {
//		return nil, GAME_RESULT_UNDECIDED, errors.New("Could not find property \"position\" in your response or position was not an integer.")
//	}
//
//	gb, err := gameMove.GameBot()
//	if err != nil {
//		return nil, GAME_RESULT_UNDECIDED, err
//	}
//
//	game, err := gb.Game()
//	if err != nil {
//		return nil, GAME_RESULT_UNDECIDED, err
//	}
//
//	gs, err := game.GameState()
//	if err != nil {
//		return nil, GAME_RESULT_UNDECIDED, err
//	}
//
//	var tttGameState TicTacToeGameState
//	err = json.Unmarshal([]byte(gs), &tttGameState)
//	if err != nil {
//		return nil, GAME_RESULT_UNDECIDED, err
//	}
//
//	// Check that the position played is within the range of the game board.
//	if len(tttGameState) < position || position < 0 {
//		msg := fmt.Sprintf("Invalid position: \"%d\" is not a valid position in a 3x3 Tic-Tac-Toe board. Valid positions are 0-8 inclusive.")
//		return nil, GAME_RESULT_UNDECIDED, errors.New(msg)
//	}
//
//	// Check that the position played has not already been played.
//	if tttGameState[position] != "" {
//		msg := fmt.Sprintf("Invalid position: The position you played, \"%d\", is already taken by \"%s\"", position, tttGameState[position])
//		return nil, GAME_RESULT_UNDECIDED, errors.New(msg)
//	}
//
//	mark := getMarkForPlaySequence(gb.PlaySequence)
//	tttGameState[position] = mark
//
//	win := isWinForMark(tttGameState, mark)
//
//	// Detect if a draw has occurred.
//	if !win {
//		spacesLeft := false
//		for _, v := range tttGameState {
//			if v == "" {
//				spacesLeft = true
//			}
//		}
//
//		if !spacesLeft {
//			return tttGameState, GAME_RESULT_DRAW, nil
//		}
//	} else {
//		return tttGameState, GAME_RESULT_WIN, nil
//	}
//
//	return tttGameState, GAME_RESULT_UNDECIDED, nil
//}

//func (tgm TicTacToeGameManager) GetGameBotForNextMove(currentMove repository.GameMove) (repository.GameBot, error) {
//	gb, err := currentMove.GameBot()
//	if err != nil {
//		return repository.GameBot{}, err
//	}
//
//	game, err := gb.Game()
//	if err != nil {
//		return repository.GameBot{}, err
//	}
//
//	gameBots, err := game.Players()
//	if err != nil {
//		return repository.GameBot{}, err
//	}
//
//	for _, b := range gameBots {
//		if b.Id != gb.Id {
//			return b, nil
//		}
//	}
//
//	return repository.GameBot{}, errors.New("Could not find GameBot for next move.")
//}

//type completeParams struct {
//	GameId    int                `json:"gameid"`
//	Winner    bool               `json:"winner"`
//	Mark      string             `json:"mark"`
//	GameState TicTacToeGameState `json:"gamestate"`
//}
//
//func (tgm TicTacToeGameManager) GetCompleteRPCParams(gb repository.GameBot, gr GameResult) (interface{}, error) {
//	game, err := gb.Game()
//	if err != nil {
//		return nil, err
//	}
//
//	gs, err := game.GameState()
//	if err != nil {
//		return nil, err
//	}
//
//	win := false
//	if gr == GAME_RESULT_WIN {
//		wm, err := game.WinningMove()
//		if err != nil {
//			return nil, err
//		}
//
//		winninggb, err := wm.GameBot()
//		if err != nil {
//			return nil, err
//		}
//
//		if gb.Id == winninggb.Id {
//			win = true
//		}
//	}
//
//	var tgs TicTacToeGameState
//	err = json.Unmarshal([]byte(gs), &tgs)
//	if err != nil {
//		return nil, err
//	}
//	cp := completeParams{
//		GameId:    game.Id,
//		Winner:    win,
//		Mark:      getMarkForPlaySequence(gb.PlaySequence),
//		GameState: tgs,
//	}
//
//	return cp, nil
//}

//type errorParams struct {
//	GameId    int    `json:"gameid"`
//	Message   string `json:"message"`
//	ErrorCode int    `json:"errorcode"`
//}
//
//func (tgm TicTacToeGameManager) GetErrorRPCParams(gm repository.GameMove, errorMessage string) interface{} {
//	gb, _ := gm.GameBot()
//	game, _ := gb.Game()
//	return errorParams{
//		GameId:    game.Id,
//		Message:   errorMessage,
//		ErrorCode: 9999,
//	}
//}

//func createGameWithPlayers(gameType repository.GameType, playerOne *repository.Bot, playerTwo *repository.Bot) (repository.Game, error) {
//
//	_, err = repository.CreateGameBot(game, *playerOne, 1)
//	if err != nil {
//		return game, err
//	}
//
//	_, err = repository.CreateGameBot(game, *playerTwo, 2)
//	if err != nil {
//		return game, err
//	}
//
//	err = createFirstGameMove(game)
//	if err != nil {
//		return game, err
//	}
//
//	return game, nil
//}
//
//func createFirstGameMove(game repository.Game) error {
//	players, err := game.Players()
//	if err != nil {
//		return err
//	}
//	firstPlayer := players[0]
//	initialGameState := make([]string, 9, 9)
//	_, err = repository.CreateGameMove(firstPlayer, initialGameState)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

//func isWinForMark(gs []string, m string) bool {
//	switch {
//	case gs[0] == m && gs[3] == m && gs[6] == m:
//		return true
//	case gs[0] == m && gs[4] == m && gs[8] == m:
//		return true
//	case gs[1] == m && gs[4] == m && gs[7] == m:
//		return true
//	case gs[2] == m && gs[5] == m && gs[8] == m:
//		return true
//	case gs[2] == m && gs[4] == m && gs[6] == m:
//		return true
//	case gs[0] == m && gs[1] == m && gs[2] == m:
//		return true
//	case gs[3] == m && gs[4] == m && gs[5] == m:
//		return true
//	case gs[6] == m && gs[7] == m && gs[8] == m:
//		return true
//	default:
//		return false
//	}
//}
