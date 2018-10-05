package types

import (
	"fmt"
	"time"
	"github.com/mongodb/mongo-go-driver/bson/bsoncodec"	
	"wordassassin/types/events"
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
		StartTime:		time.Unix(0, 0),
	}
	return
}

// Decode populates this instance from the supplied bson
func (e *Game) Decode(raw []byte) error {
	if err := bsoncodec.Unmarshal(raw, e); err != nil {
		return err
	}
	return nil
}

// GetID getter for ID field
func (g Game) GetID() string {
	return g.ID
}

// GetStatus translates the status enum
func (g Game) GetStatus() string {
	// TODO: map to string values of enum
	return string(g.Status)
}

// GetStatusReport generates a status report for this game instance
func (g Game) GetStatusReport() string {
	result := 
		fmt.Sprintf("Game Status for %s:\n\n", g.GetID()) +
		fmt.Sprintf("   Status: %s\n", 	       g.GetStatus()) +
		fmt.Sprintf("   # Players: %d\n",      g.StartPlayers)

	return result
}

// Start does whatever is needed to transition from setup to go time
func (g Game) Start() error {
	// Validate whatever needs validating
	// Set the game status to "running"
	// Assign first round of targets
	// Log what you gotta log
	return fmt.Errorf("Not implemented")
}