package types

import (
	"fmt"
)

// MockPlayerPool provides a test mock for PlayerPool dependencies
type MockPlayerPool struct {
	// TBD
}

// GetPlayerByID mock
func (mpp *MockPlayerPool) GetPlayerByID(searchid string) (*Player, error) {
	// TBD
	return nil, fmt.Errorf("Not implemented")
}

// GetAllPlayersInGame mock
func (mpp *MockPlayerPool) GetAllPlayersInGame(gameid string) ([]*Player, error) {
	// TBD
	return nil, fmt.Errorf("Not implemented")
}
