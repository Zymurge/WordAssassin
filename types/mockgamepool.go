package types

import (
	"fmt"
)

// MockGamePool provides a test mock for GamePool dependencies
type MockGamePool struct {
	GamesToReturn   []Game
	AddGameError    string
	GetGameError    string
	StartGameError  string
}

// AddGame mock
func (mgp *MockGamePool) AddGame(game Game) error {
	if mgp.AddGameError != "" {
		return fmt.Errorf(mgp.AddGameError)
	}
	return nil
}

// GetGame mock
func (mgp *MockGamePool) GetGame(id string) (*Game, bool) {
	if mgp.GetGameError != "" {
		return nil, false
	}
	result := Game{ ID: id }
	return &result, true
}


// GetGamesList mock
func (mgp *MockGamePool) GetGamesList() []Game {
	return mgp.GamesToReturn
}

// StartGame mock
func (mgp *MockGamePool) StartGame(gameid string, slackid string) (err error) {
	if mgp.StartGameError != "" {
		return fmt.Errorf(mgp.StartGameError)
	}
	return nil
}