package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// MongoID reresents an interface for any object storeable in mongo with an explicit ID field
type MongoID interface {
	GetID() string
}

// GameTracker contains the coords and methods to handle a 3 axis location on a hex map
type GameTracker struct {
	ID            string     `json:"id" bson:"_id"`
	Game          string     `json:"game"`
	Status        GameStatus `json:"status"`
	StartTime     time.Time  `json:"starttime"`
	StartPlayers  int        `json:"startplayers"`
	RemainPlayers int        `json:"remainplayers"`
}

// GetID getter for ID field
func (g GameTracker) GetID() string {
	return g.ID
}

// GameTrackerFromJSON generates a GameTracker instance from JSON. Expected JSON form should match the struct declaration. Duh!
func GameTrackerFromJSON(jsonIn []byte) (GameTracker, error) {
	result := GameTracker{}
	if err := json.Unmarshal(jsonIn, &result); err != nil {
		return result, fmt.Errorf("unmarshal error: %s", err.Error())
	}
	if result.ID == "" {
		return result, fmt.Errorf("JSON missing ID field")
	}
	return result, nil
}

// JSONForm provides the location in JSON
func (g GameTracker) JSONForm() []byte {
	j, err := json.Marshal(g)
	if err != nil {
		fmt.Println("bad things happened in JSON marshal")
		panic(err)
	}
	return j
}
