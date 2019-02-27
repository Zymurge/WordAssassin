package types

import (
	"fmt"
)

// MockPlayerPool provides a test mock for PlayerPool dependencies
type MockPlayerPool struct {
	playersToReturn []*Player
	AddPlayerError  string
	GetPlayerError  string
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
	if mpp.GetPlayerError != "" {
		return nil, fmt.Errorf(mpp.GetPlayerError)
	}
	return mpp.playersToReturn, nil
}
