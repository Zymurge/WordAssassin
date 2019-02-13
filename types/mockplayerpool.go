package types

import (
	"fmt"
)

// MockPlayerPool provides a test mock for PlayerPool dependencies
type MockPlayerPool struct {
	playersToReturn []*Player
	AddPlayerError  string
}

// AddPlayer mock
func (mpp MockPlayerPool) AddPlayer(player *Player) error {
	if mpp.AddPlayerError != "" {
		return fmt.Errorf(mpp.AddPlayerError)
	}
	return nil
}

// GetPlayerByID mock
func (mpp MockPlayerPool) GetPlayerByID(searchid string) (*Player, error) {
	// TBD
	return nil, fmt.Errorf("Not implemented")
}

// GetAllPlayersInGame mock
func (mpp MockPlayerPool) GetAllPlayersInGame(gameid string) ([]*Player, error) {
	if mpp.playersToReturn == nil {
		return nil, fmt.Errorf("Mocked error")
	}
	return mpp.playersToReturn, nil
}
