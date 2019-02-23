package events

import (
	"fmt"
	"time"

	"wordassassin/slack"
	"github.com/mongodb/mongo-go-driver/bson"
)

// GameCreatedEvent is created once per game instance
type GameCreatedEvent struct {
	ID             string    	 `json:"id" bson:"_id"`
	TimeCreated    time.Time 	 `json:"timeCreated"`
	EventType      string    	 `json:"eventType"`
	GameCreator    slack.SlackID `json:"gameCreator"`
	KillDictionary string    	 `json:"killDictionary"`
	Passcode       string    	 `json:"passcode" bson:"passcode"`
}

// NewGameCreatedEvent returns an instance of the event
// Errors:
// -- either gameid or creator is blank
// -- creator is an invalid slack id (per slack validator)
func NewGameCreatedEvent(gameid, creator, killdict, passcode string) (result GameCreatedEvent, err error) {
	var creatorID slack.SlackID
	if gameid == "" {
		err = fmt.Errorf("The request is missing GameID field")
	} else if creator == "" {
		err = fmt.Errorf("The request is missing Creator field")
	} else if creatorID, err = slack.New(creator); err != nil {
		err = fmt.Errorf("Creator is not a valid Slack ID: %v", err)
	}

	result = GameCreatedEvent{
		ID:             gameid,
		TimeCreated:    time.Now(),
		EventType:      "GameCreatedEvent",
		GameCreator:    creatorID,
		KillDictionary: killdict,
		Passcode:       passcode,
	}

	return
}

// NewGameCreatedInline returns an instance of the event with no error value. Panics on error instead.
func NewGameCreatedInline(gameid, creator, killdict, passcode string) GameCreatedEvent {
	if result, err := NewGameCreatedEvent(gameid, creator, killdict, passcode); err != nil {
		panic(err)
	} else {
		return result
	}
}

// Decode populates this instance from the supplied bson
func (e *GameCreatedEvent) Decode(raw []byte) error {
	if err := bson.Unmarshal(raw, e); err != nil {
		return err
	}
	return nil
}

// GetID returns the unique identifer for this event
func (e *GameCreatedEvent) GetID() string {
	return e.ID
}

// GetTimeCreated returns the unique identifer for this event
func (e *GameCreatedEvent) GetTimeCreated() time.Time {
	return e.TimeCreated
}