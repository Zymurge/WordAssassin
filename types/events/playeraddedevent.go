package events

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// PlayerAddedEvent is created for each time a player is added to the game
type PlayerAddedEvent struct {
	//GameEvent
	ID          string    `json:"id" bson:"_id"`
	TimeCreated time.Time `json:"timeCreated" bson:"timecreated"`
	Name        string    `json:"name" bson:"name"`
	SlackID     string    `json:"slackId" bson:"slackid"`
	Email       string    `json:"email" bson:"email"`
}

// GetID returns the unique identifer for this event
func (e *PlayerAddedEvent) GetID() string {
	return e.ID
}

// GetTimeCreated returns the unique identifer for this event
func (e *PlayerAddedEvent) GetTimeCreated() time.Time {
	return e.TimeCreated
}

// SetBSON implements the mgo.bson Setter interface to allow for polymorphic unmarshalling
func (e *PlayerAddedEvent) SetBSON(raw bson.Raw) error {
	if err := raw.Unmarshal(e); err != nil {
		return err
	}
	return nil	
}

// Decode populates this instance from the supplied bson
func (e *PlayerAddedEvent) Decode(b bson.M) error {
	if val, ok := b["_id"]; ok {
		if e.ID, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for ID")
		} 
	} else {
		return fmt.Errorf("Missing tag: _id")
	}
	if val, ok := b["timecreated"]; ok {
		if e.TimeCreated, ok = val.(time.Time); !ok {
			return fmt.Errorf("Cast issue for TimeCreated")
		} 
		e.TimeCreated = e.TimeCreated.UTC()
	} else {
		return fmt.Errorf("Missing tag: timecreated")
	}
	if val, ok := b["name"]; ok {
		if e.Name, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for Name")
		} 
	} else {
		return fmt.Errorf("Missing tag: name")
	}
	if val, ok := b["slackid"]; ok {
		if e.SlackID, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for SlackId")
		} 
	} else {
		return fmt.Errorf("Missing tag: slackid")
	}
	if val, ok := b["email"]; ok {
		if e.Email, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for Email")
		} 
	} else {
		return fmt.Errorf("Missing tag: email")
	}

	return nil
}
