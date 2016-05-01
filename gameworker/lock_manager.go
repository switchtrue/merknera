package gameworker

import (
	"sync"

	"log"

	"github.com/mleonard87/merknera/repository"
)

var lockManagerLock sync.Mutex
var gameMoveLocks map[int]sync.Mutex

func init() {
	lockManagerLock = sync.Mutex{}
	gameMoveLocks = make(map[int]sync.Mutex)
}

func GetGameMoveLock(gm repository.GameMove) {
	lockManagerLock.Lock()
	defer lockManagerLock.Unlock()

	gml, ok := gameMoveLocks[gm.Id]
	if !ok {
		gameMoveLocks[gm.Id] = sync.Mutex{}
		gml = gameMoveLocks[gm.Id]
	}
	gml.Lock()
}

func ReleaseGameMoveLock(gm repository.GameMove) {
	lockManagerLock.Lock()
	defer lockManagerLock.Unlock()

	gml, ok := gameMoveLocks[gm.Id]
	if !ok {
		log.Printf("Error locating mutex for GameMove %d to unlock.", gm.Id)
	}
	gml.Unlock()
}
