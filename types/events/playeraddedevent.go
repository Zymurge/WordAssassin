package events

import (
	"fmt"
	"time"

	"wordassassin/slack"
	"gopkg.in/mgo.v2/bson"
)

// PlayerAddedEvent is created for each time a player is added to the game
type PlayerAddedEvent struct {
	ID          string        `json:"id" bson:"_id"`
	TimeCreated time.Time     `json:"timeCreated" bson:"timecreated"`
	EventType	string	      `json:"eventType" bson:"eventtype"`
	GameID		string	      `json:"gameId" bson:"gameid"`
	SlackID     slack.SlackID `json:"slackId" bson:"slackid"`
	Name        string        `json:"name" bson:"name"`
	Email       string        `json:"email" bson:"email"`
}

// NewPlayerAddedEvent returns an instance of the event, along with an automagically calculated ID
// Errors:
// -- either gameid or slackid is blank
// -- slackid is an invalid slack id (per slack validator)
func NewPlayerAddedEvent(gameid, slackid, name, email string) (result PlayerAddedEvent, err error) {
	var actualSlackID slack.SlackID
	if gameid == "" {
		err = fmt.Errorf("The request is missing GameID field")
	} else if slackid == "" {
		err = fmt.Errorf("The request is missing SlackID field")
	} else if actualSlackID, err = slack.New(slackid); err != nil {
		err = fmt.Errorf("Player does not have a valid Slack ID: %v", err)
	}

	result = PlayerAddedEvent{
		TimeCreated: time.Now(),
		EventType:	 "PlayerAddedEvent",
		GameID:      gameid,
		SlackID:     actualSlackID,
		Name:        name,
		Email:       email,
	}

	result.ID = gameid + "+" + slackid
	return
}

// NewPlayerAddedInline returns an instance of the event with no error value. Panics on error instead.
func NewPlayerAddedInline(gameid, slackid, name, email string) PlayerAddedEvent {
	if result, err := NewPlayerAddedEvent(gameid, slackid, name, email); err != nil {
		panic(err)
	} else {
		return result
	}
}

// Decode populates this instance from the supplied bson
func (e *PlayerAddedEvent) Decode(raw []byte) error {
	if err := bson.Unmarshal(raw, e); err != nil {
		return err
	}
	return nil
}

// GetID returns the unique identifer for this event
func (e *PlayerAddedEvent) GetID() string {
	return e.ID
}

// GetTimeCreated returns the unique identifer for this event
func (e *PlayerAddedEvent) GetTimeCreated() time.Time {
	return e.TimeCreated
}