package events

import (
	"fmt"
	"time"

	bson "github.com/globalsign/mgo/bson"
)

// PlayerAddedEvent is created for each time a player is added to the game
type PlayerAddedEvent struct {
	ID          string    `json:"id" bson:"_id"`
	TimeCreated time.Time `json:"timeCreated" bson:"timecreated"`
	EventType	string	  `json:"eventType" bson:"eventtype"`
	GameID		string	  `json:"gameId" bson:"gameid"`
	SlackID     string    `json:"slackId" bson:"slackid"`
	Name        string    `json:"name" bson:"name"`
	Email       string    `json:"email" bson:"email"`
}

// NewPlayerAddedEvent returns an instance of the event, along with an automagically calculated ID
func NewPlayerAddedEvent(gameid, slackid, name, email string) (result PlayerAddedEvent, err error) {
	result = PlayerAddedEvent{
		TimeCreated: time.Now(),
		EventType:	 "PlayerAddedEvent",
		GameID:      gameid,
		SlackID:     slackid,
		Name:        name,
		Email:       email,
	}
	// Parse and validate input
	if gameid == "" {
		err = fmt.Errorf("The request is missing GameID field")
		return
	}
	if slackid == "" {
		err = fmt.Errorf("The request is missing SlackID field")
		return
	}
	result.ID = gameid + "+" + slackid
	return
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
	if val, ok := b["eventtype"]; ok {
		if e.Name, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for EventType")
		} 
	} else {
		return fmt.Errorf("Missing tag: eventType")
	}	
	if val, ok := b["name"]; ok {
		if e.Name, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for Name")
		} 
	} else {
		return fmt.Errorf("Missing tag: name")
	}	
	if val, ok := b["gameid"]; ok {
		if e.GameID, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for GameID")
		} 
	} else {
		return fmt.Errorf("Missing tag: gameid")
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
