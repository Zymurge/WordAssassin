package types

import (
	"testing"
	"github.com/stretchr/testify/require"

	events "wordassassin/types/events"
)

func TestMockAddGame(t *testing.T) {
	mpp := MockGamePool{}
	dummy := &Game{}
	require.NoError(t, mpp.AddGame(dummy), "Mock.AddGame should not error when AddGameError is unset")
	mpp.AddGameError = "mock error"
	actual := mpp.AddGame(dummy)
	require.Error(t, actual, "Mock.AddGame should error when AddError is set")
	require.Equal(t, actual.Error(), mpp.AddGameError, "Error message should passthrough unchanged")
}

func TestMockAddPlayerToGame(t *testing.T) {
	mpp := MockGamePool{}
	dummy := events.PlayerAddedEvent{}
	require.NoError(t, mpp.AddPlayerToGame("game", dummy), "Mock.AddPlayerToGame should not error when AddPlayerError is unset")
	mpp.AddPlayerError = "mock error"
	actual := mpp.AddPlayerToGame("game", dummy)
	require.Error(t, actual, "Mock.AddPlayerToGame should error when AddPlayerError is set")
	require.Equal(t, actual.Error(), mpp.AddPlayerError, "Error message should passthrough unchanged")
}
func TestCanAddPlayers(t *testing.T) {
	mpp := MockGamePool{}
	actual, err := mpp.CanAddPlayers("game")
	require.True(t, actual, "Mock.CanAddPlayers should return true when CanAddPlayers is unset")
	require.NoError(t, err, "Mock.CanAddPlayers should not error when CanAddPlayers is unset")
	mpp.CanAddError = "mock error"
	actual, err = mpp.CanAddPlayers("game")
	require.False(t, actual, "Mock.CanAddPlayers should return false when CanAddPlayers is set")
	require.Error(t, err, "Mock.CanAddPlayers should not error when CanAddPlayers is unset")
	require.Equal(t, err.Error(), mpp.CanAddError, "Error message should passthrough unchanged")
}

func TestMockGetGame(t *testing.T) {
	mpp := MockGamePool{}
	// Preset a game up front, to ensure that autogen game doesn't pick it up
	mpp.GamesToReturn = []*Game{ &Game{ ID: "preset", StartPlayers: 13, }, }
	mpp.GetGameError = ""
	actual, found := mpp.GetGame("ImATest")
	require.True(t, found, "Mock.GetGame should report found when GetGameError is unset")
	require.Equal(t, actual.GetID(), "ImATest", "Mock.GetGame should return a game ID as requested")
	actual, found = mpp.GetGame("preset")
	require.True(t, found, "Mock.GetGame should report found when GetGameError is unset")
	require.Equal(t, actual.GetID(), "preset", "Mock.GetGame should return the preset game ID as requested")
	require.Equal(t, actual.StartPlayers, 13, "Mock.GetGame should return the preset game, validated by custome field")
	mpp.GetGameError = "fail"
	_, found = mpp.GetGame("ImATest")
	require.False(t, found, "Mock.GetGame should report not found when GetGameError is set")
}

func TestMockGetGamesList(t *testing.T) {
	mpp := MockGamePool{}
	// Preset a game up front, to ensure that autogen game doesn't pick it up
	mpp.GamesToReturn = []*Game{ 
		&Game{ ID: "preset1", StartPlayers: 1, }, 
		&Game{ ID: "preset2", StartPlayers: 2, }, 
	}
	mpp.GetGameError = ""
	actual := mpp.GetGamesList()
	require.NotNil(t, actual, "Mock.GetGamesList should return a populated array")
	require.Equal(t, len(actual), 2, "Mock.GetGamesList should return the two preset games")
}

func TestMockStartGame(t *testing.T) {
	mpp := MockGamePool{}
	require.NoError(t, mpp.StartGame("duh_game", "duh_creator"), "Mock.StartGame should not error when StartGameError is unset")
	mpp.StartGameError = "mock error"
	actual := mpp.StartGame("duh_game", "duh_creator")
	require.Error(t, actual, "Mock.StartGame should error when StartError is set")
	require.Equal(t, actual.Error(), mpp.StartGameError, "Error message should passthrough unchanged")
}


