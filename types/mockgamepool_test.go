package types

import (
	"fmt"

	"wordassassin/types/events"
)

// MockGamePool provides a test mock for GamePool dependencies
type MockGamePool struct {
	GamesToReturn   []*Game
	AddGameError    string
	AddPlayerError  string
	CanAddError     string
	GetGameError    string
	StartGameError  string
}

// AddGame mock
func (mgp *MockGamePool) AddGame(game *Game) error {
	if mgp.AddGameError != "" {
		return fmt.Errorf(mgp.AddGameError)
	}
	return nil
}

// AddPlayerToGame mock
func (mgp *MockGamePool) AddPlayerToGame(gameid string, ev events.PlayerAddedEvent) error {
	if mgp.AddPlayerError != "" {
		return fmt.Errorf(mgp.AddPlayerError)
	}
	return nil
}


// CanAddPlayers mock
// return options true, nil or false, err from CanAddError
func (mgp *MockGamePool) CanAddPlayers(gameid string) (result bool, err error) {
	if mgp.CanAddError == "" {
		result = true
	} else {
		result = false
	}
	err = fmt.Errorf(mgp.CanAddError)
	return
}

// GetGame mock
func (mgp *MockGamePool) GetGame(id string) (*Game, bool) {
	if mgp.GetGameError != "" {
		return nil, false
	}
	// If preloaded with a matching ID game, then return it
	for _, g := range mgp.GamesToReturn {
		if id == g.GetID() { return g, true }
	}
	// else, provide a dummy for a positive experience
	result := Game{ ID: id }
	return &result, true
}

// GetGamesList mock
func (mgp *MockGamePool) GetGamesList() []*Game {
	return mgp.GamesToReturn
}

// StartGame mock
func (mgp *MockGamePool) StartGame(gameid string, slackid string) (err error) {
	if mgp.StartGameError != "" {
		return fmt.Errorf(mgp.StartGameError)
	}
	return nil
}