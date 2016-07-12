package games

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/mleonard87/merknera/repository"
	"github.com/mleonard87/merknera/rpchelper"
)

type GameResult string

const (
	GAME_RESULT_WIN       GameResult = "WIN"
	GAME_RESULT_DRAW      GameResult = "DRAW"
	GAME_RESULT_UNDECIDED GameResult = "UNDECIDED"

	GAME_METHOD_NAME_SUFFIX_REQUEST  = "RequestParams"
	GAME_METHOD_NAME_SUFFIX_RESPONSE = "ProcessResponse"

	GAME_METHOD_NAME_COMPLETE = "CompleteRequestParams"
	GAME_METHOD_NAME_ERROR    = "ErrorRequestParams"

	RPC_METHOD_NAME_COMPLETE = "Complete"
	RPC_METHOD_NAME_ERROR    = "Error"
)

type GameProvider interface {
	Mnemonic() string
	Name() string
	RPCNamespace() string

	GetGamesForBot(bot repository.Bot, otherBots []repository.Bot) [][]*repository.Bot

	NextPlayer(currentMove repository.GameMove) (repository.GameBot, error)

	Begin(game repository.Game) (rpcMethod string, initialPlayer repository.GameBot, gameState interface{}, err error)
	Resume(game repository.Game) (rpcMethod string, err error)
}

type GameManager struct {
	GameProvider         GameProvider
	RPCNamespace         string
	CompleteParamsMethod reflect.Method
	ErrorParamsMethod    reflect.Method
	GameStateArgType     reflect.Type // Type of the game state argument used in all method calls for a given game provider.
	RPCMethods           map[string]gameManagerRPCMethod
}

type gameManagerRPCMethod struct {
	MethodName                   string
	RequestParamsMethod          reflect.Method // Method to be invoked to get the RPC request params.
	ProcessResponseMethod        reflect.Method // Method to be invoked to process the result of the RPC request.
	RequestParamsReturnType      reflect.Type   // Type of request params to be used for the RPC call.
	ProcessResponseResultArgType reflect.Type   // Type to unmarshal the RPC results into for the Process response argument.
}

var (
	// Pre-compute the reflect.Type of error and repository.GameMove
	typeOfRepositoryGame     = reflect.TypeOf((*repository.Game)(nil))
	typeOfRepositoryGameMove = reflect.TypeOf((*repository.GameMove)(nil))
	typeOfRepositoryBot      = reflect.TypeOf((*repository.Bot)(nil))
	typeOfErrorInterface     = reflect.TypeOf((*error)(nil)).Elem()
	registeredGameManagers   = make(map[string]GameManager)
)

// A wrapper around a map with an Add method that acts as a set.
type gameRPCMethodSet struct {
	set map[string]bool
}

func (grms *gameRPCMethodSet) Add(rpcMethodName string) bool {
	_, found := grms.set[rpcMethodName]
	grms.set[rpcMethodName] = true
	return !found
}

func RegisterGameManager(gp GameProvider) error {
	_, err := repository.GetGameTypeByMnemonic(gp.Mnemonic())
	if err != nil {
		_, err := repository.CreateGameType(gp.Mnemonic(), gp.Name())
		if err != nil {
			return err
		}
	}

	gmType := reflect.TypeOf(gp)

	grms := gameRPCMethodSet{}
	grms.set = make(map[string]bool)

	// Replace the method name suffixes with nothing to get the raw RPC method name (without RPC namespace).
	r := strings.NewReplacer(GAME_METHOD_NAME_SUFFIX_REQUEST, "", GAME_METHOD_NAME_SUFFIX_RESPONSE, "")

	// Create a variable of type GameManager and get its type.
	var gmi *GameProvider
	gmiType := reflect.TypeOf(gmi).Elem()
	gmiMethods := make(map[string]bool)

	// Iterate over all the methods declared by the interface type of GameManager and add them to a map. Later this
	// is used to filter out those methods in a game implementation so that the only remaining methods are the RPC
	// method handlers.
	for i := 0; i < gmiType.NumMethod(); i++ {
		method := gmiType.Method(i)
		gmiMethods[method.Name] = true
	}

	gameManager := GameManager{
		GameProvider: gp,
		RPCNamespace: gp.RPCNamespace(),
	}

	// Find the complete params method.
	cm, found := gmType.MethodByName(GAME_METHOD_NAME_COMPLETE)
	if !found {
		log.Fatalf("Could not find a method in %s game manager with a name of %s.", gp.Name(), GAME_METHOD_NAME_COMPLETE)
	}

	// 3 Arguments plus the receiver.
	if cm.Type.NumIn() != 4 {
		log.Fatalf("Method %s of %s must have exactly three arguments.", GAME_METHOD_NAME_COMPLETE, gp.Name())
	}

	// Validate that the first argument is the game.
	gmArg := cm.Type.In(1)
	if gmArg != typeOfRepositoryGame {
		log.Fatalf("First parameter of %s in %s must be of type %s.%s", GAME_METHOD_NAME_COMPLETE, gp.Name(), typeOfRepositoryGame.PkgPath(), typeOfRepositoryGame.Name())
	}

	// Validate that the second argument is the bot.
	bArg := cm.Type.In(2)
	if bArg != typeOfRepositoryBot {
		log.Fatalf("Second parameter of %s in %s must be of type %s.%s", GAME_METHOD_NAME_COMPLETE, gp.Name(), typeOfRepositoryBot.PkgPath(), typeOfRepositoryBot.Name())
	}

	// The game state type needs to be consistent for all calls so assume this is the one that we want to use and
	// all others will be compared against it.
	gameManager.GameStateArgType = cm.Type.In(3).Elem()

	// 2 return arguments
	if cm.Type.NumOut() != 2 {
		log.Fatalf("Method %s of %s must return exactly two arguments.", GAME_METHOD_NAME_COMPLETE, gp.Name())
	}

	gameManager.CompleteParamsMethod = cm

	// Find the error params method.
	em, found := gmType.MethodByName(GAME_METHOD_NAME_ERROR)
	if !found {
		log.Fatalf("Could not find a method in %s game manager with a name of %s.", gp.Name(), GAME_METHOD_NAME_ERROR)
	}

	// 2 Arguments plus the receiver.
	if cm.Type.NumIn() != 2 {
		log.Fatalf("Method %s of %s must have exactly two arguments.", GAME_METHOD_NAME_ERROR, gp.Name())
	}

	// Validate that the first argument is the game.
	gmArg = em.Type.In(1)
	if gmArg != typeOfRepositoryGame {
		log.Fatalf("First parameter of %s in %s must be of type %s.%s", GAME_METHOD_NAME_ERROR, gp.Name(), typeOfRepositoryGame.PkgPath(), typeOfRepositoryGame.Name())
	}

	// Validate that the second argument is a string for the error message.
	gmArg = em.Type.In(1)
	if gmArg.Kind() != reflect.String {
		log.Fatalf("Second parameter of %s in %s must be of type string", GAME_METHOD_NAME_ERROR, gp.Name())
	}

	// 2 return arguments
	if em.Type.NumOut() != 2 {
		log.Fatalf("Method %s of %s must return exactly two arguments.", GAME_METHOD_NAME_ERROR, gp.Name())
	}

	gameManager = em

	// Iterate over all the methods in our specific game manager.
	for i := 0; i < gmType.NumMethod(); i++ {
		method := gmType.Method(i)

		// If the found method name is in the list of methods obtained from the GameManager interface then
		// ignore it.
		if gmiMethods[method.Name] {
			continue
		}

		// Ignore our special case complete and error methods that get the params for the Complete and Error
		// RPC calls.
		if method.Name == GAME_METHOD_NAME_COMPLETE || method.Name == GAME_METHOD_NAME_ERROR {
			continue
		}

		// Ignore un-exported methods.
		if method.PkgPath != "" {
			continue
		}

		// Add the raw method name to our set. Raw means without the "RequestParams" or "ProcessResponse"
		// suffix on the method name.
		methodName := r.Replace(method.Name)
		grms.Add(methodName)
	}

	// For each raw method name in our set validate that both the RequestParams and ProcessResponse variants exist
	// and that they have the right arguments and return arguments.
	rpcMethods := make(map[string]gameManagerRPCMethod)
	for i, _ := range grms.set {
		gmrm := gameManagerRPCMethod{}
		gmrm.MethodName = i

		// Find the method for obtaining the request params.
		requestMethodName := fmt.Sprintf("%s%s", i, GAME_METHOD_NAME_SUFFIX_REQUEST)
		requestMethod, found := gmType.MethodByName(requestMethodName)
		if !found {
			log.Fatalf("Could not find a method in %s game manager with a name of %s.", gp.Name(), requestMethodName)
		}

		gmrm.RequestParamsMethod = requestMethod

		// 2 Arguments plus the receiver.
		if requestMethod.Type.NumIn() != 3 {
			log.Fatalf("Method %s of %s must have exactly two arguments.", requestMethodName, gp.Name())
		}

		// Validate that the first argument is the game move.
		gmArg := requestMethod.Type.In(1)
		if gmArg != typeOfRepositoryGameMove {
			log.Fatalf("First parameter of %s in %s must be of type %s.%s", requestMethodName, gp.Name(), typeOfRepositoryGameMove.PkgPath(), typeOfRepositoryGameMove.Name())
		}

		// Validate that the second argument shares the expected game state type.
		if requestMethod.Type.In(2).Elem() != gameManager.GameStateArgType {
			log.Fatalf("Second parameter of %s in %s must be of the same type as specified in the third argument of %s.", requestMethodName, gp.Name(), GAME_METHOD_NAME_COMPLETE)
		}

		// 2 return arguments
		if requestMethod.Type.NumOut() != 2 {
			log.Fatalf("Method %s of %s must return exactly two arguments.", requestMethodName, gp.Name())
		}

		// Find the method for processing the response.
		responseMethodName := fmt.Sprintf("%s%s", i, GAME_METHOD_NAME_SUFFIX_RESPONSE)
		responseMethod, found := gmType.MethodByName(responseMethodName)
		if !found {
			log.Fatalf("Could not find a method in %s game manager with a name of %s.", gp.Name(), responseMethodName)
		}

		gmrm.ProcessResponseMethod = responseMethod

		// 4 arguments plus the receiver
		if responseMethod.Type.NumIn() != 4 {
			log.Fatalf("Method %s of %s must have exactly four arguments.", responseMethodName, gp.Name())
		}

		// Validate that the first argument is the game move.
		gmArg = responseMethod.Type.In(1)
		if gmArg != typeOfRepositoryGameMove {
			log.Fatalf("First parameter of %s in %s must be of type %s.%s", responseMethodName, gp.Name(), typeOfRepositoryGameMove.PkgPath(), typeOfRepositoryGameMove.Name())
		}

		// Validate that the second argument is the bot.
		gmArg = responseMethod.Type.In(2)
		if gmArg != typeOfRepositoryBot {
			log.Fatalf("Second parameter of %s in %s must be of type %s.%s", responseMethodName, gp.Name(), typeOfRepositoryBot.PkgPath(), typeOfRepositoryBot.Name())
		}

		gmrm.ProcessResponseResultArgType = requestMethod.Type.In(3).Elem()

		// 3 return arguments
		if responseMethod.Type.NumOut() != 4 {
			log.Fatalf("Method %s of %s must return exactly three arguments.", responseMethodName, gp.Name())
		}

		rpcMethods[i] = gmrm
	}

	// Validate that the CompleteRequestParams method can be found.
	completeMethod, found := gmType.MethodByName(GAME_METHOD_NAME_COMPLETE)
	if !found {
		log.Fatalf("Could not find a method in %s game manager with a name of %s.", gp.Name(), GAME_METHOD_NAME_COMPLETE)
	}

	// Validate that the CompleteRequestParams accepts the right arguments.
	completeMethodType := completeMethod.Type
	// Method needs one ins: receiver.
	if completeMethodType.NumIn() != 1 {
		log.Fatalf("%s method in %s game manager should take exactly %d arguments", GAME_METHOD_NAME_COMPLETE, gp.Name(), 1)
	}

	// Validate that the ErrorRequestParams method can be found.
	errorMethod, found := gmType.MethodByName(GAME_METHOD_NAME_ERROR)
	if !found {
		log.Fatalf("Could not find a method in %s game manager with a name of %s.", gp.Name(), GAME_METHOD_NAME_ERROR)
	}

	// Validate that the ErrorRequestParams accepts the right arguments.
	errorMethodType := errorMethod.Type
	// Method needs one ins: receiver.
	if errorMethodType.NumIn() != 1 {
		log.Fatalf("%s method in %s game manager should take exactly %d arguments", GAME_METHOD_NAME_ERROR, gp.Name(), 1)
	}

	gameManager.RPCMethods = rpcMethods

	registeredGameManagers[gp.RPCNamespace()] = gameManager

	return nil
}

func GetGameManagerConfigByMnemonic(mnemonic string) (*GameManager, error) {
	for _, gm := range registeredGameManagers {
		if gm.GameProvider.Mnemonic() == mnemonic {
			return &gm, nil
		}
	}

	return nil, errors.New("Unknown game type.")
}

func (gm *GameManager) GetRPCRequestParams(rpcMethodName string, move repository.GameMove) (params interface{}, err error) {

	method, err := getMethod(gm, rpcMethodName)
	if err != nil {
		return nil, err
	}

	gameState, err := getGameStateFromGameMove(move, gm.GameStateArgType)
	if err != nil {
		return nil, err
	}

	args := make([]reflect.Value, 3)
	args[0] = reflect.ValueOf(gm.GameProvider)
	args[1] = reflect.ValueOf(&move)
	args[2] = reflect.ValueOf(gameState)

	fmt.Println("Invoking %s", method.RequestParamsMethod)

	result := method.RequestParamsMethod.Func.Call(args)

	// result[1] error
	returnErrValue := reflect.ValueOf(result[1])
	if returnErrValue.Type().Implements(typeOfErrorInterface) {
		returnErr := returnErrValue.Interface().(error)
		return nil, returnErr
	}

	return reflect.ValueOf(result[0]), nil
}

func (gm GameManager) ProcessRPCResponse(rpcMethodName string, move repository.GameMove, resultJson json.RawMessage) (nextRPCMethodName string, nextMove repository.GameMove, err error) {
	method, err := getMethod(gm, rpcMethodName)
	if err != nil {
		return "", nil, err
	}

	gameState, err := getGameStateFromGameMove(move, gm.GameStateArgType)
	if err != nil {
		return "", nil, err
	}

	//gb, err := move.GameBot()
	//if err != nil {
	//	em := fmt.Sprintf("Unable to game bot for game move %d", move.Id)
	//	return "", nil, errors.New(em)
	//}

	//bot, err := gb.Bot()
	//if err != nil {
	//	em := fmt.Sprintf("Unable to obtain bot for game move %d", move.Id)
	//	return "", nil, errors.New(em)
	//}

	resultValue := reflect.New(method.ProcessResponseResultArgType)
	rpcResult := resultValue.Interface()

	err = json.Unmarshal(resultJson, &rpcResult)
	if err != nil {
		em := fmt.Sprintf("Unable to unmarshall result from RPC call for game move %d", move.Id)
		return "", nil, errors.New(em)
	}

	args := make([]reflect.Value, 4)
	args[0] = reflect.ValueOf(gm.GameProvider)
	args[1] = reflect.ValueOf(&move)
	args[2] = reflect.ValueOf(&rpcResult)
	args[3] = reflect.ValueOf(gameState)

	result := method.ProcessResponseMethod.Func.Call(args)

	// result[0] GameResult
	grValue := reflect.ValueOf(result[0])
	gr := grValue.Interface().(GameResult)

	// result[1] next RPC method name
	nextRpcMethodNameValue := reflect.ValueOf(result[1])
	nextRpcMethodName := nextRpcMethodNameValue.Interface().(string)

	// result[2] game state.
	newGameState := reflect.ValueOf(result[2])

	// result[3] error
	returnErrValue := reflect.ValueOf(result[3])
	if returnErrValue.Type().Implements(typeOfErrorInterface) {
		returnErr := returnErrValue.Interface().(error)
		return "", nil, returnErr
	}

	nextMove, err = gm.ProcessGameResult(gr, move)
	if err != nil {
		return "", nil, err
	}

	err = move.SetGameState(newGameState)
	if err != nil {
		log.Printf("Error setting game state (game move id: %d):\n%v\n", move.Id, err)
		return "", nil, err
	}

	err = move.MarkComplete()
	if err != nil {
		log.Printf("Error marking game move as complete (game move id: %d):\n%v\n", move.Id, err)
		return "", nil, err
	}

	return nextRpcMethodName, nextMove, nil
}

func (gm GameManager) ProcessGameResult(gameResult GameResult, move repository.GameMove) (repository.GameMove, error) {
	if gameResult == GAME_RESULT_UNDECIDED {
		nextBot, err := gm.GameProvider.NextPlayer(move)
		if err != nil {
			log.Printf("Error obtaining game bot for next move (game move id: %d):\n%v\n", err, move.Id)
			return repository.GameMove{}, err
		}

		gs, err := getGameStateFromGameMove(move, gm.GameStateArgType)
		if err != nil {
			log.Printf("Error obtaining game state (current game move id: %d, next game bot id: %d):\n%v\n", move.Id, nextBot.Id, err)
			return repository.GameMove{}, err
		}

		nextMove, err := repository.CreateGameMove(nextBot, gs)
		if err != nil {
			log.Printf("Error creating next game move (current game move id: %d, next game bot id: %d):\n%v\n", move.Id, nextBot.Id, err)
			return repository.GameMove{}, err
		}

		return nextMove, nil
	} else {
		if gameResult == GAME_RESULT_WIN {
			err := move.MarkAsWin()
			if err != nil {
				log.Printf("Error marking game move as win (game move id: %d):\n%v\n", move.Id, err)
			}
		}

		gb, err := move.GameBot()
		if err != nil {
			log.Printf("Error determining game bot (game move id: %d):\n%v\n", move.Id, err)
		}

		game, err := gb.Game()
		if err != nil {
			log.Printf("Error determining game (game move id: %d):\n%v\n", move.Id, err)
		}

		err = game.MarkComplete()
		if err != nil {
			log.Printf("Error marking game as complete (game id: %d):\n%v\n", game.Id, err)
		}

		return repository.GameMove{}, nil
	}
}

func (gm GameManager) NotifyComplete(game repository.Game, bot repository.Bot) {

	lm, err := game.LatestMove()
	if err != nil {
		return
	}

	gs, err := getGameStateFromGameMove(lm, gm.GameStateArgType)

	args := make([]reflect.Value, 4)
	args[0] = reflect.ValueOf(gm.GameProvider)
	args[1] = reflect.ValueOf(&game)
	args[2] = reflect.ValueOf(&bot)
	args[3] = reflect.ValueOf(gs)

	rmn := fmt.Sprintf("%s.%s", gm.RPCNamespace, RPC_METHOD_NAME_COMPLETE)

	result := gm.CompleteParamsMethod.Func.Call(args)

	// result[0] request params
	paramsValue := reflect.ValueOf(result[0])

	// result[1] error
	returnErrValue := reflect.ValueOf(result[1])
	if returnErrValue.Type().Implements(typeOfErrorInterface) {
		returnErr := returnErrValue.Interface().(error)
		return "", nil, returnErr
	}

	rpchelper.Call(bot.RPCEndpoint, rmn, paramsValue)
}

func (gm GameManager) NotifyError() {
	return
}

func getGameStateFromGameMove(gm repository.GameMove, gameStateType reflect.Type) (gameState interface{}, err error) {
	gsString, err := gm.GameStateString()
	if err != nil {
		em := fmt.Sprintf("Unable to retrieve game state for game move %d", gm.Id)
		return nil, errors.New(em)
	}

	gsValue := reflect.New(gameStateType)
	gs := gsValue.Interface()

	err = json.Unmarshal([]byte(gsString), &gs)
	if err != nil {
		em := fmt.Sprintf("Unable to unmarshall game state for game move %d", gm.Id)
		return nil, errors.New(em)
	}

	return gs, nil
}

func GetGameManager(rpcNamespace string) (gm GameManager, err error) {
	gm, found := registeredGameManagers[rpcNamespace]
	if !found {
		em := fmt.Sprintf("Unable to find game manager for RPC namespace %s", rpcNamespace)
		return GameManager{}, errors.New(em)
	}

	return gm, nil
}

func getMethod(gm GameManager, rpcMethodName string) (method gameManagerRPCMethod, err error) {
	method, found := gm.RPCMethods[rpcMethodName]
	if !found {
		em := fmt.Sprintf("Unable to find game manager for RPC method %s in RPC method %s.%s.", rpcMethodName, gm.RPCNamespace(), rpcMethodName)
		return gameManagerRPCMethod{}, errors.New(em)
	}

	return method, nil
}
