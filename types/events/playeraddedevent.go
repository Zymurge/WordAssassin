package events

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
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

// Decode populates this instance from the supplied bson
func (e *PlayerAddedEvent) Decode(raw []byte) error {
	if err := bson.Unmarshal(raw, e); err != nil {
		return err
	}
	return nil
}