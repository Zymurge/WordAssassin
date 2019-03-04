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
// -- either any of the arguments are blank
// -- creator is an invalid slack id (per slack validator)
func NewGameCreatedEvent(gameid string, creator slack.SlackID, killdict, passcode string) (result GameCreatedEvent, err error) {
	if gameid == "" {
		err = fmt.Errorf("The request is missing GameID field")
	} else if creator == "" {
		err = fmt.Errorf("The request is missing Creator field")
	} else if killdict == "" {
		err = fmt.Errorf("The request is missing Killdict field")
	} else if passcode == "" {
		err = fmt.Errorf("The request is missing Passcode field")
	}

	result = GameCreatedEvent{
		ID:             gameid,
		TimeCreated:    time.Now(),
		EventType:      "GameCreatedEvent",
		GameCreator:    creator,
		KillDictionary: killdict,
		Passcode:       passcode,
	}

	return
}

// NewGameCreatedInline is a test util that creates an instance of of a game event. 
// It provides a single receiver form. If any params fails validation, the func panics.
func NewGameCreatedInline(gameid, creator, killdict, passcode string) GameCreatedEvent {
	if result, err := NewGameCreatedEvent(gameid, slack.NewInline(creator), killdict, passcode); err != nil {
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