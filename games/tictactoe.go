package games

import (
	"fmt"
	"log"

	"errors"

	"encoding/json"

	"github.com/mleonard87/merknera/repository"
)

const (
	TICTACTOE_MNEMONIC             = "TICTACTOE"
	TICTACTOE_NAME                 = "Tic-Tac-Toe"
	TICTACTOE_RPC_METHOD_NEXT_MOVE = "TicTacToe.NextMove"
	TICTACTOE_RPC_METHOD_COMPLETE  = "TicTacToe.Complete"
	TICTACTOE_RPC_METHOD_ERROR     = "TicTacToe.Error"
)

func init() {
	err := RegisterGameManager(new(TicTacToeGameManager))
	if err != nil {
		log.Fatal(err)
	}
}

type TicTacToeGameState []string

func (tgs *TicTacToeGameState) MarshalJSON() ([]byte, error) {
	sa := []string(*tgs)
	resultString := "{"
	for i, m := range sa {
		switch m {
		case "X":
			resultString += "\"X\""
		case "O":
			resultString += "\"O\""
		default:
			resultString += "null"
		}
		if i != len(sa) {
			resultString += ", "
		}
	}
	resultString += "}"

	fmt.Println("Marshalling JSON")
	fmt.Println(resultString)

	return []byte(resultString), nil
}

type TicTacToeGameManager struct{}

func (tgm TicTacToeGameManager) GenerateGames(bot repository.Bot) []repository.Game {
	gameType, err := repository.GetGameTypeByMnemonic(TICTACTOE_MNEMONIC)
	if err != nil {
		log.Fatal(err)
	}

	botList, err := repository.ListBotsForGameType(gameType)
	if err != nil {
		log.Fatal(err)
	}

	var gameList []repository.Game
	for _, b := range botList {
		// If its not the same bot as we are invoking this game for then create the game.
		if b.Id != bot.Id {
			// Create a game for these two bots with the initial bot as player 1
			game1, err := createGameWithPlayers(gameType, &b, &bot)
			if err != nil {
				log.Fatal(err)
			}
			// Create a game for these two bots with the initial bot as player 2
			game2, err := createGameWithPlayers(gameType, &bot, &b)
			if err != nil {
				log.Fatal(err)
			}
			gameList = append(gameList, game1, game2)
		}
	}

	return gameList
}

func (tgm TicTacToeGameManager) Mnemonic() string {
	return TICTACTOE_MNEMONIC
}

func (tgm TicTacToeGameManager) Name() string {
	return TICTACTOE_NAME
}

func (tgm TicTacToeGameManager) GetNextMoveRPCMethodName() string {
	return TICTACTOE_RPC_METHOD_NEXT_MOVE
}

func (tgm TicTacToeGameManager) GetCompleteRPCMethodName() string {
	return TICTACTOE_RPC_METHOD_COMPLETE
}

func (tgm TicTacToeGameManager) GetErrorRPCMethodName() string {
	return TICTACTOE_RPC_METHOD_ERROR
}

type nextMoveParams struct {
	GameId    int                `json:"gameid"`
	Mark      string             `json:"mark"`
	GameState TicTacToeGameState `json:"gamestate"`
}

func (tgm TicTacToeGameManager) GetNextMoveRPCParams(gameMove repository.GameMove) (interface{}, error) {
	gb, err := gameMove.GameBot()
	if err != nil {
		return nil, err
	}

	mark := getMarkForPlaySequence(gb.PlaySequence)

	g, err := gb.Game()
	if err != nil {
		return nil, err
	}

	gs, err := g.GameState()
	if err != nil {
		return nil, err
	}

	var tttGameState TicTacToeGameState
	err = json.Unmarshal([]byte(gs), &tttGameState)
	if err != nil {
		return nil, err
	}

	params := nextMoveParams{
		GameId:    g.Id,
		Mark:      mark,
		GameState: tttGameState,
	}

	return params, nil
}

type nextMoveResponse struct {
	Position int `json:"position"`
}

func (tgm TicTacToeGameManager) GetNextMoveRPCResult(gameMove repository.GameMove) interface{} {
	return nextMoveResponse{}
}

func (tgm TicTacToeGameManager) ProcessMove(gameMove repository.GameMove, result map[string]interface{}) (interface{}, GameResult, error) {
	var position int
	if pos, ok := result["position"].(float64); ok {
		position = int(pos)
	} else {
		return nil, GAME_RESULT_UNDECIDED, errors.New("Could not find property \"position\" in your response or position was not an integer.")
	}

	gb, err := gameMove.GameBot()
	if err != nil {
		return nil, GAME_RESULT_UNDECIDED, err
	}

	game, err := gb.Game()
	if err != nil {
		return nil, GAME_RESULT_UNDECIDED, err
	}

	gs, err := game.GameState()
	if err != nil {
		return nil, GAME_RESULT_UNDECIDED, err
	}

	var tttGameState TicTacToeGameState
	err = json.Unmarshal([]byte(gs), &tttGameState)
	if err != nil {
		return nil, GAME_RESULT_UNDECIDED, err
	}

	// Check that the position played is within the range of the game board.
	if len(tttGameState) < position || position < 0 {
		msg := fmt.Sprintf("Invalid position: \"%d\" is not a valid position in a 3x3 Tic-Tac-Toe board. Valid positions are 0-8 inclusive.")
		return nil, GAME_RESULT_UNDECIDED, errors.New(msg)
	}

	// Check that the position played has not already been played.
	if tttGameState[position] != "" {
		msg := fmt.Sprintf("Invalid position: The position you played, \"%d\", is already taken by \"%s\"", position, tttGameState[position])
		return nil, GAME_RESULT_UNDECIDED, errors.New(msg)
	}

	mark := getMarkForPlaySequence(gb.PlaySequence)
	tttGameState[position] = mark

	win := isWinForMark(tttGameState, mark)

	// Detect if a draw has occurred.
	if !win {
		spacesLeft := false
		for _, v := range tttGameState {
			if v == "" {
				spacesLeft = true
			}
		}

		if !spacesLeft {
			return tttGameState, GAME_RESULT_DRAW, nil
		}
	} else {
		return tttGameState, GAME_RESULT_WIN, nil
	}

	return tttGameState, GAME_RESULT_UNDECIDED, nil
}

func (tgm TicTacToeGameManager) GetGameBotForNextMove(currentMove repository.GameMove) (repository.GameBot, error) {
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

type completeParams struct {
	GameId    int                `json:"gameid"`
	Winner    bool               `json:"winner"`
	Mark      string             `json:"mark"`
	GameState TicTacToeGameState `json:"gamestate"`
}

func (tgm TicTacToeGameManager) GetCompleteRPCParams(gb repository.GameBot, gr GameResult) (interface{}, error) {
	game, err := gb.Game()
	if err != nil {
		return nil, err
	}

	gs, err := game.GameState()
	if err != nil {
		return nil, err
	}

	win := false
	if gr == GAME_RESULT_WIN {
		wm, err := game.WinningMove()
		if err != nil {
			return nil, err
		}

		winninggb, err := wm.GameBot()
		if err != nil {
			return nil, err
		}

		if gb.Id == winninggb.Id {
			win = true
		}
	}

	var tgs TicTacToeGameState
	err = json.Unmarshal([]byte(gs), &tgs)
	if err != nil {
		return nil, err
	}
	cp := completeParams{
		GameId:    game.Id,
		Winner:    win,
		Mark:      getMarkForPlaySequence(gb.PlaySequence),
		GameState: tgs,
	}

	return cp, nil
}

type errorParams struct {
	GameId    int    `json:"gameid"`
	Message   string `json:"message"`
	ErrorCode int    `json:"errorcode"`
}

func (tgm TicTacToeGameManager) GetErrorRPCParams(gm repository.GameMove, errorMessage string) interface{} {
	gb, _ := gm.GameBot()
	game, _ := gb.Game()
	return errorParams{
		GameId:    game.Id,
		Message:   errorMessage,
		ErrorCode: 999,
	}
}

func getMarkForPlaySequence(ps int) string {
	switch ps {
	case 1:
		return "X"
	case 2:
		return "O"
	default:
		log.Fatal("Invalid play sequence for Tic-Tac-Toe")
	}

	return ""
}

func createGameWithPlayers(gameType repository.GameType, playerOne *repository.Bot, playerTwo *repository.Bot) (repository.Game, error) {
	game, err := repository.CreateGame(gameType)
	if err != nil {
		return game, err
	}

	_, err = repository.CreateGameBot(game, *playerOne, 1)
	if err != nil {
		return game, err
	}

	_, err = repository.CreateGameBot(game, *playerTwo, 2)
	if err != nil {
		return game, err
	}

	err = createFirstGameMove(game)
	if err != nil {
		return game, err
	}

	return game, nil
}

func createFirstGameMove(game repository.Game) error {
	players, err := game.Players()
	if err != nil {
		return err
	}
	firstPlayer := players[0]
	initialGameState := make([]string, 9, 9)
	_, err = repository.CreateGameMove(firstPlayer, initialGameState)
	if err != nil {
		return err
	}

	return nil
}

func isWinForMark(gs []string, m string) bool {
	switch {
	case gs[0] == m && gs[3] == m && gs[6] == m:
		return true
	case gs[0] == m && gs[4] == m && gs[8] == m:
		return true
	case gs[1] == m && gs[4] == m && gs[7] == m:
		return true
	case gs[2] == m && gs[5] == m && gs[8] == m:
		return true
	case gs[2] == m && gs[4] == m && gs[6] == m:
		return true
	case gs[0] == m && gs[1] == m && gs[2] == m:
		return true
	case gs[3] == m && gs[4] == m && gs[5] == m:
		return true
	case gs[6] == m && gs[7] == m && gs[8] == m:
		return true
	default:
		return false
	}
}
