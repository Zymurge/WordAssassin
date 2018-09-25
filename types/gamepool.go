package types

import (
	"fmt"
	"sort"
)

// GamePool manages the collection of games in a running server
type GamePool struct {
	games map[string]*Game
}

// AddGame adds a game to this pool. Enforces uniqueness of the Game.ID within the pool
func (pool *GamePool) AddGame(game *Game) error {
	// create the players map as a singleton
	if pool.games == nil {
		pool.games = make(map[string]*Game, 10)
	}
	if game.GetID() == "" {
		return fmt.Errorf("missing ID for AddGame")
	}
	if _, exists := pool.games[game.GetID()]; exists {
		return fmt.Errorf("duplicate ID on add: %s", game.GetID())
	}
	pool.games[game.GetID()] = game

	// TODO: Persist to mongo
	return nil
}

// GetGame gets the game specified by the requested ID.
// Returns:
// -- the game object for that ID
// -- true for exists if it does, false if it don't
func (pool *GamePool) GetGame(id string) (result *Game, exists bool) {
	result, exists = pool.games[id]
	return
}

// GetGamesList gives a list of each game ID separated by a newline. The result are sorted chronologically by created time
func (pool *GamePool) GetGamesList() (result []*Game) {
	for _, v := range pool.games {
		result = append(result, v)
	}
	sort.SliceStable(result,
		func(i, j int) bool {
			return result[i].TimeCreated.Before(result[j].TimeCreated)
		})
	//	return result[i].GameCreator < result[j].GameCreator } )
	return
}

// ReconstitutePool rebuilds a new GamePool from an array of Games
func (pool *GamePool) ReconstitutePool(games []*Game) error {
	for _, v := range games {
		if err := pool.AddGame(v); err != nil {
			return err
		}
	}
	return nil
}
