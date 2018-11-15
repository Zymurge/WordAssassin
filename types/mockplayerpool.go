package types

import (
	"fmt"
)

// MockPlayerPool provides a test mock for PlayerPool dependencies
type MockPlayerPool struct {
	playersToReturn []*Player
}

// GetPlayerByID mock
func (mpp *MockPlayerPool) GetPlayerByID(searchid string) (*Player, error) {
	// TBD
	return nil, fmt.Errorf("Not implemented")
}

// GetAllPlayersInGame mock
func (mpp *MockPlayerPool) GetAllPlayersInGame(gameid string) ([]*Player, error) {
	if mpp.playersToReturn != nil {
		return mpp.playersToReturn, nil
	} else {
		return nil, fmt.Errorf("Mocked error")
	}
}
