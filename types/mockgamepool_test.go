package types

import (
	"testing"
	"github.com/stretchr/testify/require"

	events "wordassassin/types/events"
	"wordassassin/slack"
)

func TestMockAddGame(t *testing.T) {
	mgp := MockGamePool{}
	dummy := &Game{}
	require.NoError(t, mgp.AddGame(dummy), "Mock.AddGame should not error when AddGameError is unset")
	mgp.AddGameError = "mock error"
	actual := mgp.AddGame(dummy)
	require.Error(t, actual, "Mock.AddGame should error when AddError is set")
	require.Equal(t, actual.Error(), mgp.AddGameError, "Error message should passthrough unchanged")
}

func TestMockAddPlayerToGame(t *testing.T) {
	mgp := MockGamePool{}
	dummy := events.PlayerAddedEvent{}
	require.NoError(t, mgp.AddPlayerToGame("game", dummy), "Mock.AddPlayerToGame should not error when AddPlayerError is unset")
	mgp.AddPlayerError = "mock error"
	actual := mgp.AddPlayerToGame("game", dummy)
	require.Error(t, actual, "Mock.AddPlayerToGame should error when AddPlayerError is set")
	require.Equal(t, actual.Error(), mgp.AddPlayerError, "Error message should passthrough unchanged")
}
func TestCanAddPlayers(t *testing.T) {
	mgp := MockGamePool{}
	actual, err := mgp.CanAddPlayers("game")
	require.True(t, actual, "Mock.CanAddPlayers should return true when CanAddPlayers is unset")
	require.NoError(t, err, "Mock.CanAddPlayers should not error when CanAddPlayers is unset")
	mgp.CanAddError = "mock error"
	actual, err = mgp.CanAddPlayers("game")
	require.False(t, actual, "Mock.CanAddPlayers should return false when CanAddPlayers is set")
	require.Error(t, err, "Mock.CanAddPlayers should not error when CanAddPlayers is unset")
	require.Equal(t, err.Error(), mgp.CanAddError, "Error message should passthrough unchanged")
}

func TestMockGetGame(t *testing.T) {
	mgp := MockGamePool{}
	// Preset a game up front, to ensure that autogen game doesn't pick it up
	mgp.GamesToReturn = []*Game{ &Game{ ID: "preset", StartPlayers: 13, }, }
	mgp.GetGameError = ""
	actual, found := mgp.GetGame("ImATest")
	require.True(t, found, "Mock.GetGame should report found when GetGameError is unset")
	require.Equal(t, actual.GetID(), "ImATest", "Mock.GetGame should return a game ID as requested")
	actual, found = mgp.GetGame("preset")
	require.True(t, found, "Mock.GetGame should report found when GetGameError is unset")
	require.Equal(t, actual.GetID(), "preset", "Mock.GetGame should return the preset game ID as requested")
	require.Equal(t, actual.StartPlayers, 13, "Mock.GetGame should return the preset game, validated by custome field")
	mgp.GetGameError = "fail"
	_, found = mgp.GetGame("ImATest")
	require.False(t, found, "Mock.GetGame should report not found when GetGameError is set")
}

func TestMockGetGamesList(t *testing.T) {
	mgp := MockGamePool{}
	// Preset a game up front, to ensure that autogen game doesn't pick it up
	mgp.GamesToReturn = []*Game{ 
		&Game{ ID: "preset1", StartPlayers: 1, }, 
		&Game{ ID: "preset2", StartPlayers: 2, }, 
	}
	mgp.GetGameError = ""
	actual := mgp.GetGamesList()
	require.NotNil(t, actual, "Mock.GetGamesList should return a populated array")
	require.Equal(t, len(actual), 2, "Mock.GetGamesList should return the two preset games")
}

func TestMockStartGame(t *testing.T) {
	mgp := MockGamePool{}
	mySlackID, sErr := slack.New("UDuh")
	require.NoError(t, sErr, "Badness when creating the slack ID")
	require.NoError(t, mgp.StartGame("duh_game", mySlackID), "Mock.StartGame should not error when StartGameError is unset")
	mgp.StartGameError = "mock error"
	actual := mgp.StartGame("duh_game", mySlackID)
	require.Error(t, actual, "Mock.StartGame should error when StartError is set")
	require.Equal(t, actual.Error(), mgp.StartGameError, "Error message should passthrough unchanged")
}


