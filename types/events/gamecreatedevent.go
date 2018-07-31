package events

import (
	"fmt"
	"time"

	bson "github.com/globalsign/mgo/bson"
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
func (e *GameCreatedEvent) Decode(b bson.M) error {
	if val, ok := b["_id"]; ok {
		if e.ID, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for ID")
		}
	} else {
		return fmt.Errorf("Missing key: _id")
	}
	if val, ok := b["timecreated"]; ok {
		if e.TimeCreated, ok = val.(time.Time); !ok {
			return fmt.Errorf("Cast issue for TimeCreated")
		}
		e.TimeCreated = e.TimeCreated.UTC()
	} else {
		return fmt.Errorf("Missing key: timecreated")
	}
	if val, ok := b["eventtype"]; ok {
		if e.KillDictionary, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for EventType")
		}
	} else {
		return fmt.Errorf("Missing key: eventtype")
	}
	if val, ok := b["gamecreator"]; ok {
		if e.GameCreator, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for GameCreator")
		}
	} else {
		return fmt.Errorf("Missing key: gamecreator")
	}
	if val, ok := b["killdictionary"]; ok {
		if e.KillDictionary, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for KillDictionary")
		}
	} else {
		return fmt.Errorf("Missing key: killdictionary")
	}
	return nil
}
