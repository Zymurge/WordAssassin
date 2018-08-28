package types

import (
	"sort"
	"fmt"
	"time"

	events "wordassassin/types/events"
)

// GameStatus allows management of game state
type GameStatus string

// Game - The elements necessary to track a single game of wordassassin.
type Game struct {
	ID             string     `json:"id" bson:"_id"`
	TimeCreated    time.Time  `json:"timeCreated" bson:"timecreated"`
	GameCreator    string     `json:"gameId" bson:"gameid"`
	KillDictionary string     `json:"name" bson:"name"`
	Passcode       string     `json:"passcode" bson:"passcode"`
	Status         GameStatus `json:"status" bson:"status"`
	StartTime      time.Time  `json:"starttime"`
	StartPlayers   int        `json:"startplayers"`
	RemainPlayers  int        `json:"remainplayers"`
	// Other possible things:
	//	TargetList
	//	NumKills
}

// Constants for GameStatus
const (
	Starting GameStatus = "starting"
	Playing  GameStatus = "playing"
	Finished GameStatus = "finished"
	Aborted  GameStatus = "aborted"
)

// NewGameFromEvent instantiates a Player from a PlayerAdedEvent
func NewGameFromEvent(ev events.GameCreatedEvent) (g Game) {
	g = Game{
		ID:             ev.ID,
		TimeCreated:    ev.TimeCreated,
		GameCreator:    ev.GameCreator,
		KillDictionary: ev.KillDictionary,
		Passcode:       ev.Passcode,
		Status:         Starting,
	}
	return
}

// GetID getter for ID field
func (g *Game) GetID() string {
	return g.ID
}

// GetStatus translates the status enum
func (g *Game) GetStatus() string {
	// TODO: map to string values of enum
	return string(g.Status)
}

// GetStatusReport generates a status report for this game instance
func (g *Game) GetStatusReport() string {
	result := 
		fmt.Sprintf("Game Status for %s:\n\n", g.GetID()) +
		fmt.Sprintf("   Status: %s\n", 	       g.GetStatus()) +
		fmt.Sprintf("   # Players: %d\n",      g.StartPlayers)

	return result
}

// Start does whatever is needed to transition from setup to go time
func (g *Game) Start() error {
	// Validate whatever needs validating
	// Set the game status to "running"
	// Assign first round of targets
	// Log what you gotta log
	return fmt.Errorf("Not implemented")
}

/*** End Game ***/

// GamePool manages the collection of games in a running server
type GamePool struct {
	games map[string]*Game
}

// AddGame adds a game to this pool. Enforces uniqueness of the Game.ID within the pool
func (pool *GamePool) AddGame(game *Game) error {
	// create the players map as a singleton
	if pool.games == nil {
		pool.games = make(map[string]*Game, 10)
	}
	if game.GetID() == "" {
		return fmt.Errorf("missing ID for AddGame")
	}
	if _, exists := pool.games[game.GetID()]; exists {
		return fmt.Errorf("duplicate ID on add: %s", game.GetID())
	}
	pool.games[game.GetID()] = game
	return nil
}

// GetGame gets the game specified by the requested ID.
// Returns:
// -- the game object for that ID
// -- true for exists if it does, false if it don't
func (pool *GamePool) GetGame(id string) (result *Game, exists bool) {
	result, exists = pool.games[id]
	return
}

// GetGamesList gives a list of each game ID separated by a newline. The result are sorted chronologically by created time
func (pool *GamePool) GetGamesList() (result []*Game) {
	///result := []*Game{}
	for _,v := range pool.games {
		result = append(result, v)
	}
	sort.SliceStable(result, 
		func(i, j int) bool { 
			return result[i].TimeCreated.Before(result[j].TimeCreated) } )
		//	return result[i].GameCreator < result[j].GameCreator } )
	return
}
