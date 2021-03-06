package events

import (
	"wordassassin/persistence"
	"time"
)

// GameEvent encapsulates the common features of any event generated in the game
type GameEvent interface {
	persistence.Persistable
	GetTimeCreated() time.Time
}

// GameStartedEvent is created when the game is started
type GameStartedEvent struct {
	GameEvent
	ID          string    `json:"id" bson:"_id"`
	TimeStarted time.Time `json:"timeStarted"`
}

// GameCompletedEvent is created when the game completes
type GameCompletedEvent struct {
	GameEvent
	ID          string    `json:"id" bson:"_id"`
	TimeStarted time.Time `json:"timeCompleted"`
}

// TargetAssignedEvent is created when a target is assigned by the game engine
type TargetAssignedEvent struct {
	GameEvent
	ID            string    `json:"id" bson:"_id"`
	GameID		  string    `json:"gameId" bson:"gameid"`
	TargetID      string    `json:"targetId"`
	KillerID      string    `json:"killerId"`
	KillWord      string    `json:"killword"`
	TimeAssigned  time.Time `json:"timeAssigned"`
	TimeCompleted time.Time `json:"timeCompleted"`
}
