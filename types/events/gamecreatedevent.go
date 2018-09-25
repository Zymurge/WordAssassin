package events

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// GameCreatedEvent is created once per game instance
type GameCreatedEvent struct {
	ID             string    `json:"id" bson:"_id"`
	TimeCreated    time.Time `json:"timeCreated"`
	EventType      string    `json:"eventType"`
	GameCreator    string    `json:"gameCreator"`
	KillDictionary string    `json:"killDictionary"`
	Passcode       string    `json:"passcode" bson:"passcode"`
}

// NewGameCreatedEvent returns an instance of the event
func NewGameCreatedEvent(gameid, creator, killdict, passcode string) (result GameCreatedEvent, err error) {
	result = GameCreatedEvent{
		ID:             gameid,
		TimeCreated:    time.Now(),
		EventType:      "GameCreatedEvent",
		GameCreator:    creator,
		KillDictionary: killdict,
		Passcode:       passcode,
	}
	// TODO: validate inputs
	if gameid == "" {
		err = fmt.Errorf("The request is missing GameID field")
		return
	}
	if creator == "" {
		err = fmt.Errorf("The request is missing Creator field")
		return
	}
	return
}

// GetID returns the unique identifer for this event
func (e *GameCreatedEvent) GetID() string {
	return e.ID
}

// GetTimeCreated returns the unique identifer for this event
func (e *GameCreatedEvent) GetTimeCreated() time.Time {
	return e.TimeCreated
}

 // Decode populates this instance from the supplied bson
func (e *GameCreatedEvent) Decode(raw []byte) error {
	if err := bson.Unmarshal(raw, e); err != nil {
		return err
	}
	return nil
}