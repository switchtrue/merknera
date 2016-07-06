package games

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/mleonard87/merknera/repository"
)

type GameResult string

const (
	GAME_RESULT_WIN       GameResult = "WIN"
	GAME_RESULT_DRAW      GameResult = "DRAW"
	GAME_RESULT_UNDECIDED GameResult = "UNDECIDED"
	GAME_RESULT_ERROR     GameResult = "ERROR"

	GAME_METHOD_NAME_SUFFIX_REQUEST  = "RequestParams"
	GAME_METHOD_NAME_SUFFIX_RESPONSE = "ProcessResponse"

	GAME_METHOD_NAME_COMPLETE = "CompleteRequestParams"
	GAME_METHOD_NAME_ERROR    = "ErrorRequestParams"
)

type GameProvider interface {
	Mnemonic() string
	Name() string
	RPCNamespace() string

	GetGamesForBot(bot repository.Bot, otherBots []repository.Bot) [][]*repository.Bot

	Begin(game repository.Game) (rpcMethod string, initialPlayer repository.GameBot, gameState interface{}, err error)
	Resume(game repository.Game) (rpcMethod string, err error)
}

type GameManager struct {
	GameManager  GameProvider
	RPCNamespace string
	RPCMethods   map[string]gameManagerRPCMethod
}

type gameManagerRPCMethod struct {
	MethodName                   string
	RequestParamsMethod          reflect.Method // Method to be invoked to get the RPC request params.
	ProcessResponseMethod        reflect.Method // Method to be invoked to process the result of the RPC request.
	GameStateArgType             reflect.Type   // Type of the game state arugment used in both RequestParams and ProcessResponse.
	RequestParamsReturnType      reflect.Type   // Type of request params to be used for the RPC call.
	ProcessResponseResultArgType reflect.Type   // Type to unmarshal the RPC results into for the Process response argument.
}

var (
	// Pre-compute the reflect.Type of error and repository.GameMove
	typeOfRepositoryGameMove = reflect.TypeOf((*repository.GameMove)(nil))
	typeOfRepositoryBot      = reflect.TypeOf((*repository.Bot)(nil))
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

func RegisterGameManager(gm GameProvider) error {
	_, err := repository.GetGameTypeByMnemonic(gm.Mnemonic())
	if err != nil {
		_, err := repository.CreateGameType(gm.Mnemonic(), gm.Name())
		if err != nil {
			return err
		}
	}

	gmType := reflect.TypeOf(gm)

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

	// For each raw method name in our set validate that both the RequestParams and ProcessResponse varients exist
	// and that they have the right arguments and return arguments.
	rpcMethods := make(map[string]gameManagerRPCMethod)
	for i, _ := range grms.set {
		gmrm := gameManagerRPCMethod{}
		gmrm.MethodName = i

		// Find the method for obtaining the request params.
		requestMethodName := fmt.Sprintf("%s%s", i, GAME_METHOD_NAME_SUFFIX_REQUEST)
		requestMethod, found := gmType.MethodByName(requestMethodName)
		if !found {
			log.Fatalf("Could not find a method in %s game manager with a name of %s.", gm.Name(), requestMethodName)
		}

		gmrm.RequestParamsMethod = requestMethod

		// 2 Arguments plus the receiver.
		if requestMethod.Type.NumIn() != 3 {
			log.Fatalf("Method %s of %s must have exactly two arguments.", requestMethodName, gm.Name())
		}

		// Validate that the first argument is the game move.
		gmArg := requestMethod.Type.In(1)
		if gmArg != typeOfRepositoryGameMove {
			log.Fatalf("First parameter of %s in %s must be of type %s.%s", requestMethodName, gm.Name(), typeOfRepositoryGameMove.PkgPath(), typeOfRepositoryGameMove.Name())
		}

		// The second argument can be any type as its the game state so just store it.
		gmrm.GameStateArgType = requestMethod.Type.In(2).Elem()

		// 2 return arguments
		if requestMethod.Type.NumOut() != 2 {
			log.Fatalf("Method %s of %s must return exactly two arguments.", requestMethodName, gm.Name())
		}

		// Find the method for processing the response.
		responseMethodName := fmt.Sprintf("%s%s", i, GAME_METHOD_NAME_SUFFIX_RESPONSE)
		responseMethod, found := gmType.MethodByName(responseMethodName)
		if !found {
			log.Fatalf("Could not find a method in %s game manager with a name of %s.", gm.Name(), responseMethodName)
		}

		gmrm.ProcessResponseMethod = responseMethod

		// 4 arguments plus the receiver
		if responseMethod.Type.NumIn() != 4 {
			log.Fatalf("Method %s of %s must have exactly four arguments.", responseMethodName, gm.Name())
		}

		// Validate that the first argument is the game move.
		gmArg = responseMethod.Type.In(1)
		if gmArg != typeOfRepositoryGameMove {
			log.Fatalf("First parameter of %s in %s must be of type %s.%s", responseMethodName, gm.Name(), typeOfRepositoryGameMove.PkgPath(), typeOfRepositoryGameMove.Name())
		}

		// Validate that the second argument is the bot.
		gmArg = responseMethod.Type.In(2)
		if gmArg != typeOfRepositoryBot {
			log.Fatalf("Second parameter of %s in %s must be of type %s.%s", responseMethodName, gm.Name(), typeOfRepositoryBot.PkgPath(), typeOfRepositoryBot.Name())
		}

		gmrm.ProcessResponseResultArgType = requestMethod.Type.In(3).Elem()

		// 3 return arguments
		if responseMethod.Type.NumOut() != 4 {
			log.Fatalf("Method %s of %s must return exactly three arguments.", responseMethodName, gm.Name())
		}

		rpcMethods[i] = gmrm
	}

	// Validate that the CompleteRequestParams method can be found.
	completeMethod, found := gmType.MethodByName(GAME_METHOD_NAME_COMPLETE)
	if !found {
		log.Fatalf("Could not find a method in %s game manager with a name of %s.", gm.Name(), GAME_METHOD_NAME_COMPLETE)
	}

	// Validate that the CompleteRequestParams accepts the right arguments.
	completeMethodType := completeMethod.Type
	// Method needs one ins: receiver.
	if completeMethodType.NumIn() != 1 {
		log.Fatalf("%s method in %s game manager should take exactly %d arguments", GAME_METHOD_NAME_COMPLETE, gm.Name(), 1)
	}

	// Validate that the ErrorRequestParams method can be found.
	errorMethod, found := gmType.MethodByName(GAME_METHOD_NAME_ERROR)
	if !found {
		log.Fatalf("Could not find a method in %s game manager with a name of %s.", gm.Name(), GAME_METHOD_NAME_ERROR)
	}

	// Validate that the ErrorRequestParams accepts the right arguments.
	errorMethodType := errorMethod.Type
	// Method needs one ins: receiver.
	if errorMethodType.NumIn() != 1 {
		log.Fatalf("%s method in %s game manager should take exactly %d arguments", GAME_METHOD_NAME_ERROR, gm.Name(), 1)
	}

	registeredGameManagers[gm.RPCNamespace()] = GameManager{
		GameManager:  gm,
		RPCNamespace: gm.RPCNamespace(),
		RPCMethods:   rpcMethods,
	}

	//testGm, err := repository.GetGameMoveById(83)
	//if err != nil {
	//	log.Fatal("Unable to find a test game move.")
	//}

	//params, err := GetRPCRequestParams("TicTacToe", "NextMove", testGm)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//fmt.Println("***")
	//fmt.Println(params)
	//fmt.Println("***")

	return nil
}

func GetGameManagerConfigByMnemonic(mnemonic string) (*GameManager, error) {
	for _, gm := range registeredGameManagers {
		if gm.GameManager.Mnemonic() == mnemonic {
			return &gm, nil
		}
	}

	return nil, errors.New("Unknown game type.")
}

func (gm *GameManager) GetRPCRequestParams(rpcMethodName string, move repository.GameMove) (params interface{}, err error) {
	//gameManagerConfig, err := getGameManagerConfig(rpcNamespace)
	//if err != nil {
	//	return nil, err
	//}

	method, err := getMethod(gm, rpcMethodName)
	if err != nil {
		return nil, err
	}

	gameState, err := getGameStateFromGameMove(move, method.GameStateArgType)
	if err != nil {
		return nil, err
	}

	args := make([]reflect.Value, 3)
	args[0] = reflect.ValueOf(gm.GameManager)
	args[1] = reflect.ValueOf(&move)
	args[2] = reflect.ValueOf(gameState)

	fmt.Println("Invoking %s", method.RequestParamsMethod)

	result := method.RequestParamsMethod.Func.Call(args)

	// TODO: figure out how to cast the error and return it from result[1]
	//errValue := reflect.ValueOf(result[0])
	//var returnErr error
	//returnErr = errValue.Interface().(error)

	return reflect.ValueOf(result[0]), nil
}

func (gm GameManager) ProcessRPCResponse(rpcMethodName string, move repository.GameMove, resultJson json.RawMessage) (gameResult GameResult, nextRPCMethodName string, newGameState interface{}, err error) {
	//gameManagerConfig, err := getGameManagerConfig(rpcNamespace)
	//if err != nil {
	//	return GAME_RESULT_ERROR, "", nil, err
	//}

	method, err := getMethod(gm, rpcMethodName)
	if err != nil {
		return GAME_RESULT_ERROR, "", nil, err
	}

	gameState, err := getGameStateFromGameMove(move, method.GameStateArgType)
	if err != nil {
		return GAME_RESULT_ERROR, "", nil, err
	}

	gb, err := move.GameBot()
	if err != nil {
		em := fmt.Sprintf("Unable to game bot for game move %d", move.Id)
		return GAME_RESULT_ERROR, "", nil, errors.New(em)
	}

	bot, err := gb.Bot()
	if err != nil {
		em := fmt.Sprintf("Unable to obtain bot for game move %d", move.Id)
		return GAME_RESULT_ERROR, "", nil, errors.New(em)
	}

	resultValue := reflect.New(method.ProcessResponseResultArgType)
	rpcResult := resultValue.Interface()

	err = json.Unmarshal(resultJson, &rpcResult)
	if err != nil {
		em := fmt.Sprintf("Unable to unmarshall result from RPC call for game move %d", move.Id)
		return GAME_RESULT_ERROR, "", nil, errors.New(em)
	}

	args := make([]reflect.Value, 4)
	args[0] = reflect.ValueOf(gm.GameManager)
	args[1] = reflect.ValueOf(&move)
	args[2] = reflect.ValueOf(&rpcResult)
	args[3] = reflect.ValueOf(gameState)

	result := method.ProcessResponseMethod.Func.Call(args)

	return

	// TODO: figure out how to cast the error and return it from result[1]
	//errValue := reflect.ValueOf(result[0])
	//var returnErr error
	//returnErr = errValue.Interface().(error)

	grValue := reflect.ValueOf(result[0])
	gr := grValue.Interface().(GameResult)

	nextRpcMethodNameValue := reflect.ValueOf(result[0])
	nextRpcMethodName := nextRpcMethodNameValue.Interface().(string)

	return gr, nextRpcMethodName, reflect.ValueOf(result[2]), nil
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
		em := fmt.Sprintf("Unable to find game manager for RPC namespace %s", rpcNamespace, rpcNamespace)
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
