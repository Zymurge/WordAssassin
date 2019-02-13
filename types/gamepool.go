package types

import (
	"fmt"
	"sort"
	"strings"
	
	events "wordassassin/types/events"
	persistence "wordassassin/persistence"
)

// GamePoolAbstraction provides abstraction for testing GamePool dependencies
type GamePoolAbstraction interface {
	AddGame(game *Game) error
	AddPlayerToGame(gameid string, ev events.PlayerAddedEvent) error
	CanAddPlayers(gameid string) (bool, error)
	GetGame(id string) (*Game, bool)
	GetGamesList() []*Game
	StartGame(gameid string, slackid string) error 
}

const (
	// GamesCollection const for the mongo collection to hold all game records
	GamesCollection    string = "games"
)

// GamePool manages the collection of games in a running server
type GamePool struct {
	games 	 map[string]*Game
	mongo 	 persistence.MongoAbstraction
	players	 PlayerPoolAbstraction
}

// NewGamePool creates an instance with an initialized pool and pointer to the persistence layer
func NewGamePool(m persistence.MongoAbstraction, pp PlayerPoolAbstraction) (result *GamePool) {
	games := make(map[string]*Game, 10)
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
	//fmt.Printf("NewGamePool: Restored %d games\n", len(result.games))
	return
}

// AddGame adds a game to this pool and persists the addition. Enforces uniqueness of the Game.ID within the pool
func (pool *GamePool) AddGame(game *Game) error {
	if pool.games == nil {
		return fmt.Errorf("uninitialized pool. Use NewGamePool")
	}
	if game.GetID() == "" {
		return fmt.Errorf("missing ID for AddGame")
	}
	if err := pool.addGameToMap(game); err != nil {
		return err
	}
	if err := pool.persistGame(game); err != nil {
		return err
	}
	return nil
}

// AddPlayerToGame encapsulates whatever needs to happen when associating a new player with a game
// Note: with the current design, that really only means incrementing the player count, since the linkage
// is from the player pool to the actual game, and not bi-directional
func (pool *GamePool) AddPlayerToGame(gameid string, ev events.PlayerAddedEvent) error {
	if accepting, err := pool.CanAddPlayers(gameid); !accepting {
		return err
	}
	game, _ := pool.GetGame(gameid)

	// Create the Player instance and add to the PlayerPool
	player := NewPlayerFromEvent(ev)
	if addErr := pool.players.AddPlayer(&player); addErr != nil {
		// Should catch all dups at the event level
		if strings.Contains(addErr.Error(), "duplicate") {
			return fmt.Errorf("PlayerPool: attempt to add duplicate player: %s in game: %s", player.GetID(), gameid)
			
		}
		return fmt.Errorf("PlayerPool: issue on AddPlayer add to : %s", addErr.Error())
	}
	// FIXME: Persist the pPool, or should it auto persist on state change?

	game.StartPlayers++
	return nil
}

// CanAddPlayers validates that a game exists and is in the proper state to accept new players
func (pool *GamePool) CanAddPlayers(gameid string) (accepting bool, err error) {
	accepting = true
	game, exists := pool.GetGame(gameid)
	if !exists {
		err = fmt.Errorf("The requested GameID: %s doesn't exist on this server", gameid)
		accepting = false
	} else if game.Status != Starting {
		err = fmt.Errorf("The requested GameID: %s is not accepting players. State=%s", gameid, game.Status)
		accepting = false
	}
	return
}


// GetGame gets the game specified by the requested ID.
// Returns:
// -- the game object for that ID
// -- true for exists if it does, false if it don't
func (pool *GamePool) GetGame(id string) (*Game, bool) {
	game, exists := pool.games[id]
	if exists {
		return game, true
	}
	return nil, false
}

// GetGamesList gives a list of each game ID separated by a newline. The result are sorted chronologically by created time
func (pool *GamePool) GetGamesList() (result []*Game) {
	for _, v := range pool.games {
		result = append(result, v)
	}
	sort.SliceStable(result,
		func(i, j int) bool {
			return result[i].TimeCreated.Before(result[j].TimeCreated)
		})
	return
}

// ReconstitutePool rebuilds a new GamePool from an array of Games
func (pool *GamePool) ReconstitutePool(games []*Game) error {
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
// -- gameid or creator empty
// -- gameid not exists and in 'starting' state
// -- slackid does not match the creating slackid
func (pool *GamePool) StartGame(gameid string, creator string) error {
	if gameid == "" || creator == "" {
		return fmt.Errorf("Game start requires a non-empty game ID and creator ID")
	}
	game, exists := pool.GetGame(gameid)
	if !exists {
		return fmt.Errorf("The requested GameID: %s doesn't exist on this server", gameid)
	}
	if game.Status != Starting {
		return fmt.Errorf("The requested GameID: %s is not accepting players. State=%s", gameid, game.Status)
	}
	if game.GameCreator != creator {
		return fmt.Errorf("GameID: %s cannot be started by non-creator. %s tried though", gameid, creator)
	}
	//var players []*Player
	if players, err := pool.players.GetAllPlayersInGame(game.GetID()); err != nil {
		panic("Error on GetAllPlayersInGame")
	} else {
		return game.Start(players)
	}
}

func (pool *GamePool) addGameToMap(game *Game) error {
	if _, exists := pool.games[game.GetID()]; exists {
		return fmt.Errorf("duplicate ID on add: %s", game.GetID())
	}
	pool.games[game.GetID()] = game
	return nil
}

// turn a bson array of bytes into an array of Game instances
func bytesToGames(inBytes [][]byte) []*Game {
    ret := make([]*Game, len(inBytes))

    for i, b := range inBytes {
		ret[i] = &Game{}
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
