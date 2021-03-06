package types

import (
	"fmt"
	persistence "wordassassin/persistence"
	"wordassassin/slack"
)

// PlayerPoolAbstraction provides abstraction for testing PlayerPool dependencies
type PlayerPoolAbstraction interface {
	AddPlayer(player *Player) error
	GetPlayerByID(searchid string) (*Player, error)
	GetAllPlayersInGame(gameid string) ([]*Player, error)
}

const (
	// PlayersCollection const for the mongo collection to hold all player records
	PlayersCollection string = "players"
)

// PlayerPool manages the collection of players in a single game
// TODO: restructure as mongo backed collection
type PlayerPool struct {
	players map[string]*Player
	mongo   *persistence.MongoAbstraction
}

// AddPlayer adds a player to this pool. Enforces uniqueness of the Player.ID within the pool
// Errors:
//   Player.ID field is blank
//   Player.ID already exists in the pool
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

// GetPlayerByID fetches the player when the ID is known.
// Errors:
//   ID not found.
func (pool *PlayerPool) GetPlayerByID(searchid string) (*Player, error) {
	result, exists := pool.players[searchid]
	if !exists {
		return nil, fmt.Errorf("missing ID: %s", searchid)
	}
	return result, nil
}

// GetPlayer fetches the player specificed by the given game and slack ID combo. Errors on ID not found.
// Errors:
//   ID+slackid not found.
func (pool *PlayerPool) GetPlayer(gameid string, slackid slack.SlackID) (*Player, error) {
	searchid := gameid + "+" + slackid.ToString()
	return pool.GetPlayerByID(searchid)
}

// GetAllPlayersInGame fetches all of the players for a given gameid.
func (pool *PlayerPool) GetAllPlayersInGame(gameid string) (playersInGame []*Player, err error) {
	/*
		//TODO: restructure to use mongo fetch
		query := bson.NewDocument(
			bson.EC.String("gameid", gameid)
		)
		result := mongo.FetchAllFromCollection()
		players = PlayerPool.bytesToPlayers(result)
	*/
	for _, v := range pool.players {
		if v.GameID == gameid {
			playersInGame = append(playersInGame, v)
		}
	}
	return
}
