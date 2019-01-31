package types

import (
	"time"

	events "wordassassin/types/events"
)

// PlayerStatus allows management of player state
type PlayerStatus int

// Player - It's what it sounds like.
type Player struct {
	ID          string	 	`json:"id" bson:"_id"`
	TimeCreated time.Time	`json:"timeCreated" bson:"timecreated"`
	GameID		string	  	`json:"gameId" bson:"gameid"`
	Name        string		`json:"name" bson:"name"`
	SlackID     string		`json:"slackId" bson:"slackid"`
	Email       string		`json:"email" bson:"email"`
	Status		PlayerStatus `json:"status" bson:"status"`
	Kills		int			`json:"kills" bson:"kills"`
	Target		string		`json:"target" bson:"target"`
	KillWord	string		`json:"killword" bson:"killword"`
}
	
// Constants for PlayerStatus
const (
	Alive PlayerStatus = iota + 1
	Dead
)

// NewPlayer instantiates a player from the limited fields needed for an event
func NewPlayer(gameid, slackid, name, email string) (p Player, err error) {
	var ev events.PlayerAddedEvent
	if ev, err = events.NewPlayerAddedEvent(gameid, slackid, name, email); err == nil {
		p = NewPlayerFromEvent(ev)
	}
	return
}

// NewPlayerFromEvent instantiates a Player from a PlayerAdedEvent
func NewPlayerFromEvent(ev events.PlayerAddedEvent) (p Player) {
	p = Player{
		ID:				ev.ID,
		GameID:			ev.GameID,
		TimeCreated:	ev.TimeCreated,
		Name:			ev.Name,
		SlackID:		ev.SlackID,
		Email:			ev.Email,
		Status:			Alive,
		Kills:			0,
	}
	return
}

// GetID getter for ID field
func (p *Player) GetID() string {
	return p.ID
}

// SetTarget sets not just the target element but the kill word too. Bonus!
func (p *Player) SetTarget(targetID string, killWord string) {
	p.Target = targetID
	p.KillWord = killWord
}