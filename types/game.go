package types

import (
	"fmt"
	"time"

	events "wordassassin/types/events"
)

// Game - The elements necessary to track a single game of wordassassin.
type Game struct {
	ID         		string	 	`json:"id" bson:"_id"`
	TimeCreated    	time.Time	`json:"timeCreated" bson:"timecreated"`
	GameCreator		string	  	`json:"gameId" bson:"gameid"`
	KillDictionary  string		`json:"name" bson:"name"`
	Passcode		string		`json:"passcode" bson:"passcode"`
	Status			string		`json:"status" bson:"status"`
// Other possible things:
//	NumPlayers
//	TargetList
//	NumKills
}

// Constants for Game.Status
const (
	GameStatusStarting string = "starting"
	GameStatusPlaying  string = "playing"
	GameStatusFinished string = "finished"
	GameStatusAborted  string = "aborted"

	PlayerStatusAlive  string = "alive"
	PlayerStatusDead   string = "dead"
)

// NewGameFromEvent instantiates a Player from a PlayerAdedEvent
func NewGameFromEvent(ev events.GameCreatedEvent) (g Game) {
	g = Game{
		ID:				ev.ID,
		TimeCreated:	ev.TimeCreated,
		GameCreator:	ev.GameCreator,
		KillDictionary:	ev.KillDictionary,
		Passcode:		ev.Passcode,
		Status:			GameStatusStarting,
	}
	return
}

// GetID getter for ID field
func (g *Game) GetID() string {
	return g.ID
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
