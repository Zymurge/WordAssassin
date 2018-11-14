package types

import (
//	"github.com/mongodb/mongo-go-driver/bson"
	"fmt"
	"sort"
	persistence "wordassassin/persistence"
)

const (
	// GamesCollection const for the mongo collection to hold all game records
	GamesCollection    string = "games"
)

// GamePool manages the collection of games in a running server
type GamePool struct {
	games 	 map[string]Game
	mongo 	 persistence.MongoAbstraction
	players	 PlayerPoolAbstraction
}

// NewGamePool creates an instance with an initialized pool and pointer to the persistence layer
func NewGamePool(m persistence.MongoAbstraction, pp PlayerPoolAbstraction) (result *GamePool) {
	games := make(map[string]Game, 10)
	result = &GamePool{
		games:	games,
		mongo:	m,
	}
	result.players = pp

	// Reconstitute from mongo automatically
	// TODO: better error and logging on reconstitution issues
	existingBytes, err := result.mongo.FetchAllFromCollection(GamesCollection)
	if err != nil { panic(err) }
	gamesList := bytesToGames(existingBytes)
	err = result.ReconstitutePool(gamesList)
	if err != nil { panic(err) }
	fmt.Printf("NewGamePool: Restored %d games\n", len(result.games))
	return
}

// AddGame adds a game to this pool and persists the addition. Enforces uniqueness of the Game.ID within the pool
func (pool *GamePool) AddGame(game Game) error {
	if pool.games == nil {
		return fmt.Errorf("uninitialized pool. Use NewGamePool")
	}
	if game.GetID() == "" {
		return fmt.Errorf("missing ID for AddGame")
	}
	if err := pool.addGameToMap(game); err != nil {
		return err
	}
	if err := pool.persistGame(&game); err != nil {
		return err
	}
	return nil
}

// GetGame gets the game specified by the requested ID.
// Returns:
// -- the game object for that ID
// -- true for exists if it does, false if it don't
func (pool *GamePool) GetGame(id string) (*Game, bool) {
	game, exists := pool.games[id]
	if exists {
		return &game, true
	}
	return nil, false
}

// GetGamesList gives a list of each game ID separated by a newline. The result are sorted chronologically by created time
func (pool *GamePool) GetGamesList() (result []Game) {
	for _, v := range pool.games {
		result = append(result, v)
	}
	sort.SliceStable(result,
		func(i, j int) bool {
			return result[i].TimeCreated.Before(result[j].TimeCreated)
		})
	//	return result[i].GameCreator < result[j].GameCreator } )
	return
}

// ReconstitutePool rebuilds a new GamePool from an array of Games
func (pool *GamePool) ReconstitutePool(games []Game) error {
	for _, game := range games {
		if err := pool.addGameToMap(game); err != nil {
			return err
		}
	}
	return nil
}

// StartGame calls the start sequence for the specified game on behalf of requestor.
// Only the original game creator is allowed to start a given gameid.
// Errors returned:
// -- gameid or slackid empty
// -- gameid not exists and in 'starting' state
// -- slackid does not match the creating slackid
func (pool *GamePool) StartGame(gameid string, slackid string) (err error) {
	if gameid == "" || slackid == "" {
		err = fmt.Errorf("Game start requires a non-empty game ID and slack ID")
		return
	}
	game, exists := pool.GetGame(gameid)
	if !exists {
		err = fmt.Errorf("The requested GameID: %s doesn't exist on this server", gameid)
		return
	}
	if game.Status != Starting {
		err = fmt.Errorf("The requested GameID: %s is not accepting players. State=%s", gameid, game.Status)
		return
	}
	if game.GameCreator != slackid {
		err = fmt.Errorf("GameID: %s cannot be started by non-creator. %s tried though", gameid, slackid)
		return
	}

	// TODO: get list of players for this game
	players := []*Player{}
	return game.Start(players)
}

func (pool *GamePool) addGameToMap(game Game) error {
	if _, exists := pool.games[game.GetID()]; exists {
		return fmt.Errorf("duplicate ID on add: %s", game.GetID())
	}
	pool.games[game.GetID()] = game
	return nil
}

// turn a bson array of bytes into an array of Game instances
func bytesToGames(inBytes [][]byte) []Game {
    ret := make([]Game, len(inBytes))

    for i, b := range inBytes {
        if err := ret[i].Decode(b); err != nil {
			panic( "bytesToGames: " + err.Error())
		}
    }

    return ret
}

func (pool *GamePool) persistGame(game *Game) error {
	if mongoErr := pool.mongo.WriteCollection("games", game); mongoErr != nil {
		return mongoErr
	}

	return nil
}

// Get all of the players for a given gameid. 
func (pool *GamePool) playersInGame(gameid string) (playersInGame []*Player, err error) {
	return pool.players.GetAllPlayersInGame(gameid)
}
