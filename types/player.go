package types

import (
	"fmt"
	//"encoding/json"
	//"fmt"
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

/*** End Player ***/

// PlayerPool manages the collection of players in a single game
type PlayerPool struct {
	players		map[string]*Player
}

// AddPlayer adds a player to this pool. Enforces uniqueness of the Player.ID within the pool
func (pool *PlayerPool) AddPlayer(player *Player) error {
	// create the players map as a singleton
	if pool.players == nil {
		pool.players = make(map[string]*Player, 10)
	}
	if player.GetID() == "" {
		return fmt.Errorf("missing ID for AddPlayer")
	}
	if _, exists := pool.players[player.GetID()]; exists {
		return fmt.Errorf("duplicate ID on add: %s", player.GetID())
	}
	pool.players[player.GetID()] = player
	return nil
}

// GetPlayerByID fetches the player when the ID is known. Errors on ID not found.
func (pool *PlayerPool) GetPlayerByID(searchid string) (*Player, error) {
	result, exists := pool.players[searchid]
	if !exists {
		return nil, fmt.Errorf("missing ID: %s", searchid)
	}
	return result, nil
}

// GetPlayer fetches the player specificed by the given game and slack ID combo. Errors on ID not found.
func (pool *PlayerPool) GetPlayer(gameid, slackid string) (*Player, error) {
	searchid := gameid + "+" + slackid
	return pool.GetPlayerByID(searchid)
}